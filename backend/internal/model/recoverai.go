package model

import (
	"strings"

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
type EmergencyLog struct {
	BaseModel
	UserID             uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User               *User     `json:"user,omitempty"`
	Actions            []string  `gorm:"type:jsonb;serializer:json" json:"actions"`
	GeneratedScript    string    `gorm:"type:text" json:"generated_script"`
	GroundingExercise  string    `gorm:"type:text" json:"grounding_exercise"`
	EncouragingMessage string    `gorm:"type:text" json:"encouraging_message"`
}
