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
	// supportThreadLimit bounds how much conversation is loaded at once.
	supportThreadLimit = 200

	// supportMessageMaxLen matches the handler's binding tag; both exist so the
	// rule holds whether the call arrives over HTTP or from another service.
	supportMessageMaxLen = 2000
)

// --- DTOs (mirrored by frontend/src/types/anchorOneTypes.ts) -----------------

// SupportMessageData is one message as the chat UI renders it.
type SupportMessageData struct {
	ID          uuid.UUID  `json:"id"`
	PatientID   uuid.UUID  `json:"patient_id"`
	CaregiverID uuid.UUID  `json:"caregiver_id"`
	SenderID    uuid.UUID  `json:"sender_id"`
	SenderRole  string     `json:"sender_role"`
	Kind        string     `json:"kind"`
	Body        string     `json:"body"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`

	// Emergency carries the alert's location link and voice-note transcript,
	// present only on emergency-kind messages. Inlined so a caregiver reading
	// the thread has everything without navigating away mid-crisis.
	Emergency *EmergencyAlertData `json:"emergency,omitempty"`
}

// SupportThread is the whole conversation plus who is on the other end of it.
//
// Linked is false when the user has not chosen a caregiver yet. That is an
// ordinary state, not an error: the UI answers it by pointing at the recovery
// plan screen where a caregiver is chosen, rather than showing an error.
type SupportThread struct {
	Linked        bool                 `json:"linked"`
	PatientID     uuid.UUID            `json:"patient_id"`
	PatientName   string               `json:"patient_name"`
	CaregiverID   *uuid.UUID           `json:"caregiver_id"`
	CaregiverName string               `json:"caregiver_name"`
	Messages      []SupportMessageData `json:"messages"`
	Unread        int64                `json:"unread"`
}

// SupportService owns the private conversation between a person in recovery and
// the caregiver they linked.
//
// Access is strictly the two of them. An administrator is not admitted, even
// though admins may read the caregiver overview: an overview is oversight, a
// conversation is not.
type SupportService interface {
	GetThread(ctx context.Context, actor Actor, patientID uuid.UUID) (*SupportThread, error)
	Send(ctx context.Context, actor Actor, patientID uuid.UUID, body string) (*SupportThread, error)
	MarkRead(ctx context.Context, actor Actor, patientID uuid.UUID) error

	// PostEmergencyAlert delivers a crisis alert into the thread. It skips the
	// ordinary length and authorisation checks on purpose: the caller has
	// already established the link, and an alert is composed from a stored
	// EmergencyLog rather than from user input arriving over the wire.
	PostEmergencyAlert(ctx context.Context, patientID, caregiverID uuid.UUID, entry *model.EmergencyLog) error

	// UnreadForUser counts everything waiting for this person, whichever side
	// of the conversation they are on. Drives the navigation badge.
	UnreadForUser(ctx context.Context, actor Actor) (int64, error)

	// UnreadByPatient powers the caregiver's list of the people they support.
	UnreadByPatient(ctx context.Context, caregiverID uuid.UUID) (map[uuid.UUID]int64, error)
}

type supportService struct {
	careAccess
	messages repository.SupportRepository
	users    repository.UserRepository
	log      *zap.Logger
	now      func() time.Time
}

func NewSupportService(
	messages repository.SupportRepository,
	profiles repository.RecoverAIRepository,
	users repository.UserRepository,
	log *zap.Logger,
) SupportService {
	return &supportService{
		careAccess: careAccess{profiles: profiles},
		messages:   messages,
		users:      users,
		log:        log,
		now:        time.Now,
	}
}

func (s *supportService) GetThread(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
) (*SupportThread, error) {
	_, caregiverID, err := s.resolveThread(ctx, actor, patientID)
	if err != nil {
		return nil, err
	}

	thread, err := s.buildThread(ctx, patientID, caregiverID)
	if err != nil {
		return nil, err
	}

	// Opening the thread is what "read" means, so the badge clears on view
	// rather than needing a separate tap.
	if thread.Linked {
		if err := s.messages.MarkRead(ctx, patientID, *caregiverID, actor.ID, s.now().UTC()); err != nil {
			return nil, apperr.Internal(err)
		}
		thread.Unread = 0
	}

	return thread, nil
}

func (s *supportService) Send(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
	body string,
) (*SupportThread, error) {
	relation, caregiverID, err := s.resolveThread(ctx, actor, patientID)
	if err != nil {
		return nil, err
	}
	if caregiverID == nil {
		return nil, apperr.Unprocessable(
			"There is no caregiver linked to this account yet, so there is nobody to message.")
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil, apperr.Validation(apperr.Fields{"body": {"is required"}})
	}
	if len(body) > supportMessageMaxLen {
		return nil, apperr.Validation(apperr.Fields{"body": {"must be at most 2000"}})
	}

	msg := &model.SupportMessage{
		PatientID:   patientID,
		CaregiverID: *caregiverID,
		SenderID:    actor.ID,
		SenderRole:  relation.AuthorRole(),
		Kind:        model.SupportMessageKindMessage,
		Body:        body,
	}
	if err := s.messages.Create(ctx, msg); err != nil {
		return nil, apperr.Internal(err)
	}

	s.log.Info("support message sent",
		zap.String("patient_id", patientID.String()),
		zap.String("sender_role", msg.SenderRole),
	)

	thread, err := s.buildThread(ctx, patientID, caregiverID)
	if err != nil {
		return nil, err
	}

	// The sender has by definition read everything before their own reply.
	if err := s.messages.MarkRead(ctx, patientID, *caregiverID, actor.ID, s.now().UTC()); err != nil {
		return nil, apperr.Internal(err)
	}
	thread.Unread = 0

	return thread, nil
}

