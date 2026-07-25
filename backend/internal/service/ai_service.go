package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
)

const (
	geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

	// coachHistoryTurns is how many past turns are replayed to the model.
	coachHistoryTurns = 10

	// geminiAttempts is the total number of tries per call. Gemini answers a
	// user in crisis, so one retry on a transient upstream failure is worth
	// the extra latency.
	geminiAttempts = 2
	geminiBackoff  = 750 * time.Millisecond
)

// CheckinAnalysis is the structured risk assessment for one check-in.
type CheckinAnalysis struct {
	Summary            string   `json:"summary"`
	Emotion            string   `json:"emotion"`
	Craving            int      `json:"craving"`
	Risk               string   `json:"risk"`
	Triggers           []string `json:"triggers"`
	RecommendedActions []string `json:"recommended_actions"`
}

// EmergencyPlan is the crisis intervention package.
type EmergencyPlan struct {
	ImmediateActions   []string `json:"immediate_actions"`
	EmergencySMS       string   `json:"emergency_sms"`
	GroundingExercise  string   `json:"grounding_exercise"`
	EncouragingMessage string   `json:"encouraging_message"`
}

// AIService is the reasoning layer: every Gemini interaction lives behind it.
type AIService interface {
	// Available reports whether a Gemini key is configured.
	Available() bool
	AnalyzeCheckin(ctx context.Context, transcript string, profile *model.RecoveryProfile) (*CheckinAnalysis, error)
	GenerateEmergencyPlan(ctx context.Context, profile model.RecoveryProfile, lastCheckin *model.Checkin) (*EmergencyPlan, error)
	ChatCoach(ctx context.Context, history []model.CoachMessage, newMsg string, profile *model.RecoveryProfile) (string, error)
	Educate(ctx context.Context, query string) (string, error)
}

type aiService struct {
	cfg    config.AI
	log    *zap.Logger
	client *http.Client
}

// NewAIService builds the Gemini-backed reasoning service. A missing API key is
// not an error: the service reports itself unavailable and every call returns
// apperr.Unavailable, so the rest of the app keeps working.
func NewAIService(cfg config.AI, log *zap.Logger) AIService {
	return &aiService{
		cfg:    cfg,
		log:    log,
		client: &http.Client{Timeout: cfg.RequestTimeout},
	}
}

func (s *aiService) Available() bool { return s.cfg.GeminiEnabled() }

// --- prompts (verbatim from the PRD) ---

const coachSystemPrompt = `You are a compassionate recovery coach supporting someone with a substance use disorder.
Never diagnose. Never prescribe medication. Never shame the user.
Use motivational interviewing and CBT techniques.
Keep responses below 120 words.
Always end with one practical action the user can take right now.
If the user describes an immediate medical emergency or intent to harm themselves, tell them to contact emergency services or a crisis line straight away.`

const riskSystemPrompt = `You are a clinical risk-detection assistant for a substance use recovery platform.
Analyze the user's check-in and score their relapse risk.
craving is an integer from 1 (no craving) to 10 (overwhelming craving).
risk is LOW, MEDIUM or HIGH.
triggers lists the concrete situations, people or feelings driving the risk.
recommended_actions lists 3 to 5 short, concrete, immediately doable steps.
Be honest: do not understate a high risk.`

const emergencySystemPrompt = `You are a crisis intervention planner for a substance use recovery platform.
The user has just pressed an emergency button and needs help in the next few minutes.
immediate_actions: 5 very short, direct, physically doable steps.
emergency_sms: a first-person message the user can send to their caregiver asking for help. Warm, direct, no shame. Address the caregiver by name when one is given.
grounding_exercise: one brief grounding technique with concrete instructions.
encouraging_message: two or three compassionate sentences reminding them of their progress.
Never prescribe medication. Never shame the user.`

const educationSystemPrompt = `You are a recovery education assistant.
Explain the topic using plain English a distressed person can follow.
Maximum 200 words.
Use markdown bullet points.
End with a "Three things you can do" section containing exactly three actionable tips.
Never diagnose and never prescribe medication.`

// --- response schemas: these are what force correctly typed JSON ---
//
// Without an explicit schema Gemini answers craving as prose
// ("Moderate-to-High") and risk in mixed case ("High"), which fails to
// unmarshal into the Go structs and breaks risk comparisons downstream.

var checkinAnalysisSchema = map[string]any{
	"type": "OBJECT",
	"properties": map[string]any{
		"summary": map[string]any{"type": "STRING"},
		"emotion": map[string]any{"type": "STRING"},
		"craving": map[string]any{"type": "INTEGER"},
		"risk": map[string]any{
			"type": "STRING",
			"enum": []string{model.RiskLow, model.RiskMedium, model.RiskHigh},
		},
		"triggers":            map[string]any{"type": "ARRAY", "items": map[string]any{"type": "STRING"}},
		"recommended_actions": map[string]any{"type": "ARRAY", "items": map[string]any{"type": "STRING"}},
	},
	"required":         []string{"summary", "emotion", "craving", "risk", "triggers", "recommended_actions"},
	"propertyOrdering": []string{"summary", "emotion", "craving", "risk", "triggers", "recommended_actions"},
}

