package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// m0008EmergencyAlerts turns Emergency Mode from a plan the user reads into an
// alert they can actually send.
//
// emergency_logs gains what was sent (the chosen script), whether a location
// was attached, the voice note's transcript, and the caregiver's
// acknowledgement. support_messages gains a kind + emergency_id so an alert
// lands in the same conversation as everything else while still being
// distinguishable from an ordinary message.
//
// Every statement is idempotent: step 0004/0007 AutoMigrate the live model
// structs, so on a database created after those structs gained these fields the
// columns already exist.
func m0008EmergencyAlerts() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "0008_emergency_alerts",
		Migrate: func(tx *gorm.DB) error {
			stmts := []string{
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS sent_message text`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS shared_at timestamptz`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS caregiver_id uuid`,
				`ALTER TABLE emergency_logs
					ADD COLUMN IF NOT EXISTS share_location boolean NOT NULL DEFAULT false`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS location_lat double precision`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS location_lng double precision`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS audio_transcript text`,
				`ALTER TABLE emergency_logs ADD COLUMN IF NOT EXISTS acknowledged_at timestamptz`,
				`CREATE INDEX IF NOT EXISTS idx_emergency_logs_caregiver
					ON emergency_logs (caregiver_id, created_at DESC)
					WHERE shared_at IS NOT NULL`,

				`ALTER TABLE support_messages
					ADD COLUMN IF NOT EXISTS kind varchar(50) NOT NULL DEFAULT 'message'`,
				`ALTER TABLE support_messages ADD COLUMN IF NOT EXISTS emergency_id uuid`,
				`CREATE INDEX IF NOT EXISTS idx_support_messages_kind
					ON support_messages (kind)`,
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
				`DROP INDEX IF EXISTS idx_support_messages_kind`,
				`ALTER TABLE support_messages DROP COLUMN IF EXISTS emergency_id`,
				`ALTER TABLE support_messages DROP COLUMN IF EXISTS kind`,

				`DROP INDEX IF EXISTS idx_emergency_logs_caregiver`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS acknowledged_at`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS audio_transcript`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS location_lng`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS location_lat`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS share_location`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS caregiver_id`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS shared_at`,
				`ALTER TABLE emergency_logs DROP COLUMN IF EXISTS sent_message`,
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