func (s *supportService) PostEmergencyAlert(
	ctx context.Context,
	patientID, caregiverID uuid.UUID,
	entry *model.EmergencyLog,
) error {
	msg := &model.SupportMessage{
		PatientID:   patientID,
		CaregiverID: caregiverID,
		SenderID:    patientID,
		SenderRole:  model.AuthorRoleUser,
		Kind:        model.SupportMessageKindEmergency,
		EmergencyID: &entry.ID,
		Body:        entry.SentMessage,
	}

	if err := s.messages.Create(ctx, msg); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (s *supportService) MarkRead(ctx context.Context, actor Actor, patientID uuid.UUID) error {
	_, caregiverID, err := s.resolveThread(ctx, actor, patientID)
	if err != nil {
		return err
	}
	if caregiverID == nil {
		return nil
	}

	if err := s.messages.MarkRead(ctx, patientID, *caregiverID, actor.ID, s.now().UTC()); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (s *supportService) UnreadForUser(ctx context.Context, actor Actor) (int64, error) {
	// As a caregiver: everything the people they support have sent.
	byPatient, err := s.messages.CountUnreadByPatient(ctx, actor.ID)
	if err != nil {
		return 0, apperr.Internal(err)
	}

	var total int64
	for _, count := range byPatient {
		total += count
	}

	// As a patient: whatever their own caregiver has sent them.
	caregiverID, err := s.linkedCaregiver(ctx, actor.ID)
	if err != nil {
		return 0, err
	}
	if caregiverID != nil {
		mine, err := s.messages.CountUnreadFor(ctx, actor.ID, *caregiverID, actor.ID)
		if err != nil {
			return 0, apperr.Internal(err)
		}
		total += mine
	}

	return total, nil
}

func (s *supportService) UnreadByPatient(
	ctx context.Context,
	caregiverID uuid.UUID,
) (map[uuid.UUID]int64, error) {
	counts, err := s.messages.CountUnreadByPatient(ctx, caregiverID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return counts, nil
}

// --- internals ---------------------------------------------------------------

// resolveThread authorises the actor for this patient's conversation and
// returns the caregiver half of the pair.
//
// Only the two parties are admitted. RelationAdmin is intentionally absent from
// the allow-list.
func (s *supportService) resolveThread(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
) (CareRelation, *uuid.UUID, error) {
	relation, err := s.requireOneOf(ctx, actor, patientID, RelationSelf, RelationCaregiver)
	if err != nil {
		return "", nil, err
	}

	caregiverID, err := s.linkedCaregiver(ctx, patientID)
	if err != nil {
		return "", nil, err
	}

	return relation, caregiverID, nil
}

// buildThread assembles the conversation and the names on either side of it.
func (s *supportService) buildThread(
	ctx context.Context,
	patientID uuid.UUID,
	caregiverID *uuid.UUID,
) (*SupportThread, error) {
	thread := &SupportThread{
		PatientID:   patientID,
		CaregiverID: caregiverID,
		Messages:    []SupportMessageData{},
	}

	patient, err := s.users.GetByID(ctx, patientID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if patient != nil {
		thread.PatientName = patient.FullName()
	}

	if caregiverID == nil {
		return thread, nil
	}
	thread.Linked = true

	caregiver, err := s.users.GetByID(ctx, *caregiverID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if caregiver != nil {
		thread.CaregiverName = caregiver.FullName()
	}

	messages, err := s.messages.ListThread(ctx, patientID, *caregiverID, supportThreadLimit)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	thread.Messages = make([]SupportMessageData, 0, len(messages))
	for i := range messages {
		data := toSupportMessageData(&messages[i])

		// Only emergency messages carry a log, and a thread holds very few of
		// them, so loading them one by one is cheaper than a join that would
		// run for every ordinary message too.
		if messages[i].EmergencyID != nil {
			entry, err := s.profiles.GetEmergencyLog(ctx, *messages[i].EmergencyID)
			if err != nil {
				return nil, apperr.Internal(err)
			}
			if entry != nil {
				alert := ToEmergencyAlertData(entry)
				data.Emergency = &alert
			}
		}

		thread.Messages = append(thread.Messages, data)
	}

	return thread, nil
}

func toSupportMessageData(msg *model.SupportMessage) SupportMessageData {
	kind := msg.Kind
	if kind == "" {
		kind = model.SupportMessageKindMessage
	}

	return SupportMessageData{
		ID:          msg.ID,
		PatientID:   msg.PatientID,
		CaregiverID: msg.CaregiverID,
		SenderID:    msg.SenderID,
		SenderRole:  msg.SenderRole,
		Kind:        kind,
		Body:        msg.Body,
		ReadAt:      msg.ReadAt,
		CreatedAt:   msg.CreatedAt,
	}
}
