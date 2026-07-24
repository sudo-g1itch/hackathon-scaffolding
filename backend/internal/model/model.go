// Package model defines the GORM structs for the database schema.
//
// Every model embeds BaseModel for a UUIDv7 PK + timestamps + soft delete.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is embedded in every entity for consistent PK and timestamp handling.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate generates a UUIDv7 if the ID has not been set.
func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		b.ID = id
	}
	return nil
}

// TODO: Add your domain models here. Example:
//
// type User struct {
//     BaseModel
//     Email     string `gorm:"uniqueIndex;not null" json:"email"`
//     FirstName string `gorm:"not null" json:"first_name"`
//     LastName  string `gorm:"not null" json:"last_name"`
// }