var emergencyPlanSchema = map[string]any{
	"type": "OBJECT",
	"properties": map[string]any{
		"immediate_actions":   map[string]any{"type": "ARRAY", "items": map[string]any{"type": "STRING"}},
		"emergency_sms":       map[string]any{"type": "STRING"},
		"grounding_exercise":  map[string]any{"type": "STRING"},
		"encouraging_message": map[string]any{"type": "STRING"},
	},
	"required": []string{"immediate_actions", "emergency_sms", "grounding_exercise", "encouraging_message"},
	"propertyOrdering": []string{
		"immediate_actions", "emergency_sms", "grounding_exercise", "encouraging_message",
	},
}

func (s *aiService) AnalyzeCheckin(
	ctx context.Context,
	transcript string,
	profile *model.RecoveryProfile,
) (*CheckinAnalysis, error) {
	prompt := fmt.Sprintf("%sCheck-in transcript:\n%q", profileContext(profile), transcript)

	var analysis CheckinAnalysis
	if err := s.generateJSON(ctx, riskSystemPrompt, prompt, checkinAnalysisSchema, &analysis); err != nil {
		return nil, err
	}

	// Defence in depth: the schema constrains these, but never trust a model
	// with values the UI colour-codes and the caregiver dashboard alerts on.
	analysis.Risk = model.NormalizeRisk(analysis.Risk)
	analysis.Craving = model.ClampCraving(analysis.Craving)

	return &analysis, nil
}

func (s *aiService) GenerateEmergencyPlan(
	ctx context.Context,
	profile model.RecoveryProfile,
	lastCheckin *model.Checkin,
) (*EmergencyPlan, error) {
	var b strings.Builder
	b.WriteString(profileContext(&profile))

	if name := strings.TrimSpace(profile.CaregiverName); name != "" {
		b.WriteString("Caregiver to address in the SMS: " + name + "\n")
	} else {
		b.WriteString("No caregiver name is known — address the SMS neutrally (e.g. \"Hi\").\n")
	}
	if lastCheckin != nil {
		b.WriteString(fmt.Sprintf(
			"Most recent check-in — risk %s, emotion %q, craving %d/10, triggers: %s.\n",
			lastCheckin.Risk, lastCheckin.Emotion, lastCheckin.Craving,
			joinOr(lastCheckin.Triggers, "none recorded"),
		))
	}
	b.WriteString("\nGenerate the emergency plan now.")

	var plan EmergencyPlan
	if err := s.generateJSON(ctx, emergencySystemPrompt, b.String(), emergencyPlanSchema, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *aiService) ChatCoach(
	ctx context.Context,
	history []model.CoachMessage,
	newMsg string,
	profile *model.RecoveryProfile,
) (string, error) {
	// Replay the conversation as real multi-turn content rather than flattening
	// it into one string, so the model tracks who said what.
	contents := make([]geminiContent, 0, len(history)+1)
	for _, msg := range history {
		role := "model"
		if msg.Role == model.CoachRoleUser {
			role = "user"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Message}},
		})
	}
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: newMsg}},
	})

	system := coachSystemPrompt
	if ctxText := profileContext(profile); ctxText != "" {
		system += "\n\nContext about this user:\n" + ctxText
	}

	return s.generate(ctx, geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: system}}},
		Contents:          contents,
		GenerationConfig:  s.generationConfig(nil),
	})
}

func (s *aiService) Educate(ctx context.Context, query string) (string, error) {
	return s.generate(ctx, geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: educationSystemPrompt}}},
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: "Explain: " + query}},
		}},
		GenerationConfig: s.generationConfig(nil),
	})
}

// profileContext renders the personalisation block shared by every prompt.
func profileContext(profile *model.RecoveryProfile) string {
	if profile == nil {
		return ""
	}

	var b strings.Builder
	if goal := strings.TrimSpace(profile.Goal); goal != "" {
		b.WriteString("Recovery goal: " + goal + "\n")
	}
	if substance := strings.TrimSpace(profile.Substance); substance != "" {
		b.WriteString("Substance of concern: " + substance + "\n")
	}

	// The open goals are what the person is actually working on right now.
	// Without them the coach can only speak in generalities, and its
	// suggestions drift away from the plan the user built.
	if plan := goalContext(profile.Goals); plan != "" {
		b.WriteString(plan)
	}
	return b.String()
}

// goalContext renders the active recovery plan as prompt context: what each
// goal is and how far along it is.
func goalContext(goals []model.RecoveryGoal) string {
	open := make([]string, 0, len(goals))
	for i := range goals {
		if !goals[i].IsOpen() {
			continue
		}
		open = append(open, fmt.Sprintf("- %s (%d/%d %s, %d%% done)",
			goals[i].Title,
			goals[i].CurrentValue,
			goals[i].TargetValue,
			goals[i].Unit,
			goals[i].ProgressPercent(),
		))
	}

	if len(open) == 0 {
		return ""
	}
	return "Active recovery goals:\n" + strings.Join(open, "\n") + "\n"
}

