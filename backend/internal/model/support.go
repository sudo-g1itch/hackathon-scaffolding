package model

import (
	"time"

	"github.com/google/uuid"
)

// SupportMessage is one turn of the private conversation between a person in
// recovery and the caregiver they linked.
//
// The thread is identified by the (patient, caregiver) pair rather than by a
// separate thread row: a profile links to exactly one caregiver at a time, so
// the pair is already the thread. Keeping PatientID separate from SenderID is
// what lets both sides read the same conversation while every message still
// records who wrote it.
//
// This is deliberately NOT the AI coach conversation (CoachMessage). The coach
// is private to the user; this thread is shared with a human by explicit choice.
type SupportMessage struct {
	BaseModel
	PatientID   uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_id"`
	Patient     *User     `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	CaregiverID uuid.UUID `gorm:"type:uuid;not null;index" json:"caregiver_id"`
	Caregiver   *User     `gorm:"foreignKey:CaregiverID" json:"caregiver,omitempty"`

	// SenderID is always either PatientID or CaregiverID; SenderRole says which,
	// so the UI can align a bubble without comparing UUIDs.
	SenderID   uuid.UUID `gorm:"type:uuid;not null;index" json:"sender_id"`
	SenderRole string    `gorm:"type:varchar(50);not null" json:"sender_role"`

	// Kind separates an ordinary message from an emergency alert, so the
	// caregiver's thread can shout about one and not the other. An alert also
	// carries the EmergencyLog it came from.
	Kind          string     `gorm:"type:varchar(50);not null;default:'message';index" json:"kind"`
	EmergencyID   *uuid.UUID `gorm:"type:uuid;index" json:"emergency_id,omitempty"`

	Body string `gorm:"type:text;not null" json:"body"`

	// ReadAt is set when the *other* party opens the thread. Nil means unread,
	// which is what drives the unread badge on both dashboards.
	ReadAt *time.Time `json:"read_at,omitempty"`
}

// Support message kinds.
const (
	SupportMessageKindMessage   = "message"
	SupportMessageKindEmergency = "emergency"
)

// Recipient returns the user this message was addressed to.
func (m *SupportMessage) Recipient() uuid.UUID {
	if m.SenderID == m.PatientID {
		return m.CaregiverID
	}
	return m.PatientID
}
