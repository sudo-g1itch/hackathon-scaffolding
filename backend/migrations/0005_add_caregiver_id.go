package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// m0005AddCaregiverID links a recovery profile to a caregiver account.
//
// IDEMPOTENCY (why this step drops before it adds): step 0004 runs
// `AutoMigrate` against the LIVE model structs, not a snapshot of how they
// looked in 0004's day. `model.RecoveryProfile` has since gained the
// `Caregiver` belongs-to relation, so on a database created today 0004 already
// emits both the `caregiver_id` column AND a `fk_recovery_profiles_caregiver`
// constraint — and this step's bare `ADD CONSTRAINT` then fails with SQLSTATE
// 42710, blocking every later migration.
//
// This was edited after being applied, which rule 3.7 otherwise forbids. It is
// the one case the rule cannot cover: the step is unrunnable on a fresh
// database, and a new step cannot repair it because the run dies here first.
// The edit is a no-op wherever 0005 already succeeded — gormigrate will not
// re-run it — and the DROP/ADD pair guarantees the ON DELETE SET NULL
// behaviour this app relies on, whichever way the constraint was created.
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

			// Replace rather than add: AutoMigrate's version of this constraint
			// has no ON DELETE clause, so deleting a caregiver account would be
			// blocked instead of unlinking the people they support.
			stmts := []string{
				`ALTER TABLE recovery_profiles
					DROP CONSTRAINT IF EXISTS fk_recovery_profiles_caregiver`,
				`ALTER TABLE recovery_profiles
					ADD CONSTRAINT fk_recovery_profiles_caregiver
					FOREIGN KEY (caregiver_id)
					REFERENCES users(id)
					ON DELETE SET NULL`,
			}

			for _, stmt := range stmts {
				if err := tx.Exec(stmt).Error; err != nil {
					return err
				}
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
