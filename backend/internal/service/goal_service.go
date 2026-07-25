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

// goalUpdateLimit bounds how much of a goal's progress log is returned.
const goalUpdateLimit = 100

// --- DTOs (mirrored by frontend/src/types/anchorOneTypes.ts) -----------------

// GoalData is one goal as every screen sees it. Progress is computed server-side
// so the user's dashboard, the goal card and the caregiver's view can never
// disagree about how far along something is.
type GoalData struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Category        string     `json:"category"`
	Status          string     `json:"status"`
	TargetValue     int        `json:"target_value"`
	CurrentValue    int        `json:"current_value"`
	Unit            string     `json:"unit"`
	ProgressPercent int        `json:"progress_percent"`
	TargetDate      *time.Time `json:"target_date"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedByRole   string     `json:"created_by_role"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// DaysRemaining is nil when the goal has no target date. It goes negative
	// when a date has passed, which the UI renders as "overdue" rather than
	// hiding.
	DaysRemaining *int `json:"days_remaining"`
}

// GoalUpdateData is one entry of a goal's progress feed.
type GoalUpdateData struct {
	ID         uuid.UUID `json:"id"`
	GoalID     uuid.UUID `json:"goal_id"`
	AuthorID   uuid.UUID `json:"author_id"`
	AuthorName string    `json:"author_name"`
	AuthorRole string    `json:"author_role"`
	Kind       string    `json:"kind"`
	Value      int       `json:"value"`
	Delta      int       `json:"delta"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

// GoalDetail is a goal plus its progress log.
type GoalDetail struct {
	Goal    GoalData         `json:"goal"`
	Updates []GoalUpdateData `json:"updates"`
}

// GoalSummary is the roll-up the dashboards show: how many goals are in play
// and how far along the plan is overall.
type GoalSummary struct {
	Active          int    `json:"active"`
	Completed       int    `json:"completed"`
	Paused          int    `json:"paused"`
	Archived        int    `json:"archived"`
	Total           int    `json:"total"`
	AverageProgress int    `json:"average_progress"`
	NextGoalTitle   string `json:"next_goal_title"`
}

// GoalInput is the editable shape of a goal. A nil pointer on an update means
// "leave this field alone", which is what lets a caregiver's encouragement and
// the user's own edit touch the same record without clobbering each other.
type GoalInput struct {
	Title       string
	Description string
	Category    string
	TargetValue int
	Unit        string
	TargetDate  *time.Time
}

// GoalPatch is a partial update. Every field is optional.
type GoalPatch struct {
	Title        *string
	Description  *string
	Category     *string
	Status       *string
	TargetValue  *int
	CurrentValue *int
	Unit         *string
	TargetDate   *time.Time

	// ClearTargetDate distinguishes "do not touch the date" (nil TargetDate,
	// false here) from "remove the date" (true).
	ClearTargetDate bool
}

// ProgressInput logs movement on a goal.
type ProgressInput struct {
	// Exactly one of Delta or Value is applied: Delta nudges ("+1 day"), Value
	// sets an absolute position ("I'm at 42"). Delta wins when both are given.
	Delta *int
	Value *int
	Note  string
	Kind  string
}

// GoalService owns the recovery plan: the goals a person is working towards and
// the progress logged against them.
//
// Every method takes the acting user, not just an id, because a caregiver may
// read and encourage — but only the person in recovery may change what their
// own plan actually is.
type GoalService interface {
	List(ctx context.Context, actor Actor, patientID uuid.UUID) ([]GoalData, error)
	Summary(ctx context.Context, userID uuid.UUID) (*GoalSummary, error)
	Get(ctx context.Context, actor Actor, goalID uuid.UUID) (*GoalDetail, error)
	Create(ctx context.Context, actor Actor, patientID uuid.UUID, in GoalInput) (*GoalData, error)
	Update(ctx context.Context, actor Actor, goalID uuid.UUID, patch GoalPatch) (*GoalData, error)
	Delete(ctx context.Context, actor Actor, goalID uuid.UUID) error
	LogProgress(ctx context.Context, actor Actor, goalID uuid.UUID, in ProgressInput) (*GoalDetail, error)
}

type goalService struct {
	careAccess
	goals repository.GoalRepository
	users repository.UserRepository
	log   *zap.Logger
	now   func() time.Time
}

func NewGoalService(
	goals repository.GoalRepository,
	profiles repository.RecoverAIRepository,
	users repository.UserRepository,
	log *zap.Logger,
) GoalService {
	return &goalService{
		careAccess: careAccess{profiles: profiles},
		goals:      goals,
		users:      users,
		log:        log,
		now:        time.Now,
	}
}

func (s *goalService) List(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
) ([]GoalData, error) {
	if _, err := s.requireOneOf(
		ctx, actor, patientID, RelationSelf, RelationCaregiver, RelationAdmin,
	); err != nil {
		return nil, err
	}

	goals, err := s.goals.ListByUser(ctx, patientID, nil)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return s.toGoalDataList(goals), nil
}

func (s *goalService) Summary(ctx context.Context, userID uuid.UUID) (*GoalSummary, error) {
	counts, err := s.goals.CountByUserStatus(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	summary := &GoalSummary{
		Active:    int(counts[model.GoalStatusActive]),
		Completed: int(counts[model.GoalStatusCompleted]),
		Paused:    int(counts[model.GoalStatusPaused]),
		Archived:  int(counts[model.GoalStatusArchived]),
	}
	summary.Total = summary.Active + summary.Completed + summary.Paused + summary.Archived

	// Average progress is over open goals only: finished goals would pin it at
	// 100% and make a stalled plan look healthy.
	open, err := s.goals.ListByUser(ctx, userID, []string{model.GoalStatusActive})
	if err != nil {
		return nil, apperr.Internal(err)
	}

	if len(open) > 0 {
		total := 0
		soonest := -1

		for i := range open {
			total += open[i].ProgressPercent()

			// "Next" goal = the open one with the nearest target date, falling
			// back to the first open goal when none of them have dates.
			if soonest < 0 {
				soonest = i
				continue
			}
			if isSooner(open[i].TargetDate, open[soonest].TargetDate) {
				soonest = i
			}
		}

		summary.AverageProgress = total / len(open)
		if soonest >= 0 {
			summary.NextGoalTitle = open[soonest].Title
		}
	}

	return summary, nil
}

// isSooner reports whether a should sort before b, treating "no date" as the
// far future so a dated goal always wins.
func isSooner(a, b *time.Time) bool {
	switch {
	case a == nil:
		return false
	case b == nil:
		return true
	default:
		return a.Before(*b)
	}
}

func (s *goalService) Get(ctx context.Context, actor Actor, goalID uuid.UUID) (*GoalDetail, error) {
	goal, err := s.loadGoal(ctx, actor, goalID, RelationSelf, RelationCaregiver, RelationAdmin)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, goal)
}

func (s *goalService) Create(
	ctx context.Context,
	actor Actor,
	patientID uuid.UUID,
	in GoalInput,
) (*GoalData, error) {
	// A caregiver may suggest a goal for someone they support; an admin may
	// not, because a plan imposed by an administrator is not a recovery plan.
	relation, err := s.requireOneOf(ctx, actor, patientID, RelationSelf, RelationCaregiver)
	if err != nil {
		return nil, err
	}

	title := model.NormalizeGoalTitle(in.Title)
	if title == "" {
		return nil, apperr.Validation(apperr.Fields{"title": {"is required"}})
	}

	category := normalizeCategory(in.Category)

	// A goal with no target is not measurable, so it defaults to a simple
	// one-step "done / not done" rather than being rejected.
	target := in.TargetValue
	if target < 1 {
		target = 1
	}

	goal := &model.RecoveryGoal{
		UserID:        patientID,
		Title:         title,
		Description:   strings.TrimSpace(in.Description),
		Category:      category,
		Status:        model.GoalStatusActive,
		TargetValue:   target,
		CurrentValue:  0,
		Unit:          normalizeUnit(in.Unit),
		TargetDate:    in.TargetDate,
		CreatedByRole: relation.AuthorRole(),
	}

	if err := s.goals.Create(ctx, goal); err != nil {
		return nil, apperr.Internal(err)
	}

	s.log.Info("recovery goal created",
		zap.String("goal_id", goal.ID.String()),
		zap.String("user_id", patientID.String()),
		zap.String("created_by_role", goal.CreatedByRole),
	)

	data := s.toGoalData(goal)
	return &data, nil
}

func (s *goalService) Update(
	ctx context.Context,
	actor Actor,
	goalID uuid.UUID,
	patch GoalPatch,
) (*GoalData, error) {
	// Only the person in recovery edits the substance of their own plan.
	goal, err := s.loadGoal(ctx, actor, goalID, RelationSelf)
	if err != nil {
		return nil, err
	}

	if patch.Title != nil {
		title := model.NormalizeGoalTitle(*patch.Title)
		if title == "" {
			return nil, apperr.Validation(apperr.Fields{"title": {"is required"}})
		}
		goal.Title = title
	}
	if patch.Description != nil {
		goal.Description = strings.TrimSpace(*patch.Description)
	}
	if patch.Category != nil {
		goal.Category = normalizeCategory(*patch.Category)
	}
	if patch.Unit != nil {
		goal.Unit = normalizeUnit(*patch.Unit)
	}
	if patch.TargetValue != nil {
		if *patch.TargetValue < 1 {
			return nil, apperr.Validation(apperr.Fields{"target_value": {"must be at least 1"}})
		}
		goal.TargetValue = *patch.TargetValue
	}
	if patch.CurrentValue != nil {
		if *patch.CurrentValue < 0 {
			return nil, apperr.Validation(apperr.Fields{"current_value": {"must be 0 or more"}})
		}
		goal.CurrentValue = *patch.CurrentValue
	}
	if patch.ClearTargetDate {
		goal.TargetDate = nil
	} else if patch.TargetDate != nil {
		goal.TargetDate = patch.TargetDate
	}
	if patch.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*patch.Status))
		if !model.ValidGoalStatus(status) {
			return nil, apperr.Validation(apperr.Fields{
				"status": {"must be one of: " + strings.Join(model.GoalStatuses, ", ")},
			})
		}
		s.applyStatus(goal, status)
	}

	// Reaching the target completes the goal without the user having to say so
	// twice — the same rule that runs when progress is logged.
	s.settleCompletion(goal)

	if err := s.goals.Update(ctx, goal); err != nil {
		return nil, apperr.Internal(err)
	}

	data := s.toGoalData(goal)
	return &data, nil
}

func (s *goalService) Delete(ctx context.Context, actor Actor, goalID uuid.UUID) error {
	if _, err := s.loadGoal(ctx, actor, goalID, RelationSelf); err != nil {
		return err
	}

	if err := s.goals.Delete(ctx, goalID); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (s *goalService) LogProgress(
	ctx context.Context,
	actor Actor,
	goalID uuid.UUID,
	in ProgressInput,
) (*GoalDetail, error) {
	// A caregiver may add a note or encouragement, but may not move someone
	// else's numbers. Progress is self-reported by definition.
	relations := []CareRelation{RelationSelf, RelationCaregiver}
	goal, err := s.loadGoal(ctx, actor, goalID, relations...)
	if err != nil {
		return nil, err
	}

	relation, err := s.resolve(ctx, actor, goal.UserID)
	if err != nil {
		return nil, err
	}

	kind := normalizeUpdateKind(in.Kind)
	note := strings.TrimSpace(in.Note)

	moves := in.Delta != nil || in.Value != nil
	if relation != RelationSelf {
		if moves {
			return nil, apperr.Forbidden("Only the person in recovery can log progress on their goal.")
		}
		if kind == model.GoalUpdateKindProgress {
			kind = model.GoalUpdateKindEncouragement
		}
	}

	if !moves && note == "" {
		return nil, apperr.Validation(apperr.Fields{"note": {"is required when no progress is logged"}})
	}

	previous := goal.CurrentValue
	if moves {
		next := previous
		switch {
		case in.Delta != nil:
			next = previous + *in.Delta
		case in.Value != nil:
			next = *in.Value
		}

		// Progress cannot go below zero or above the target: the log records
		// what was asked for, the goal records what is possible.
		if next < 0 {
			next = 0
		}
		if next > goal.TargetValue {
			next = goal.TargetValue
		}
		goal.CurrentValue = next
		kind = model.GoalUpdateKindProgress
	}

	s.settleCompletion(goal)

	if err := s.goals.Update(ctx, goal); err != nil {
		return nil, apperr.Internal(err)
	}

	entry := &model.GoalUpdate{
		GoalID:     goal.ID,
		AuthorID:   actor.ID,
		AuthorRole: relation.AuthorRole(),
		Kind:       kind,
		Value:      goal.CurrentValue,
		Delta:      goal.CurrentValue - previous,
		Note:       note,
	}
	if err := s.goals.CreateUpdate(ctx, entry); err != nil {
		return nil, apperr.Internal(err)
	}

	return s.detail(ctx, goal)
}

// --- internals ---------------------------------------------------------------

// loadGoal fetches a goal and authorises the actor against its owner in one
// step, so no caller can forget the second half.
func (s *goalService) loadGoal(
	ctx context.Context,
	actor Actor,
	goalID uuid.UUID,
	allowed ...CareRelation,
) (*model.RecoveryGoal, error) {
	goal, err := s.goals.GetByID(ctx, goalID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if goal == nil {
		return nil, apperr.NotFound("goal")
	}

	if _, err := s.requireOneOf(ctx, actor, goal.UserID, allowed...); err != nil {
		return nil, err
	}
	return goal, nil
}

// applyStatus moves a goal between states, keeping CompletedAt consistent.
func (s *goalService) applyStatus(goal *model.RecoveryGoal, status string) {
	goal.Status = status

	switch status {
	case model.GoalStatusCompleted:
		if goal.CompletedAt == nil {
			at := s.now().UTC()
			goal.CompletedAt = &at
		}
		// Completing a goal fills its bar — otherwise a plan shows "done" at 60%.
		if goal.CurrentValue < goal.TargetValue {
			goal.CurrentValue = goal.TargetValue
		}
	default:
		// Reopening a goal clears the completion stamp, so a later re-completion
		// records when it actually happened.
		goal.CompletedAt = nil
	}
}

// settleCompletion completes a goal that has reached its target, and reopens
// one whose target was raised beyond its current value.
func (s *goalService) settleCompletion(goal *model.RecoveryGoal) {
	switch {
	case goal.Status == model.GoalStatusActive && goal.CurrentValue >= goal.TargetValue:
		s.applyStatus(goal, model.GoalStatusCompleted)
	case goal.Status == model.GoalStatusCompleted && goal.CurrentValue < goal.TargetValue:
		s.applyStatus(goal, model.GoalStatusActive)
	}
}

func (s *goalService) detail(ctx context.Context, goal *model.RecoveryGoal) (*GoalDetail, error) {
	updates, err := s.goals.ListUpdates(ctx, goal.ID, goalUpdateLimit)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	feed := make([]GoalUpdateData, 0, len(updates))
	for i := range updates {
		feed = append(feed, toGoalUpdateData(&updates[i]))
	}

	return &GoalDetail{Goal: s.toGoalData(goal), Updates: feed}, nil
}

func (s *goalService) toGoalDataList(goals []model.RecoveryGoal) []GoalData {
	out := make([]GoalData, 0, len(goals))
	for i := range goals {
		out = append(out, s.toGoalData(&goals[i]))
	}
	return out
}

func (s *goalService) toGoalData(goal *model.RecoveryGoal) GoalData {
	data := GoalData{
		ID:              goal.ID,
		UserID:          goal.UserID,
		Title:           goal.Title,
		Description:     goal.Description,
		Category:        goal.Category,
		Status:          goal.Status,
		TargetValue:     goal.TargetValue,
		CurrentValue:    goal.CurrentValue,
		Unit:            goal.Unit,
		ProgressPercent: goal.ProgressPercent(),
		TargetDate:      goal.TargetDate,
		CompletedAt:     goal.CompletedAt,
		CreatedByRole:   goal.CreatedByRole,
		CreatedAt:       goal.CreatedAt,
		UpdatedAt:       goal.UpdatedAt,
	}

	if goal.TargetDate != nil {
		days := int(goal.TargetDate.UTC().Truncate(day).Sub(s.now().UTC().Truncate(day)) / day)
		data.DaysRemaining = &days
	}

	return data
}

func toGoalUpdateData(update *model.GoalUpdate) GoalUpdateData {
	data := GoalUpdateData{
		ID:         update.ID,
		GoalID:     update.GoalID,
		AuthorID:   update.AuthorID,
		AuthorRole: update.AuthorRole,
		Kind:       update.Kind,
		Value:      update.Value,
		Delta:      update.Delta,
		Note:       update.Note,
		CreatedAt:  update.CreatedAt,
	}
	if update.Author != nil {
		data.AuthorName = update.Author.FullName()
	}
	return data
}

func normalizeCategory(raw string) string {
	category := strings.ToLower(strings.TrimSpace(raw))
	if !model.ValidGoalCategory(category) {
		return model.GoalCategoryOther
	}
	return category
}

func normalizeUnit(raw string) string {
	unit := strings.TrimSpace(raw)
	if unit == "" {
		return "days"
	}
	if len(unit) > 50 {
		return unit[:50]
	}
	return unit
}

func normalizeUpdateKind(raw string) string {
	switch kind := strings.ToLower(strings.TrimSpace(raw)); kind {
	case model.GoalUpdateKindNote,
		model.GoalUpdateKindEncouragement,
		model.GoalUpdateKindStatus,
		model.GoalUpdateKindProgress:
		return kind
	default:
		return model.GoalUpdateKindProgress
	}
}
