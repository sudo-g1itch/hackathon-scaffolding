package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
)

func m0004CreateRecoveraiTables() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "0004_create_recoverai_tables",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(
				&model.RecoveryProfile{},
				&model.Checkin{},
				&model.CoachMessage{},
				&model.EmergencyLog{},
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(
				&model.RecoveryProfile{},
				&model.Checkin{},
				&model.CoachMessage{},
				&model.EmergencyLog{},
			)
		},
	}
}
