package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
)

const (
	// timelineLimit bounds how much history the timeline and trend chart read.
	timelineLimit = 50

	// streakLookbackDays bounds the streak scan.
	streakLookbackDays = 400

	day = 24 * time.Hour
)

// --- transport-agnostic DTOs -------------------------------------------------
//
// These types are the API contract for the recovery features, mirrored by
// frontend/src/types/anchorOneTypes.ts. They replace the untyped
// map[string]interface{} payloads, which let the two sides drift silently.

// Capabilities tells the client which optional integrations are actually
// configured, so the UI can disable the microphone or explain itself instead of
// failing at the moment the user taps it.
type Capabilities struct {
	AI    bool `json:"ai"`
	Voice bool `json:"voice"`
}

// ProfileData is the user-facing view of a recovery profile.
type ProfileData struct {
	Goal                string     `json:"goal"`
	Substance           string     `json:"substance"`
	CaregiverID         *uuid.UUID `json:"caregiver_id"`
	CaregiverName       string     `json:"caregiver_name"`
	CaregiverPhone      string     `json:"caregiver_phone"`
	EmergencyContact    string     `json:"emergency_contact"`
	LinkedCaregiverName string     `json:"linked_caregiver_name"`

	// ShareCheckinDetails is the user's consent for their linked caregiver to
	// read what a check-in said, not just how risky it scored. Off by default.
	ShareCheckinDetails bool `json:"share_checkin_details"`
}

// ProfileInput is the editable subset of a recovery profile.
type ProfileInput struct {
	Goal                string
	Substance           string
	CaregiverName       string
	CaregiverPhone      string
	EmergencyContact    string
	ShareCheckinDetails bool
}

// DashboardData backs the main dashboard.
type DashboardData struct {
	CurrentMood    string         `json:"current_mood"`
	RiskBadge      string         `json:"risk_badge"`
	CravingLevel   int            `json:"craving_level"`
	RecoveryStreak int            `json:"recovery_streak"`
	TotalCheckins  int64          `json:"total_checkins"`
	EmergencyCount int64          `json:"emergency_count"`
	LastCheckin    *model.Checkin `json:"last_checkin"`
	Profile        *ProfileData   `json:"profile"`
	Capabilities   Capabilities   `json:"capabilities"`

	// The recovery plan at a glance, and anything waiting from the caregiver,
	// so the home screen shows the whole picture in one request.
	Goals          GoalSummary `json:"goals"`
	UnreadMessages int64       `json:"unread_messages"`
}

// Timeline event kinds.
const (
	TimelineEventCheckin   = "checkin"
	TimelineEventEmergency = "emergency"
)

// TimelineEvent is one entry in the merged, reverse-chronological history.
type TimelineEvent struct {
	ID                 uuid.UUID `json:"id"`
	Type               string    `json:"type"`
	OccurredAt         time.Time `json:"occurred_at"`
	Risk               string    `json:"risk,omitempty"`
	Emotion            string    `json:"emotion,omitempty"`
	Craving            int       `json:"craving,omitempty"`
	Summary            string    `json:"summary,omitempty"`
	Triggers           []string  `json:"triggers,omitempty"`
	Actions            []string  `json:"actions,omitempty"`
	GeneratedScript    string    `json:"generated_script,omitempty"`
	GroundingExercise  string    `json:"grounding_exercise,omitempty"`
	EncouragingMessage string    `json:"encouraging_message,omitempty"`
	Source             string    `json:"source,omitempty"`
}

// CaregiverPatient and the caregiver's views of someone they support live in
// care_service.go, which composes goals and messaging alongside the risk
// signals. This file stays the user's own view of their own recovery.

// CaregiverOption is an assignable caregiver account.
type CaregiverOption struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// EmergencyResult pairs the persisted log with the plan the user is shown.
type EmergencyResult struct {
	Log  *model.EmergencyLog `json:"log"`
	Plan *EmergencyPlan      `json:"plan"`
}

