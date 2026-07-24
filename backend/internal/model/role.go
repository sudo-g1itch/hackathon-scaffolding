package model

type Role struct {
	BaseModel
	Name        string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	IsSystem    bool         `gorm:"not null;default:false" json:"is_system"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	UserCount   int          `gorm:"-" json:"user_count"`
}
