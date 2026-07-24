package model

import "github.com/google/uuid"

type Permission struct {
	BaseModel
	Slug        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Module      string `gorm:"type:varchar(50);not null;index" json:"module"`
	Description string `gorm:"type:varchar(255)" json:"description"`
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permission_id"`
}