// RecoverAIService orchestrates the recovery features: voice → reasoning →
// persistence → the views the UI reads.
type RecoverAIService interface {
	Capabilities() Capabilities

	// Check-ins
	ProcessVoiceCheckin(ctx context.Context, userID uuid.UUID, audio []byte, mimeType string) (*model.Checkin, error)
	ProcessTextCheckin(ctx context.Context, userID uuid.UUID, transcript string) (*model.Checkin, error)

	// Views
	GetDashboardData(ctx context.Context, userID uuid.UUID) (*DashboardData, error)
	GetTimeline(ctx context.Context, userID uuid.UUID) ([]TimelineEvent, error)

	// Crisis
	TriggerEmergency(ctx context.Context, userID uuid.UUID) (*EmergencyResult, error)

	// Coach & education
	SendCoachMessage(ctx context.Context, userID uuid.UUID, message string) ([]model.CoachMessage, error)
	GetCoachHistory(ctx context.Context, userID uuid.UUID) ([]model.CoachMessage, error)
	Educate(ctx context.Context, query string) (string, error)

	// Voice helpers
	Transcribe(ctx context.Context, audio []byte, mimeType string) (string, error)
	Speak(ctx context.Context, text string) ([]byte, error)

	// Profile & caregiver
	GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileData, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, in ProfileInput) (*ProfileData, error)
	ListAvailableCaregivers(ctx context.Context, userID uuid.UUID) ([]CaregiverOption, error)
	SetCaregiver(ctx context.Context, userID uuid.UUID, caregiverID *uuid.UUID) error
}

type recoverAIService struct {
	repo    repository.RecoverAIRepository
	users   repository.UserRepository
	goals   GoalService
	support SupportService
	ai      AIService
	voice   VoiceService
	log     *zap.Logger
	now     func() time.Time
}

func NewRecoverAIService(
	repo repository.RecoverAIRepository,
	users repository.UserRepository,
	goals GoalService,
	support SupportService,
	ai AIService,
	voice VoiceService,
	log *zap.Logger,
) RecoverAIService {
	return &recoverAIService{
		repo:    repo,
		users:   users,
		goals:   goals,
		support: support,
		ai:      ai,
		voice:   voice,
		log:     log,
		now:     time.Now,
	}
}

func (s *recoverAIService) Capabilities() Capabilities {
	return Capabilities{AI: s.ai.Available(), Voice: s.voice.Available()}
}

// --- check-ins ---------------------------------------------------------------

func (s *recoverAIService) ProcessVoiceCheckin(
	ctx context.Context,
	userID uuid.UUID,
	audio []byte,
	mimeType string,
) (*model.Checkin, error) {
	transcript, err := s.voice.TranscribeAudio(ctx, audio, mimeType)
	if err != nil {
		return nil, err
	}

	// Silence is a user outcome, not a server error: say so plainly instead of
	// sending an empty string to Gemini and storing a meaningless check-in.
	if strings.TrimSpace(transcript) == "" {
		return nil, apperr.Unprocessable(
			"We could not hear anything in that recording. Please check your microphone and try again.")
	}

	return s.analyze(ctx, userID, transcript, model.CheckinSourceVoice)
}

func (s *recoverAIService) ProcessTextCheckin(
	ctx context.Context,
	userID uuid.UUID,
	transcript string,
) (*model.Checkin, error) {
	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		return nil, apperr.Validation(apperr.Fields{"transcript": {"is required"}})
	}
	return s.analyze(ctx, userID, transcript, model.CheckinSourceText)
}

// analyze is the shared tail of both check-in paths: reason over the transcript,
// then persist the full assessment.
func (s *recoverAIService) analyze(
	ctx context.Context,
	userID uuid.UUID,
	transcript, source string,
) (*model.Checkin, error) {
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	analysis, err := s.ai.AnalyzeCheckin(ctx, transcript, profile)
	if err != nil {
		return nil, err
	}

	checkin := &model.Checkin{
		UserID:             userID,
		Transcript:         transcript,
		Summary:            analysis.Summary,
		Emotion:            analysis.Emotion,
		Craving:            analysis.Craving,
		Risk:               analysis.Risk,
		Triggers:           analysis.Triggers,
		RecommendedActions: analysis.RecommendedActions,
		Source:             source,
	}

	if err := s.repo.CreateCheckin(ctx, checkin); err != nil {
		return nil, apperr.Internal(err)
	}

	s.log.Info("check-in analysed",
		zap.String("user_id", userID.String()),
		zap.String("risk", checkin.Risk),
		zap.Int("craving", checkin.Craving),
		zap.String("source", source),
	)

	return checkin, nil
}

