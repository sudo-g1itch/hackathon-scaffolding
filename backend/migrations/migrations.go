// Package migrations owns the database schema's history.
//
// Rules:
//   - Every change is a gormigrate step with a stable, sortable ID.
//   - Steps are APPEND-ONLY. Never edit a step that has been applied.
//   - Bare AutoMigrate is never used.
package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// All returns every migration in the order it must be applied.
// Add new steps to the END of this slice.
func All() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		m0001EnableExtensions(),
		m0002CreateUsersTable(),
		m0003CreateRolesPermissions(),
	}
}

func options() *gormigrate.Options {
	return &gormigrate.Options{
		TableName:                 "schema_migrations",
		IDColumnName:              "id",
		IDColumnSize:              255,
		UseTransaction:            true,
		ValidateUnknownMigrations: true,
	}
}

func New(db *gorm.DB) *gormigrate.Gormigrate {
	return gormigrate.New(db, options(), All())
}

// Up applies every pending migration.
func Up(db *gorm.DB, log *zap.Logger) error {
	applied, err := appliedIDs(db)
	if err != nil {
		return err
	}

	log.Info("applying migrations",
		zap.Int("defined", len(All())),
		zap.Int("already_applied", len(applied)),
	)

	if err := New(db).Migrate(); err != nil {
		return fmt.Errorf("migrations: applying: %w", err)
	}

	after, err := appliedIDs(db)
	if err != nil {
		return err
	}
	log.Info("migrations up to date",
		zap.Int("applied_now", len(after)-len(applied)),
		zap.Int("total_applied", len(after)),
	)
	return nil
}

// Down rolls back the most recently applied migration.
func Down(db *gorm.DB, log *zap.Logger) error {
	log.Warn("rolling back the last migration")
	if err := New(db).RollbackLast(); err != nil {
		return fmt.Errorf("migrations: rolling back: %w", err)
	}
	log.Info("rollback complete")
	return nil
}

// Status reports which migrations are applied.
func Status(db *gorm.DB) ([]StepStatus, error) {
	applied, err := appliedIDs(db)
	if err != nil {
		return nil, err
	}

	out := make([]StepStatus, 0, len(All()))
	for _, m := range All() {
		_, ok := applied[m.ID]
		out = append(out, StepStatus{ID: m.ID, Applied: ok})
	}
	return out, nil
}

type StepStatus struct {
	ID      string `json:"id"`
	Applied bool   `json:"applied"`
}

func appliedIDs(db *gorm.DB) (map[string]struct{}, error) {
	out := map[string]struct{}{}

	if !db.Migrator().HasTable(options().TableName) {
		return out, nil
	}

	var ids []string
	if err := db.Table(options().TableName).Pluck(options().IDColumnName, &ids).Error; err != nil {
		return nil, fmt.Errorf("migrations: reading applied migrations: %w", err)
	}
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out, nil
}
