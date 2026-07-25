package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Risk levels are the vocabulary shared by the AI layer, the database and the
// frontend. Gemini is constrained to these exact values by a response schema,
// and NormalizeRisk is the safety net for anything that slips through.
const (
	RiskLow    = "LOW"
	RiskMedium = "MEDIUM"
	RiskHigh   = "HIGH"
)

// NormalizeRisk coerces free-form model output ("High", "high risk", "") into
// one of the three canonical levels, defaulting to MEDIUM when a level was
// clearly reported but is unrecognised, and LOW when nothing was reported.
func NormalizeRisk(raw string) string {
	switch s := strings.ToUpper(strings.TrimSpace(raw)); {
	case s == "":
		return RiskLow
	case strings.Contains(s, RiskHigh), strings.Contains(s, "SEVERE"), strings.Contains(s, "CRITICAL"):
		return RiskHigh
	case strings.Contains(s, RiskMedium), strings.Contains(s, "MODERATE"):
		return RiskMedium
	case strings.Contains(s, RiskLow), strings.Contains(s, "MINIMAL"):
		return RiskLow
	default:
		return RiskMedium
	}
}

// Coach message authors.
const (
	CoachRoleUser = "user"
	CoachRoleAI   = "ai"
)

// CravingMin and CravingMax bound the self-reported craving intensity scale.
const (
	CravingMin = 1
	CravingMax = 10
)

// ClampCraving keeps an AI-supplied craving score inside the 1-10 scale.
func ClampCraving(v int) int {
	switch {
	case v < CravingMin:
		return CravingMin
	case v > CravingMax:
		return CravingMax
	default:
		return v
	}
}

// RecoveryProfile is the personalisation record every AI prompt is grounded in.
type RecoveryProfile struct {
	BaseModel
	UserID           uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	User             *User      `json:"user,omitempty"`
	Goal             string     `gorm:"type:varchar(255)" json:"goal"`
	Substance        string     `gorm:"type:varchar(100)" json:"substance"`
	CaregiverID      *uuid.UUID `gorm:"type:uuid;index" json:"caregiver_id,omitempty"`
	Caregiver        *User      `gorm:"foreignKey:CaregiverID" json:"caregiver,omitempty"`
	CaregiverName    string     `gorm:"type:varchar(150)" json:"caregiver_name"`
	CaregiverPhone   string     `gorm:"type:varchar(50)" json:"caregiver_phone"`
	EmergencyContact string     `gorm:"type:varchar(150)" json:"emergency_contact"`

	// ShareCheckinDetails is the user's explicit consent for their linked
	// caregiver to read what a check-in actually SAID — its summary and
	// triggers. It defaults to false: a caregiver always sees risk, mood,
	// craving and streak, and only ever sees the narrative if the person in
	// recovery turns this on. Nothing widens the caregiver view except this
	// flag, and the raw transcript is never shared at all.
	ShareCheckinDetails bool `gorm:"not null;default:false" json:"share_checkin_details"`

	// Goals is the multi-goal recovery plan. Preloaded for AI prompts so the
	// coach knows what the person is actually working towards.
	Goals []RecoveryGoal `gorm:"foreignKey:UserID;references:UserID" json:"goals,omitempty"`
}

// Checkin is one analysed voice (or text) check-in.
type Checkin struct {
	BaseModel
	UserID             uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User               *User     `json:"user,omitempty"`
	Transcript         string    `gorm:"type:text" json:"transcript"`
	Summary            string    `gorm:"type:text" json:"summary"`
	Emotion            string    `gorm:"type:varchar(100)" json:"emotion"`
	Craving            int       `gorm:"type:int" json:"craving"`
	Risk               string    `gorm:"type:varchar(50);index" json:"risk"` // LOW, MEDIUM, HIGH
	Triggers           []string  `gorm:"type:jsonb;serializer:json" json:"triggers"`
	RecommendedActions []string  `gorm:"type:jsonb;serializer:json" json:"recommended_actions"`
	Source             string    `gorm:"type:varchar(20);not null;default:'voice'" json:"source"` // voice | text
}

// Check-in input channels.
const (
	CheckinSourceVoice = "voice"
	CheckinSourceText  = "text"
)

// CoachMessage is one turn of the recovery-coach conversation.
type CoachMessage struct {
	BaseModel
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User    *User     `json:"user,omitempty"`
	Role    string    `gorm:"type:varchar(50);not null" json:"role"` // "user" or "ai"
	Message string    `gorm:"type:text;not null" json:"message"`
}

