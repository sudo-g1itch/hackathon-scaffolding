package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// m0006ExtendRecoveraiTables persists the parts of the AI output that were
// previously computed and thrown away: a check-in's recommended actions and an
// emergency plan's grounding exercise + encouraging message. It also adds the
// composite indexes the dashboard/streak/timeline queries read on every load.
//
// Every statement is idempotent (IF NOT EXISTS): step 0004 AutoMigrates the
// live model structs, so on a database created *after* those structs gained
// these fields the columns already exist.
func m0006ExtendRecoveraiTables() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "0006_extend_recoverai_tables",
		Migrate: func(tx *gorm.DB) error {
			stmts := []string{
				`ALTER TABLE checkins ADD COLUMN IF NOT EXISTS recommended_actions jsonb`,
				`ALTER TABLE checkins ADD COLUMN IF NOT EXISTS source varchar(20) NOT NULL DEFAULT 'voice'`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS grounding_exercise text`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS encouraging_message text`,
				`CREATE INDEX IF NOT EXISTS idx_checkins_user_created
					ON checkins (user_id, created_at DESC)`,
				`CREATE INDEX IF NOT EXISTS idx_coach_messages_user_created
					ON coach_messages (user_id, created_at DESC)`,
				`CREATE INDEX IF NOT EXISTS idx_emergency_logs_user_created
					ON emergency_logs (user_id, created_at DESC)`,
			}

			for _, stmt := range stmts {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			stmts := []string{
				`DROP INDEX IF EXISTS idx_emergency_logs_user_created`,
				`DROP INDEX IF EXISTS idx_coach_messages_user_created`,
				`DROP INDEX IF EXISTS idx_checkins_user_created`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS encouraging_message`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS grounding_exercise`,
				`ALTER TABLE checkins DROP COLUMN IF EXISTS source`,
				`ALTER TABLE checkins DROP COLUMN IF EXISTS recommended_actions`,
			}

			for _, stmt := range stmts {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		},
	}
}
