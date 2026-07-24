package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
)

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]model.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	GetRoleByName(ctx context.Context, name string) (*model.Role, error)
	CreateRole(ctx context.Context, role *model.Role, permissionIDs []uuid.UUID) error
	UpdateRole(ctx context.Context, role *model.Role, permissionIDs []uuid.UUID) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ListPermissions(ctx context.Context) ([]model.Permission, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) ListRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	if err != nil {
		return nil, err
	}

	for i := range roles {
		var count int64
		_ = r.db.WithContext(ctx).Model(&model.User{}).Where("role = ?", roles[i].Name).Count(&count).Error
		roles[i].UserCount = int(count)
	}

	return roles, nil
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var count int64
	_ = r.db.WithContext(ctx).Model(&model.User{}).Where("role = ?", role.Name).Count(&count).Error
	role.UserCount = int(count)

	return &role, nil
}

func (r *roleRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "LOWER(name) = LOWER(?)", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) CreateRole(ctx context.Context, role *model.Role, permissionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}

		if len(permissionIDs) > 0 {
			for _, pid := range permissionIDs {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pid}
				if err := tx.Create(&rp).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *roleRepository) UpdateRole(ctx context.Context, role *model.Role, permissionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(role).Error; err != nil {
			return err
		}

		if err := tx.Where("role_id = ?", role.ID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}

		if len(permissionIDs) > 0 {
			for _, pid := range permissionIDs {
				rp := model.RolePermission{RoleID: role.ID, PermissionID: pid}
				if err := tx.Create(&rp).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *roleRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Role{}, "id = ?", id).Error
	})
}

func (r *roleRepository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.WithContext(ctx).Order("module ASC, name ASC").Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}
