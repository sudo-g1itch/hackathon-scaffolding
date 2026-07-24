package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func m0003CreateRolesPermissions() *gormigrate.Migration {
	type Permission struct {
		ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
		CreatedAt   time.Time      `gorm:"not null"`
		UpdatedAt   time.Time      `gorm:"not null"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		Slug        string         `gorm:"type:varchar(100);uniqueIndex;not null"`
		Name        string         `gorm:"type:varchar(100);not null"`
		Module      string         `gorm:"type:varchar(50);not null;index"`
		Description string         `gorm:"type:varchar(255)"`
	}

	type Role struct {
		ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
		CreatedAt   time.Time      `gorm:"not null"`
		UpdatedAt   time.Time      `gorm:"not null"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		Name        string         `gorm:"type:varchar(100);uniqueIndex;not null"`
		Description string         `gorm:"type:varchar(255)"`
		IsSystem    bool           `gorm:"not null;default:false"`
	}

	type RolePermission struct {
		RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
		PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
	}

	return &gormigrate.Migration{
		ID: "0003_create_roles_permissions",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&Permission{}, &Role{}, &RolePermission{}); err != nil {
				return err
			}

			// Seed standard permissions
			seedPermissions := []Permission{
				{Slug: "users:read", Name: "View Users", Module: "User Management", Description: "View user list and user profiles"},
				{Slug: "users:write", Name: "Create/Edit Users", Module: "User Management", Description: "Create, update, and manage user accounts"},
				{Slug: "users:delete", Name: "Delete Users", Module: "User Management", Description: "Deactivate or delete user accounts"},

				{Slug: "roles:read", Name: "View Roles", Module: "Roles & Permissions", Description: "View system and custom roles"},
				{Slug: "roles:write", Name: "Manage Roles", Module: "Roles & Permissions", Description: "Create and modify roles and permission matrices"},
				{Slug: "roles:delete", Name: "Delete Roles", Module: "Roles & Permissions", Description: "Delete custom roles"},

				{Slug: "permissions:read", Name: "View Permissions", Module: "Roles & Permissions", Description: "View granular system permissions"},

				{Slug: "analytics:read", Name: "View Analytics", Module: "Analytics & Reports", Description: "View system analytics and operational reports"},

				{Slug: "settings:read", Name: "View Settings", Module: "System Settings", Description: "View application system settings"},
				{Slug: "settings:write", Name: "Modify Settings", Module: "System Settings", Description: "Update global system configurations"},
			}

			permMap := make(map[string]uuid.UUID)
			now := time.Now().UTC()

			for i := range seedPermissions {
				id, _ := uuid.NewV7()
				seedPermissions[i].ID = id
				seedPermissions[i].CreatedAt = now
				seedPermissions[i].UpdatedAt = now

				var existing Permission
				if err := tx.Where("slug = ?", seedPermissions[i].Slug).First(&existing).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						if err := tx.Create(&seedPermissions[i]).Error; err != nil {
							return err
						}
						permMap[seedPermissions[i].Slug] = id
					}
				} else {
					permMap[seedPermissions[i].Slug] = existing.ID
				}
			}

			// Seed default system roles
			seedRoles := []Role{
				{Name: "admin", Description: "Full administrative access to all features and settings", IsSystem: true},
				{Name: "manager", Description: "Management access for users, reports, and team operations", IsSystem: true},
				{Name: "user", Description: "Standard user access with basic operational permissions", IsSystem: true},
			}

			roleMap := make(map[string]uuid.UUID)

			for i := range seedRoles {
				var existing Role
				if err := tx.Where("name = ?", seedRoles[i].Name).First(&existing).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						id, _ := uuid.NewV7()
						seedRoles[i].ID = id
						seedRoles[i].CreatedAt = now
						seedRoles[i].UpdatedAt = now
						if err := tx.Create(&seedRoles[i]).Error; err != nil {
							return err
						}
						roleMap[seedRoles[i].Name] = id
					}
				} else {
					roleMap[seedRoles[i].Name] = existing.ID
				}
			}

			// Assign all permissions to Admin role
			if adminID, ok := roleMap["admin"]; ok {
				for _, permID := range permMap {
					var rp RolePermission
					if err := tx.Where("role_id = ? AND permission_id = ?", adminID, permID).First(&rp).Error; err == gorm.ErrRecordNotFound {
						tx.Create(&RolePermission{RoleID: adminID, PermissionID: permID})
					}
				}
			}

			// Assign read/write permissions to Manager role
			if mgrID, ok := roleMap["manager"]; ok {
				managerPerms := []string{"users:read", "users:write", "roles:read", "analytics:read", "settings:read"}
				for _, slug := range managerPerms {
					if permID, ok := permMap[slug]; ok {
						var rp RolePermission
						if err := tx.Where("role_id = ? AND permission_id = ?", mgrID, permID).First(&rp).Error; err == gorm.ErrRecordNotFound {
							tx.Create(&RolePermission{RoleID: mgrID, PermissionID: permID})
						}
					}
				}
			}

			// Assign basic permissions to User role
			if userID, ok := roleMap["user"]; ok {
				userPerms := []string{"analytics:read", "settings:read"}
				for _, slug := range userPerms {
					if permID, ok := permMap[slug]; ok {
						var rp RolePermission
						if err := tx.Where("role_id = ? AND permission_id = ?", userID, permID).First(&rp).Error; err == gorm.ErrRecordNotFound {
							tx.Create(&RolePermission{RoleID: userID, PermissionID: permID})
						}
					}
				}
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&RolePermission{}, &Role{}, &Permission{})
		},
	}
}
