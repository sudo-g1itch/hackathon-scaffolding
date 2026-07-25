package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Goal lifecycle states. A goal is always in exactly one of these.
const (
	GoalStatusActive    = "active"
	GoalStatusCompleted = "completed"
	GoalStatusPaused    = "paused"
	GoalStatusArchived  = "archived"
)

// GoalStatuses is the whitelist every layer validates against, so a status can
// never be invented by a client or a prompt.
var GoalStatuses = []string{
	GoalStatusActive,
	GoalStatusCompleted,
	GoalStatusPaused,
	GoalStatusArchived,
}

// Goal categories. These group a recovery plan into the life areas the PRD
// talks about, and drive the icon/colour the UI picks.
const (
	GoalCategorySobriety = "sobriety"
	GoalCategoryHealth   = "health"
	GoalCategoryRoutine  = "routine"
	GoalCategorySocial   = "social"
	GoalCategoryWork     = "work"
	GoalCategoryOther    = "other"
)

var GoalCategories = []string{
	GoalCategorySobriety,
	GoalCategoryHealth,
	GoalCategoryRoutine,
	GoalCategorySocial,
	GoalCategoryWork,
	GoalCategoryOther,
}

// Authors of a goal or a progress update. A caregiver may suggest a goal and
// leave encouragement on it; the record always says which of the two acted, so
// the UI never attributes a caregiver's note to the person in recovery.
const (
	AuthorRoleUser      = "user"
	AuthorRoleCaregiver = "caregiver"
)

// Progress-update kinds. Progress moves the number; note and encouragement do
// not — they are the conversation that happens around the number.
const (
	GoalUpdateKindProgress      = "progress"
	GoalUpdateKindNote          = "note"
	GoalUpdateKindEncouragement = "encouragement"
	GoalUpdateKindStatus        = "status"
)

// ValidGoalStatus reports whether s is a status the app recognises.
func ValidGoalStatus(s string) bool { return contains(GoalStatuses, s) }

// ValidGoalCategory reports whether s is a category the app recognises.
func ValidGoalCategory(s string) bool { return contains(GoalCategories, s) }

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// RecoveryGoal is one measurable commitment inside a person's recovery plan.
//
// A user may hold many goals at once ("90 days sober", "gym twice a week",
// "call my sponsor daily"), which is why progress lives on the goal rather than
// on the profile. Every goal is measured the same way — a current value moving
// towards a target — so one progress bar, one percentage and one streak
// calculation serve every kind of goal.
type RecoveryGoal struct {
	BaseModel
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User        *User     `json:"user,omitempty"`
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `gorm:"type:varchar(50);not null;default:'other'" json:"category"`
	Status      string    `gorm:"type:varchar(50);not null;default:'active';index" json:"status"`

	// TargetValue is what "done" means, in Unit (e.g. 90 days, 12 sessions).
	// CurrentValue is where the person is now. TargetValue is never zero — the
	// service defaults it to 1 so a goal is always a completable thing.
	TargetValue  int    `gorm:"not null;default:1" json:"target_value"`
	CurrentValue int    `gorm:"not null;default:0" json:"current_value"`
	Unit         string `gorm:"type:varchar(50);not null;default:'days'" json:"unit"`

	TargetDate  *time.Time `json:"target_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// CreatedByRole records whether the person themself or their caregiver put
	// this goal on the plan. A caregiver-suggested goal is still the user's to
	// accept, edit or archive.
	CreatedByRole string `gorm:"type:varchar(50);not null;default:'user'" json:"created_by_role"`

	Updates []GoalUpdate `gorm:"foreignKey:GoalID" json:"updates,omitempty"`
}

// ProgressPercent is the single definition of "how far along is this goal",
// clamped to 0-100 so a target that was later lowered cannot report 340%.
func (g *RecoveryGoal) ProgressPercent() int {
	if g.TargetValue <= 0 {
		return 0
	}

	percent := g.CurrentValue * 100 / g.TargetValue
	switch {
	case percent < 0:
		return 0
	case percent > 100:
		return 100
	default:
		return percent
	}
}

// IsOpen reports whether the goal still counts towards "what am I working on".
func (g *RecoveryGoal) IsOpen() bool { return g.Status == GoalStatusActive }

// NormalizeGoalTitle trims a title and caps it at the column width, so an
// over-long paste is stored rather than rejected by Postgres.
func NormalizeGoalTitle(raw string) string {
	title := strings.TrimSpace(raw)
	if len(title) > 200 {
		return title[:200]
	}
	return title
}

// GoalUpdate is one entry in a goal's progress log: a value change, a note from
// the person, or a word of encouragement from their caregiver. Keeping them in
// one table means the goal detail view is a single chronological feed rather
// than three lists the UI has to interleave.
type GoalUpdate struct {
	BaseModel
	GoalID     uuid.UUID `gorm:"type:uuid;not null;index" json:"goal_id"`
	AuthorID   uuid.UUID `gorm:"type:uuid;not null;index" json:"author_id"`
	Author     *User     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	AuthorRole string    `gorm:"type:varchar(50);not null;default:'user'" json:"author_role"`
	Kind       string    `gorm:"type:varchar(50);not null;default:'progress'" json:"kind"`

	// Value is the goal's current value *after* this update, and Delta how much
	// it moved. Storing both means the feed can say "+3 (12/90)" without
	// replaying every earlier row.
	Value int    `gorm:"not null;default:0" json:"value"`
	Delta int    `gorm:"not null;default:0" json:"delta"`
	Note  string `gorm:"type:text" json:"note"`
}
