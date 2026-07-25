package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
)

// m0007CreateRecoveryPlanTables introduces the multi-goal recovery plan and the
// patient <-> caregiver support thread.
//
//   - recovery_goals   — many goals per user, each measured current/target.
//   - goal_updates     — the chronological progress log on a goal.
//   - support_messages — the shared conversation between a person and the
//     caregiver they linked.
//
// It also adds recovery_profiles.share_checkin_details: the explicit consent
// that lets a caregiver read a check-in's narrative. It defaults to FALSE, so
// existing users' privacy is unchanged by this migration.
func m0007CreateRecoveryPlanTables() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "0007_create_recovery_plan_tables",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(
				&model.RecoveryGoal{},
				&model.GoalUpdate{},
				&model.SupportMessage{},
			); err != nil {
				return err
			}

			stmts := []string{
				`ALTER TABLE recovery_profiles
					ADD COLUMN IF NOT EXISTS share_checkin_details boolean NOT NULL DEFAULT false`,

				// The goals list is always read as "this user's goals, newest
				// first", and the caregiver overview reads only open ones.
				`CREATE INDEX IF NOT EXISTS idx_recovery_goals_user_created
					ON recovery_goals (user_id, created_at DESC)`,
				`CREATE INDEX IF NOT EXISTS idx_recovery_goals_user_status
					ON recovery_goals (user_id, status)`,
				`CREATE INDEX IF NOT EXISTS idx_goal_updates_goal_created
					ON goal_updates (goal_id, created_at DESC)`,

				// A thread is (patient, caregiver) read chronologically; the
				// unread badge counts rows where read_at IS NULL.
				`CREATE INDEX IF NOT EXISTS idx_support_messages_thread
					ON support_messages (patient_id, caregiver_id, created_at DESC)`,
				`CREATE INDEX IF NOT EXISTS idx_support_messages_unread
					ON support_messages (patient_id, caregiver_id, sender_id)
					WHERE read_at IS NULL`,
			}

			for _, stmt := range stmts {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}

			// A goal belongs to a user and dies with them; an update belongs to
			// a goal and dies with it. Left to GORM these would be nullable
			// orphans, which the progress feed has no way to render.
			//
			// AutoMigrate also emits its own has-many constraint
			// (fk_recovery_goals_updates) with no ON DELETE clause. Leaving it
			// in place next to ours would mean a hard delete has to satisfy
			// both, and the cascade would never fire — so it is dropped and
			// ours is the single rule.
			fks := []string{
				`ALTER TABLE recovery_goals
					DROP CONSTRAINT IF EXISTS fk_recovery_goals_user`,
				`ALTER TABLE recovery_goals
					ADD CONSTRAINT fk_recovery_goals_user
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`,
				`ALTER TABLE goal_updates
					DROP CONSTRAINT IF EXISTS fk_recovery_goals_updates`,
				`ALTER TABLE goal_updates
					DROP CONSTRAINT IF EXISTS fk_goal_updates_goal`,
				`ALTER TABLE goal_updates
					ADD CONSTRAINT fk_goal_updates_goal
					FOREIGN KEY (goal_id) REFERENCES recovery_goals(id) ON DELETE CASCADE`,
			}

			for _, stmt := range fks {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(
				`ALTER TABLE recovery_profiles DROP COLUMN IF EXISTS share_checkin_details`,
			).Error; err != nil {
				return err
			}

			return tx.Migrator().DropTable(
				&model.SupportMessage{},
				&model.GoalUpdate{},
				&model.RecoveryGoal{},
			)
		},
	}
}