// --- views -------------------------------------------------------------------

func (s *recoverAIService) GetDashboardData(ctx context.Context, userID uuid.UUID) (*DashboardData, error) {
	lastCheckin, err := s.repo.GetLastCheckin(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	totalCheckins, err := s.repo.CountCheckins(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	emergencyCount, err := s.repo.CountEmergencyLogs(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	streak, err := s.recoveryStreak(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	goals, err := s.goals.Summary(ctx, userID)
	if err != nil {
		return nil, err
	}

	unread, err := s.support.UnreadForUser(ctx, Actor{ID: userID, Role: model.RoleUser})
	if err != nil {
		return nil, err
	}

	data := &DashboardData{
		CurrentMood:    "Unknown",
		RiskBadge:      model.RiskLow,
		RecoveryStreak: streak,
		TotalCheckins:  totalCheckins,
		EmergencyCount: emergencyCount,
		LastCheckin:    lastCheckin,
		Profile:        toProfileData(profile),
		Capabilities:   s.Capabilities(),
		Goals:          *goals,
		UnreadMessages: unread,
	}

	if lastCheckin != nil {
		if lastCheckin.Emotion != "" {
			data.CurrentMood = lastCheckin.Emotion
		}
		data.RiskBadge = model.NormalizeRisk(lastCheckin.Risk)
		data.CravingLevel = lastCheckin.Craving
	}

	return data, nil
}

// GetTimeline merges check-ins and emergencies into one reverse-chronological
// history — the "chronological recovery history" the PRD describes, rather than
// two disconnected lists.
func (s *recoverAIService) GetTimeline(ctx context.Context, userID uuid.UUID) ([]TimelineEvent, error) {
	checkins, err := s.repo.ListCheckins(ctx, userID, timelineLimit)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	logs, err := s.repo.ListEmergencyLogs(ctx, userID, timelineLimit)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	events := make([]TimelineEvent, 0, len(checkins)+len(logs))

	for _, c := range checkins {
		events = append(events, TimelineEvent{
			ID:         c.ID,
			Type:       TimelineEventCheckin,
			OccurredAt: c.CreatedAt,
			Risk:       c.Risk,
			Emotion:    c.Emotion,
			Craving:    c.Craving,
			Summary:    c.Summary,
			Triggers:   c.Triggers,
			Actions:    c.RecommendedActions,
			Source:     c.Source,
		})
	}

	for _, l := range logs {
		events = append(events, TimelineEvent{
			ID:                 l.ID,
			Type:               TimelineEventEmergency,
			OccurredAt:         l.CreatedAt,
			Risk:               model.RiskHigh,
			Actions:            l.Actions,
			GeneratedScript:    l.GeneratedScript,
			GroundingExercise:  l.GroundingExercise,
			EncouragingMessage: l.EncouragingMessage,
		})
	}

	sortEventsDesc(events)
	return events, nil
}

func sortEventsDesc(events []TimelineEvent) {
	// Insertion sort over two already-sorted runs; the slice is bounded by
	// 2*timelineLimit, so this stays trivial.
	for i := 1; i < len(events); i++ {
		for j := i; j > 0 && events[j].OccurredAt.After(events[j-1].OccurredAt); j-- {
			events[j], events[j-1] = events[j-1], events[j]
		}
	}
}

// recoveryStreak counts consecutive days, ending today or yesterday, on which
// the user checked in. A gap breaks the streak; a missing today does not, so the
// number does not collapse to zero before the user has checked in this morning.
func (s *recoverAIService) recoveryStreak(ctx context.Context, userID uuid.UUID) (int, error) {
	days, err := s.repo.CheckinDays(ctx, userID, streakLookbackDays)
	if err != nil {
		return 0, apperr.Internal(err)
	}
	return calculateStreak(days, s.now()), nil
}

func calculateStreak(days []time.Time, now time.Time) int {
	if len(days) == 0 {
		return 0
	}

	today := now.UTC().Truncate(day)
	previous := days[0].UTC().Truncate(day)

	// More than one day since the most recent check-in: the streak is over.
	if today.Sub(previous) > day {
		return 0
	}

	streak := 1
	for _, raw := range days[1:] {
		current := raw.UTC().Truncate(day)

		switch previous.Sub(current) {
		case 0: // same day — nothing to count
			continue
		case day: // consecutive
			streak++
			previous = current
		default: // gap
			return streak
		}
	}
	return streak
}

// --- crisis ------------------------------------------------------------------

func (s *recoverAIService) TriggerEmergency(ctx context.Context, userID uuid.UUID) (*EmergencyResult, error) {
	// A profile is created on demand: emergency is the one feature that must
	// never fail because the user has not filled in a form yet.
	profile, err := s.ensureProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	lastCheckin, err := s.repo.GetLastCheckin(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Fall back to the linked caregiver's account name so the generated SMS is
	// still addressed to a person when the free-text field was left blank. This
	// is prompt context only and is not written back to the profile.
	prompt := *profile
	if strings.TrimSpace(prompt.CaregiverName) == "" && prompt.Caregiver != nil {
		prompt.CaregiverName = prompt.Caregiver.FullName()
	}

	plan, err := s.ai.GenerateEmergencyPlan(ctx, prompt, lastCheckin)
	if err != nil {
		return nil, err
	}

	entry := &model.EmergencyLog{
		UserID:             userID,
		Actions:            plan.ImmediateActions,
		GeneratedScript:    plan.EmergencySMS,
		GroundingExercise:  plan.GroundingExercise,
		EncouragingMessage: plan.EncouragingMessage,
	}
	if err := s.repo.CreateEmergencyLog(ctx, entry); err != nil {
		return nil, apperr.Internal(err)
	}

	s.log.Warn("emergency triggered", zap.String("user_id", userID.String()))

	return &EmergencyResult{Log: entry, Plan: plan}, nil
}

// --- coach & education -------------------------------------------------------

func (s *recoverAIService) SendCoachMessage(
	ctx context.Context,
	userID uuid.UUID,
	message string,
) ([]model.CoachMessage, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, apperr.Validation(apperr.Fields{"message": {"is required"}})
	}

	// Read the prior conversation BEFORE storing this turn, otherwise the new
	// message arrives at the model twice — once inside the history and once as
	// the current turn.
	history, err := s.repo.GetCoachHistory(ctx, userID, coachHistoryTurns)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	// Ask the model first: if the AI is unavailable, nothing is persisted and
	// the user can simply retry, rather than leaving a dangling user turn.
	reply, err := s.ai.ChatCoach(ctx, history, message, profile)
	if err != nil {
		return nil, err
	}

	userMsg := &model.CoachMessage{UserID: userID, Role: model.CoachRoleUser, Message: message}
	if err := s.repo.CreateCoachMessage(ctx, userMsg); err != nil {
		return nil, apperr.Internal(err)
	}

	aiMsg := &model.CoachMessage{UserID: userID, Role: model.CoachRoleAI, Message: reply}
	if err := s.repo.CreateCoachMessage(ctx, aiMsg); err != nil {
		return nil, apperr.Internal(err)
	}

	return append(history, *userMsg, *aiMsg), nil
}

func (s *recoverAIService) GetCoachHistory(ctx context.Context, userID uuid.UUID) ([]model.CoachMessage, error) {
	history, err := s.repo.GetCoachHistory(ctx, userID, coachHistoryTurns*4)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return history, nil
}

func (s *recoverAIService) Educate(ctx context.Context, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", apperr.Validation(apperr.Fields{"query": {"is required"}})
	}
	return s.ai.Educate(ctx, query)
}

// --- voice -------------------------------------------------------------------

func (s *recoverAIService) Transcribe(ctx context.Context, audio []byte, mimeType string) (string, error) {
	return s.voice.TranscribeAudio(ctx, audio, mimeType)
}

func (s *recoverAIService) Speak(ctx context.Context, text string) ([]byte, error) {
	return s.voice.SynthesizeSpeech(ctx, text)
}

// --- profile & caregiver -----------------------------------------------------

func (s *recoverAIService) GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileData, error) {
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if profile == nil {
		// No row yet is not an error — report an empty profile.
		return &ProfileData{}, nil
	}
	return toProfileData(profile), nil
}

func (s *recoverAIService) UpdateProfile(
	ctx context.Context,
	userID uuid.UUID,
	in ProfileInput,
) (*ProfileData, error) {
	profile := &model.RecoveryProfile{
		UserID:              userID,
		Goal:                strings.TrimSpace(in.Goal),
		Substance:           strings.TrimSpace(in.Substance),
		CaregiverName:       strings.TrimSpace(in.CaregiverName),
		CaregiverPhone:      strings.TrimSpace(in.CaregiverPhone),
		EmergencyContact:    strings.TrimSpace(in.EmergencyContact),
		ShareCheckinDetails: in.ShareCheckinDetails,
	}

	if err := s.repo.UpsertProfile(ctx, profile); err != nil {
		return nil, apperr.Internal(err)
	}

	return s.GetProfile(ctx, userID)
}

// ensureProfile returns the user's profile, creating an empty one if needed.
func (s *recoverAIService) ensureProfile(
	ctx context.Context,
	userID uuid.UUID,
) (*model.RecoveryProfile, error) {
	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if profile != nil {
		return profile, nil
	}

	created := &model.RecoveryProfile{UserID: userID}
	if err := s.repo.UpsertProfile(ctx, created); err != nil {
		return nil, apperr.Internal(err)
	}

	// Re-read so the caller sees the row as stored (including any preloads).
	profile, err = s.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if profile == nil {
		return created, nil
	}
	return profile, nil
}

func (s *recoverAIService) ListAvailableCaregivers(
	ctx context.Context,
	userID uuid.UUID,
) ([]CaregiverOption, error) {
	users, err := s.repo.ListAvailableCaregivers(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	options := make([]CaregiverOption, 0, len(users))
	for i := range users {
		options = append(options, CaregiverOption{
			ID:   users[i].ID,
			Name: users[i].FullName(),
		})
	}
	return options, nil
}

// SetCaregiver links a caregiver to the user's profile. A nil caregiverID
// unlinks. The target is validated so a user cannot point their profile at an
// arbitrary account and expose their risk data to it.
func (s *recoverAIService) SetCaregiver(
	ctx context.Context,
	userID uuid.UUID,
	caregiverID *uuid.UUID,
) error {
	if caregiverID != nil {
		if *caregiverID == userID {
			return apperr.Unprocessable("You cannot assign yourself as your own caregiver.")
		}

		caregiver, err := s.users.GetByID(ctx, *caregiverID)
		if err != nil {
			return apperr.Internal(err)
		}
		if caregiver == nil {
			return apperr.NotFound("caregiver")
		}
		if !caregiver.IsActive {
			return apperr.Unprocessable("That caregiver account is inactive.")
		}
		if caregiver.Role != model.RoleCaregiver {
			return apperr.Unprocessable("That account is not a caregiver.")
		}
	}

	// The profile row must exist before it can be pointed at a caregiver.
	if _, err := s.ensureProfile(ctx, userID); err != nil {
		return err
	}

	if err := s.repo.SetCaregiver(ctx, userID, caregiverID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func toProfileData(profile *model.RecoveryProfile) *ProfileData {
	if profile == nil {
		return &ProfileData{}
	}

	data := &ProfileData{
		Goal:                profile.Goal,
		Substance:           profile.Substance,
		CaregiverID:         profile.CaregiverID,
		CaregiverName:       profile.CaregiverName,
		CaregiverPhone:      profile.CaregiverPhone,
		EmergencyContact:    profile.EmergencyContact,
		ShareCheckinDetails: profile.ShareCheckinDetails,
	}
	if profile.Caregiver != nil {
		data.LinkedCaregiverName = profile.Caregiver.FullName()
	}
	return data
}
