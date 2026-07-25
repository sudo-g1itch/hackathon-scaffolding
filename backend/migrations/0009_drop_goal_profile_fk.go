package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// m0009DropGoalProfileFK removes a constraint GORM invents from the
// `RecoveryProfile.Goals` has-many:
//
//	fk_recovery_profiles_goals: recovery_goals.user_id -> recovery_profiles.user_id
//
// That relation exists so open goals can be preloaded into AI prompts. As a
// database constraint it says something the app does not mean — that a goal may
// only exist if its owner already has a recovery_profiles row. A brand-new user
// who opens "My Goals" before ever touching their recovery plan has no such
// row, so their first goal fails to insert.
//
// The correct rule is already in place: fk_recovery_goals_user ties a goal to a
// USER, with ON DELETE CASCADE. This step drops the accidental one and leaves
// that as the single truth.
//
// Idempotent, and safe to run before or after AutoMigrate re-creates it, since
// AutoMigrate only adds the constraint when it is absent — and by this point
// every AutoMigrate step has already run.
func m0009DropGoalProfileFK() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "0009_drop_goal_profile_fk",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec(
				`ALTER TABLE recovery_goals DROP CONSTRAINT IF EXISTS fk_recovery_profiles_goals`,
			).Error
		},
		Rollback: func(tx *gorm.DB) error {
			// Deliberately not restored: re-adding it would fail against any
			// data created since, where goals legitimately exist without a
			// profile row.
			return nil
		},
	}
}
