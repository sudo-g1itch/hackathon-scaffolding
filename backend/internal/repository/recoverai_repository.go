package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
)

// RecoverAIRepository owns every query behind the recovery features.
//
// Lookup methods return (nil, nil) when a row simply does not exist: "this user
// has never checked in" is an ordinary state, not an error, and returning a
// zero-valued struct instead would make an empty check-in look like a real one.
type RecoverAIRepository interface {
	// Profile
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*model.RecoveryProfile, error)
	UpsertProfile(ctx context.Context, profile *model.RecoveryProfile) error

	// Checkins
	CreateCheckin(ctx context.Context, checkin *model.Checkin) error
	ListCheckins(ctx context.Context, userID uuid.UUID, limit int) ([]model.Checkin, error)
	GetLastCheckin(ctx context.Context, userID uuid.UUID) (*model.Checkin, error)
	CountCheckins(ctx context.Context, userID uuid.UUID) (int64, error)

	// CheckinDays lists the distinct local dates the user checked in on,
	// newest first — the input to the recovery-streak calculation.
	CheckinDays(ctx context.Context, userID uuid.UUID, limit int) ([]time.Time, error)

	// Coach
	CreateCoachMessage(ctx context.Context, msg *model.CoachMessage) error
	GetCoachHistory(ctx context.Context, userID uuid.UUID, limit int) ([]model.CoachMessage, error)

	// Emergency
	CreateEmergencyLog(ctx context.Context, log *model.EmergencyLog) error
	GetEmergencyLog(ctx context.Context, logID uuid.UUID) (*model.EmergencyLog, error)
	UpdateEmergencyLog(ctx context.Context, log *model.EmergencyLog) error
	ListEmergencyLogs(ctx context.Context, userID uuid.UUID, limit int) ([]model.EmergencyLog, error)
	CountEmergencyLogs(ctx context.Context, userID uuid.UUID) (int64, error)

	// Caregiver
	GetCaregiverPatients(ctx context.Context, caregiverID uuid.UUID) ([]model.RecoveryProfile, error)
	ListAvailableCaregivers(ctx context.Context, excludeUserID uuid.UUID) ([]model.User, error)
	SetCaregiver(ctx context.Context, userID uuid.UUID, caregiverID *uuid.UUID) error
}

type recoverAIRepository struct {
	db *gorm.DB
}

func NewRecoverAIRepository(db *gorm.DB) RecoverAIRepository {
	return &recoverAIRepository{db: db}
}

func (r *recoverAIRepository) GetProfileByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.RecoveryProfile, error) {
	var profile model.RecoveryProfile

	// Open goals are preloaded because every AI prompt is grounded in them —
	// the coach should know what the person is working towards, not just which
	// substance they named. Completed and archived goals are noise in a prompt.
	err := r.db.WithContext(ctx).
		Preload("Caregiver").
		Preload("Goals", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", model.GoalStatusActive).Order("created_at asc")
		}).
		Where("user_id = ?", userID).
		First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// UpsertProfile creates the profile row or updates the mutable columns of an
// existing one, keyed by the unique user_id. Done in a single statement so two
// concurrent calls cannot both decide to insert.
func (r *recoverAIRepository) UpsertProfile(ctx context.Context, profile *model.RecoveryProfile) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"goal", "substance", "caregiver_name", "caregiver_phone",
				"emergency_contact", "share_checkin_details", "updated_at",
			}),
		}).
		Create(profile).Error
}

func (r *recoverAIRepository) CreateCheckin(ctx context.Context, checkin *model.Checkin) error {
	return r.db.WithContext(ctx).Create(checkin).Error
}

func (r *recoverAIRepository) ListCheckins(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]model.Checkin, error) {
	var checkins []model.Checkin

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&checkins).Error

	return checkins, err
}

func (r *recoverAIRepository) GetLastCheckin(ctx context.Context, userID uuid.UUID) (*model.Checkin, error) {
	var checkin model.Checkin

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		First(&checkin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &checkin, nil
}

func (r *recoverAIRepository) CountCheckins(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.Checkin{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

func (r *recoverAIRepository) CheckinDays(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]time.Time, error) {
	var days []time.Time

	err := r.db.WithContext(ctx).
		Model(&model.Checkin{}).
		Where("user_id = ?", userID).
		Distinct("date_trunc('day', created_at) AS day").
		Order("day desc").
		Limit(limit).
		Pluck("day", &days).Error

	return days, err
}

func (r *recoverAIRepository) CreateCoachMessage(ctx context.Context, msg *model.CoachMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// GetCoachHistory returns the most recent `limit` messages in chronological
// order (oldest first), which is the order both the prompt and the chat UI want.
func (r *recoverAIRepository) GetCoachHistory(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]model.CoachMessage, error) {
	var history []model.CoachMessage

	// Take the newest N, then flip: a plain ascending query with a limit would
	// return the *oldest* N and lose the recent conversation.
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&history).Error
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}
	return history, nil
}

func (r *recoverAIRepository) CreateEmergencyLog(ctx context.Context, log *model.EmergencyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *recoverAIRepository) GetEmergencyLog(
	ctx context.Context,
	logID uuid.UUID,
) (*model.EmergencyLog, error) {
	var entry model.EmergencyLog

	err := r.db.WithContext(ctx).Where("id = ?", logID).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

// UpdateEmergencyLog writes only the columns the send flow touches. The plan
// itself (actions, grounding, encouragement) is a record of what was generated
// and is never rewritten afterwards.
func (r *recoverAIRepository) UpdateEmergencyLog(ctx context.Context, log *model.EmergencyLog) error {
	return r.db.WithContext(ctx).
		Model(&model.EmergencyLog{}).
		Where("id = ?", log.ID).
		Select(
			"sent_message", "shared_at", "caregiver_id", "share_location",
			"location_lat", "location_lng", "audio_transcript",
			"acknowledged_at", "updated_at",
		).
		Updates(log).Error
}

func (r *recoverAIRepository) ListEmergencyLogs(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]model.EmergencyLog, error) {
	var logs []model.EmergencyLog

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&logs).Error

	return logs, err
}

func (r *recoverAIRepository) CountEmergencyLogs(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.EmergencyLog{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

func (r *recoverAIRepository) GetCaregiverPatients(
	ctx context.Context,
	caregiverID uuid.UUID,
) ([]model.RecoveryProfile, error) {
	var profiles []model.RecoveryProfile

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("caregiver_id = ?", caregiverID).
		Find(&profiles).Error

	return profiles, err
}

// ListAvailableCaregivers returns the active caregiver accounts a user may
// assign, excluding the requester so nobody links themselves.
func (r *recoverAIRepository) ListAvailableCaregivers(
	ctx context.Context,
	excludeUserID uuid.UUID,
) ([]model.User, error) {
	var caregivers []model.User

	err := r.db.WithContext(ctx).
		Where("role = ? AND is_active = ? AND id <> ?", model.RoleCaregiver, true, excludeUserID).
		Order("first_name asc, last_name asc").
		Find(&caregivers).Error

	return caregivers, err
}

// SetCaregiver links (or, with a nil caregiverID, unlinks) the user's caregiver.
func (r *recoverAIRepository) SetCaregiver(
	ctx context.Context,
	userID uuid.UUID,
	caregiverID *uuid.UUID,
) error {
	return r.db.WithContext(ctx).
		Model(&model.RecoveryProfile{}).
		Where("user_id = ?", userID).
		Update("caregiver_id", caregiverID).Error
}
