package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func m0002CreateUsersTable() *gormigrate.Migration {
	type User struct {
		ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
		CreatedAt    time.Time      `gorm:"not null"`
		UpdatedAt    time.Time      `gorm:"not null"`
		DeletedAt    gorm.DeletedAt `gorm:"index"`
		Email        string         `gorm:"type:citext;uniqueIndex;not null"`
		PasswordHash string         `gorm:"not null"`
		FirstName    string         `gorm:"not null"`
		LastName     string         `gorm:"not null"`
		Role         string         `gorm:"type:varchar(50);not null;default:'user';index"`
		IsActive     bool           `gorm:"not null;default:true"`
		AvatarURL    *string        `gorm:"type:varchar(512)"`
		LastLoginAt  *time.Time
	}

	return &gormigrate.Migration{
		ID: "0002_create_users_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&User{}); err != nil {
				return err
			}

			// Seed default admin user if not exists
			var count int64
			if err := tx.Model(&User{}).Where("email = ?", "admin@hackathon.local").Count(&count).Error; err != nil {
				return err
			}

			if count == 0 {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
				if err != nil {
					return err
				}

				id, _ := uuid.NewV7()
				admin := User{
					ID:           id,
					CreatedAt:    time.Now().UTC(),
					UpdatedAt:    time.Now().UTC(),
					Email:        "admin@hackathon.local",
					PasswordHash: string(hashedPassword),
					FirstName:    "Admin",
					LastName:     "User",
					Role:         "admin",
					IsActive:     true,
				}

				if err := tx.Create(&admin).Error; err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&User{})
		},
	}
}
