package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
)

// GoalRepository owns every query behind the multi-goal recovery plan.
//
// As elsewhere in this codebase, a lookup that finds nothing returns
// (nil, nil): "this goal does not exist" is the caller's decision to turn into
// a 404, not the repository's.
type GoalRepository interface {
	Create(ctx context.Context, goal *model.RecoveryGoal) error
	Update(ctx context.Context, goal *model.RecoveryGoal) error
	Delete(ctx context.Context, goalID uuid.UUID) error

	GetByID(ctx context.Context, goalID uuid.UUID) (*model.RecoveryGoal, error)

	// ListByUser returns a user's goals newest first. An empty statuses slice
	// means "every status".
	ListByUser(ctx context.Context, userID uuid.UUID, statuses []string) ([]model.RecoveryGoal, error)

	// CountByUserStatus returns how many goals the user holds per status —
	// one grouped query rather than one count per status.
	CountByUserStatus(ctx context.Context, userID uuid.UUID) (map[string]int64, error)

	CreateUpdate(ctx context.Context, update *model.GoalUpdate) error
	ListUpdates(ctx context.Context, goalID uuid.UUID, limit int) ([]model.GoalUpdate, error)
}

type goalRepository struct {
	db *gorm.DB
}

func NewGoalRepository(db *gorm.DB) GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) Create(ctx context.Context, goal *model.RecoveryGoal) error {
	return r.db.WithContext(ctx).Create(goal).Error
}

// Update writes the mutable columns explicitly. A blanket Save would also
// persist the preloaded Updates slice and could null a column the caller never
// touched.
func (r *goalRepository) Update(ctx context.Context, goal *model.RecoveryGoal) error {
	return r.db.WithContext(ctx).
		Model(&model.RecoveryGoal{}).
		Where("id = ?", goal.ID).
		Select(
			"title", "description", "category", "status", "target_value",
			"current_value", "unit", "target_date", "completed_at", "updated_at",
		).
		Updates(goal).Error
}

func (r *goalRepository) Delete(ctx context.Context, goalID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.RecoveryGoal{}, "id = ?", goalID).Error
}

func (r *goalRepository) GetByID(ctx context.Context, goalID uuid.UUID) (*model.RecoveryGoal, error) {
	var goal model.RecoveryGoal

	err := r.db.WithContext(ctx).Where("id = ?", goalID).First(&goal).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &goal, nil
}

func (r *goalRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	statuses []string,
) ([]model.RecoveryGoal, error) {
	var goals []model.RecoveryGoal

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	err := query.Order("created_at desc").Find(&goals).Error

	return goals, err
}

func (r *goalRepository) CountByUserStatus(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	type row struct {
		Status string
		Total  int64
	}

	var rows []row

	err := r.db.WithContext(ctx).
		Model(&model.RecoveryGoal{}).
		Select("status, count(*) AS total").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, item := range rows {
		counts[item.Status] = item.Total
	}
	return counts, nil
}

func (r *goalRepository) CreateUpdate(ctx context.Context, update *model.GoalUpdate) error {
	return r.db.WithContext(ctx).Create(update).Error
}

// ListUpdates returns a goal's progress log newest first, with the author
// preloaded so the feed can say who wrote each entry.
func (r *goalRepository) ListUpdates(
	ctx context.Context,
	goalID uuid.UUID,
	limit int,
) ([]model.GoalUpdate, error) {
	var updates []model.GoalUpdate

	err := r.db.WithContext(ctx).
		Preload("Author").
		Where("goal_id = ?", goalID).
		Order("created_at desc").
		Limit(limit).
		Find(&updates).Error

	return updates, err
}
