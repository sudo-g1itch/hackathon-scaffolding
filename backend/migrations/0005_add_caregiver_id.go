package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func m0005AddCaregiverID() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "0005_add_caregiver_id",
		Migrate: func(tx *gorm.DB) error {
			type RecoveryProfile struct {
				CaregiverID *string `gorm:"type:uuid;index"`
			}

			if err := tx.AutoMigrate(&RecoveryProfile{}); err != nil {
				return err
			}

			// Add foreign key constraint
			if err := tx.Exec(`
				ALTER TABLE recovery_profiles
				ADD CONSTRAINT fk_recovery_profiles_caregiver
				FOREIGN KEY (caregiver_id)
				REFERENCES users(id)
				ON DELETE SET NULL;
			`).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE recovery_profiles DROP CONSTRAINT IF EXISTS fk_recovery_profiles_caregiver;`).Error; err != nil {
				return err
			}
			if err := tx.Migrator().DropColumn("recovery_profiles", "caregiver_id"); err != nil {
				return err
			}
			return nil
		},
	}
}