// EmergencyLog records a triggered crisis intervention and the full plan that
// was generated for it, so the timeline can replay exactly what the user saw.
//
// The fields below GeneratedScript record what the user then chose to DO with
// that plan: which words they actually sent, whether they attached their
// location, and the voice note they recorded. Nothing there is filled in unless
// they pressed send — a triggered emergency that the user worked through alone
// is still a valid, complete record.
type EmergencyLog struct {
	BaseModel
	UserID             uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User               *User     `json:"user,omitempty"`
	Actions            []string  `gorm:"type:jsonb;serializer:json" json:"actions"`
	GeneratedScript    string    `gorm:"type:text" json:"generated_script"`
	GroundingExercise  string    `gorm:"type:text" json:"grounding_exercise"`
	EncouragingMessage string    `gorm:"type:text" json:"encouraging_message"`

	// SentMessage is the script as sent — a preset, the AI draft, or the user's
	// own words. Stored separately from GeneratedScript so the record shows
	// what was said, not what was suggested.
	SentMessage string     `gorm:"type:text" json:"sent_message"`
	SharedAt    *time.Time `json:"shared_at,omitempty"`
	CaregiverID *uuid.UUID `gorm:"type:uuid;index" json:"caregiver_id,omitempty"`

	// Location is attached only when the user turns the toggle on. Nil lat/lng
	// means they did not share it — the app never reads location otherwise.
	ShareLocation bool     `gorm:"not null;default:false" json:"share_location"`
	LocationLat   *float64 `json:"location_lat,omitempty"`
	LocationLng   *float64 `json:"location_lng,omitempty"`

	// AudioTranscript is the voice note the user recorded to send. Unlike a
	// check-in transcript — which is private by default — this exists only
	// because the user recorded it explicitly in order to send it.
	AudioTranscript string `gorm:"type:text" json:"audio_transcript"`

	// AcknowledgedAt is stamped when the caregiver confirms they have seen it,
	// which is the only honest way for the app to tell the user help is coming.
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// mapsURLTemplate builds a link any phone or desktop can open. Defined once so
// the alert message, the caregiver's list and the timeline agree.
const mapsURLTemplate = "https://www.google.com/maps/search/?api=1&query=%f,%f"

// LocationURL returns a Google Maps link for a shared location, or "" when the
// user did not share one.
func (e *EmergencyLog) LocationURL() string {
	if !e.ShareLocation || e.LocationLat == nil || e.LocationLng == nil {
		return ""
	}
	return fmt.Sprintf(mapsURLTemplate, *e.LocationLat, *e.LocationLng)
}

// EmergencyScript is one ready-to-send message.
type EmergencyScript struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Body  string `json:"body"`
}

// emergencyScriptTemplates are the presets offered alongside the AI draft.
//
// They exist because the AI is optional and because a person in crisis should
// not have to compose a sentence. %[1]s is the sender's first name, %[2]s the
// name they address their caregiver by.
var emergencyScriptTemplates = []struct {
	id, label, body string
}{
	{
		"reach-out", "I need someone",
		"%[2]s, it's %[1]s. I'm struggling right now and I don't want to be alone. Can you reply or come over?",
	},
	{
		"craving", "Strong craving",
		"%[2]s, it's %[1]s. I'm having a really strong craving and I'm scared I'll act on it. Please talk to me.",
	},
	{
		"unsafe", "I'm not safe",
		"%[2]s, it's %[1]s. I don't feel safe right now. Please come and find me as soon as you can.",
	},
	{
		"slipped", "I slipped",
		"%[2]s, it's %[1]s. I slipped and I didn't want to hide it from you. I want to get back on track.",
	},
	{
		"just-talk", "Just need to talk",
		"%[2]s, it's %[1]s. Nothing has happened yet, but it's a hard night. Can we talk for a few minutes?",
	},
}

// EmergencyScriptPresets renders the presets for one person, so what the user
// sees is already addressed and signed rather than a fill-in-the-blank form.
func EmergencyScriptPresets(senderName, caregiverName string) []EmergencyScript {
	sender := strings.TrimSpace(senderName)
	if sender == "" {
		sender = "your friend"
	}

	caregiver := strings.TrimSpace(caregiverName)
	if caregiver == "" {
		caregiver = "Hi"
	}

	scripts := make([]EmergencyScript, 0, len(emergencyScriptTemplates))
	for _, t := range emergencyScriptTemplates {
		scripts = append(scripts, EmergencyScript{
			ID:    t.id,
			Label: t.label,
			Body:  fmt.Sprintf(t.body, sender, caregiver),
		})
	}
	return scripts
}