func joinOr(items []string, fallback string) string {
	if len(items) == 0 {
		return fallback
	}
	return strings.Join(items, ", ")
}

// --- Gemini wire format ---

type geminiPart struct {
	Text    string `json:"text,omitempty"`
	Thought bool   `json:"thought,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiThinkingConfig struct {
	ThinkingLevel string `json:"thinkingLevel,omitempty"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string                `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]any        `json:"responseSchema,omitempty"`
	ThinkingConfig   *geminiThinkingConfig `json:"thinkingConfig,omitempty"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	Contents          []geminiContent         `json:"contents"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content      geminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback *struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// generationConfig builds the per-call config. A JSON schema switches the model
// into structured-output mode; thinking is dialled down for latency on the
// Gemini 3 family (older families reject the field, so it is only sent there).
func (s *aiService) generationConfig(schema map[string]any) *geminiGenerationConfig {
	cfg := &geminiGenerationConfig{}
	if schema != nil {
		cfg.ResponseMimeType = "application/json"
		cfg.ResponseSchema = schema
	}
	if strings.HasPrefix(s.cfg.GeminiModel, "gemini-3") {
		cfg.ThinkingConfig = &geminiThinkingConfig{ThinkingLevel: "low"}
	}
	return cfg
}

// generateJSON runs a structured-output call and unmarshals it into out.
func (s *aiService) generateJSON(
	ctx context.Context,
	system, prompt string,
	schema map[string]any,
	out any,
) error {
	raw, err := s.generate(ctx, geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: system}}},
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: prompt}},
		}},
		GenerationConfig: s.generationConfig(schema),
	})
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(raw), out); err != nil {
		s.log.Error("gemini returned unparseable JSON",
			zap.Error(err), zap.String("raw", truncate(raw, 500)))
		return apperr.Unavailable("The AI service returned an unreadable response. Please try again.").Wrap(err)
	}
	return nil
}

// generate performs the HTTP call, retrying once on a transient failure.
func (s *aiService) generate(ctx context.Context, payload geminiRequest) (string, error) {
	if !s.Available() {
		return "", apperr.Unavailable(
			"AI features are not configured on this server. Set AI_GEMINI_API_KEY to enable them.")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("service: encoding gemini request: %w", err))
	}

	var lastErr error
	for attempt := 1; attempt <= geminiAttempts; attempt++ {
		text, retryable, err := s.doGenerate(ctx, body)
		if err == nil {
			return text, nil
		}
		lastErr = err

		if !retryable || attempt == geminiAttempts {
			break
		}

		s.log.Warn("gemini call failed, retrying",
			zap.Int("attempt", attempt), zap.Error(err))

		select {
		case <-ctx.Done():
			return "", apperr.Unavailable("The AI request was cancelled.").Wrap(ctx.Err())
		case <-time.After(geminiBackoff):
		}
	}

	s.log.Error("gemini call failed", zap.Error(lastErr), zap.String("model", s.cfg.GeminiModel))

	return "", apperr.Unavailable(
		"The AI service is temporarily unavailable. Please try again in a moment.").Wrap(lastErr)
}

// doGenerate makes one attempt. The bool reports whether a retry may help.
func (s *aiService) doGenerate(ctx context.Context, body []byte) (string, bool, error) {
	url := fmt.Sprintf("%s/%s:generateContent", geminiBaseURL, s.cfg.GeminiModel)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Content-Type", "application/json")
	// Header rather than a query parameter, so the key never lands in a URL,
	// proxy log or error string.
	req.Header.Set("x-goog-api-key", s.cfg.GeminiAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", true, fmt.Errorf("service: calling gemini: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", true, fmt.Errorf("service: reading gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return "", retryable, fmt.Errorf("service: gemini status %d: %s",
			resp.StatusCode, truncate(string(respBody), 500))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", false, fmt.Errorf("service: decoding gemini response: %w", err)
	}
	if parsed.Error != nil {
		return "", parsed.Error.Code >= 500, fmt.Errorf("service: gemini error: %s", parsed.Error.Message)
	}
	if parsed.PromptFeedback != nil && parsed.PromptFeedback.BlockReason != "" {
		return "", false, fmt.Errorf("service: gemini blocked the prompt: %s", parsed.PromptFeedback.BlockReason)
	}
	if len(parsed.Candidates) == 0 {
		return "", true, fmt.Errorf("service: gemini returned no candidates")
	}

	text := candidateText(parsed.Candidates[0].Content)
	if text == "" {
		return "", true, fmt.Errorf("service: gemini returned an empty answer (finish reason %q)",
			parsed.Candidates[0].FinishReason)
	}
	return text, false, nil
}

// candidateText joins every answer part, skipping the model's internal
// "thought" parts — on thinking models the answer is not always parts[0].
func candidateText(content geminiContent) string {
	var b strings.Builder
	for _, part := range content.Parts {
		if part.Thought || part.Text == "" {
			continue
		}
		b.WriteString(part.Text)
	}
	return strings.TrimSpace(b.String())
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
