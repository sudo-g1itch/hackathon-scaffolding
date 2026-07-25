package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
)

// SupportRepository owns the patient <-> caregiver conversation.
//
// Every method is scoped by the (patient, caregiver) pair rather than by a
// single user id: that pair IS the thread, and requiring both halves means a
// query can never accidentally return a message from a different pairing.
type SupportRepository interface {
	Create(ctx context.Context, msg *model.SupportMessage) error

	// ListThread returns the most recent `limit` messages in chronological
	// order (oldest first) — the order the chat UI renders.
	ListThread(ctx context.Context, patientID, caregiverID uuid.UUID, limit int) ([]model.SupportMessage, error)

	// MarkRead stamps every unread message in the thread that the reader did
	// NOT send. Reading your own message is meaningless.
	MarkRead(ctx context.Context, patientID, caregiverID, readerID uuid.UUID, at time.Time) error

	// CountUnreadFor counts messages addressed to readerID across the thread.
	CountUnreadFor(ctx context.Context, patientID, caregiverID, readerID uuid.UUID) (int64, error)

	// CountUnreadByPatient returns, for one caregiver, the unread count per
	// patient — one query for the whole caregiver dashboard instead of one per
	// person they support.
	CountUnreadByPatient(ctx context.Context, caregiverID uuid.UUID) (map[uuid.UUID]int64, error)

	// LastMessageAt reports when the thread last saw any activity.
	LastMessageAt(ctx context.Context, patientID, caregiverID uuid.UUID) (*time.Time, error)
}

type supportRepository struct {
	db *gorm.DB
}

func NewSupportRepository(db *gorm.DB) SupportRepository {
	return &supportRepository{db: db}
}

func (r *supportRepository) Create(ctx context.Context, msg *model.SupportMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *supportRepository) ListThread(
	ctx context.Context,
	patientID, caregiverID uuid.UUID,
	limit int,
) ([]model.SupportMessage, error) {
	var messages []model.SupportMessage

	// Take the newest N, then flip: an ascending query with a limit would
	// return the oldest N and hide the live conversation.
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND caregiver_id = ?", patientID, caregiverID).
		Order("created_at desc").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (r *supportRepository) MarkRead(
	ctx context.Context,
	patientID, caregiverID, readerID uuid.UUID,
	at time.Time,
) error {
	return r.db.WithContext(ctx).
		Model(&model.SupportMessage{}).
		Where("patient_id = ? AND caregiver_id = ? AND sender_id <> ? AND read_at IS NULL",
			patientID, caregiverID, readerID).
		Update("read_at", at).Error
}

func (r *supportRepository) CountUnreadFor(
	ctx context.Context,
	patientID, caregiverID, readerID uuid.UUID,
) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.SupportMessage{}).
		Where("patient_id = ? AND caregiver_id = ? AND sender_id <> ? AND read_at IS NULL",
			patientID, caregiverID, readerID).
		Count(&count).Error

	return count, err
}

func (r *supportRepository) CountUnreadByPatient(
	ctx context.Context,
	caregiverID uuid.UUID,
) (map[uuid.UUID]int64, error) {
	type row struct {
		PatientID uuid.UUID
		Total     int64
	}

	var rows []row

	// Unread *for the caregiver* means messages the patient sent, so the
	// sender filter is the caregiver's own id.
	err := r.db.WithContext(ctx).
		Model(&model.SupportMessage{}).
		Select("patient_id, count(*) AS total").
		Where("caregiver_id = ? AND sender_id <> ? AND read_at IS NULL", caregiverID, caregiverID).
		Group("patient_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(rows))
	for _, item := range rows {
		counts[item.PatientID] = item.Total
	}
	return counts, nil
}

func (r *supportRepository) LastMessageAt(
	ctx context.Context,
	patientID, caregiverID uuid.UUID,
) (*time.Time, error) {
	var at *time.Time

	err := r.db.WithContext(ctx).
		Model(&model.SupportMessage{}).
		Select("max(created_at)").
		Where("patient_id = ? AND caregiver_id = ?", patientID, caregiverID).
		Scan(&at).Error

	return at, err
}
