package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
)

type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Role      *string `json:"role,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	Password  *string `json:"password,omitempty"`
}

type CreateRoleRequest struct {
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          *string     `json:"name,omitempty"`
	Description   *string     `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type RBACService interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*model.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error

	ListRoles(ctx context.Context) ([]model.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*model.Role, error)
	CreateRole(ctx context.Context, req CreateRoleRequest) (*model.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req UpdateRoleRequest) (*model.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ListPermissions(ctx context.Context) ([]model.Permission, error)
}

type rbacService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewRBACService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) RBACService {
	return &rbacService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *rbacService) CreateUser(ctx context.Context, req CreateUserRequest) (*model.User, error) {
	email := strings.TrimSpace(req.Email)
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	role := strings.TrimSpace(req.Role)

	fields := apperr.Fields{}
	if email == "" {
		fields["email"] = []string{"email is required"}
	}
	if req.Password == "" || len(req.Password) < 6 {
		fields["password"] = []string{"password must be at least 6 characters"}
	}
	if firstName == "" {
		fields["first_name"] = []string{"first_name is required"}
	}
	if lastName == "" {
		fields["last_name"] = []string{"last_name is required"}
	}
	if role == "" {
		role = model.RoleUser
	}
	if len(fields) > 0 {
		return nil, apperr.Validation(fields)
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		return nil, apperr.Conflict("Email %q is already registered", email)
	}

	user := &model.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		IsActive:  req.IsActive,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, apperr.Internal(err)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperr.Internal(err)
	}

	return user, nil
}

func (s *rbacService) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil {
		return nil, apperr.NotFound("user")
	}
	return user, nil
}

func (s *rbacService) UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil {
		return nil, apperr.NotFound("user")
	}

	if req.FirstName != nil {
		fn := strings.TrimSpace(*req.FirstName)
		if fn == "" {
			return nil, apperr.Validation(apperr.Fields{"first_name": []string{"first_name cannot be empty"}})
		}
		user.FirstName = fn
	}
	if req.LastName != nil {
		ln := strings.TrimSpace(*req.LastName)
		if ln == "" {
			return nil, apperr.Validation(apperr.Fields{"last_name": []string{"last_name cannot be empty"}})
		}
		user.LastName = ln
	}
	if req.Role != nil {
		user.Role = strings.TrimSpace(*req.Role)
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		if len(*req.Password) < 6 {
			return nil, apperr.Validation(apperr.Fields{"password": []string{"password must be at least 6 characters"}})
		}
		if err := user.SetPassword(*req.Password); err != nil {
			return nil, apperr.Internal(err)
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperr.Internal(err)
	}
	return user, nil
}

func (s *rbacService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return apperr.Internal(err)
	}
	if user == nil {
		return apperr.NotFound("user")
	}
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (s *rbacService) ListRoles(ctx context.Context) ([]model.Role, error) {
	roles, err := s.roleRepo.ListRoles(ctx)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return roles, nil
}

func (s *rbacService) GetRole(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	role, err := s.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if role == nil {
		return nil, apperr.NotFound("role")
	}
	return role, nil
}

func (s *rbacService) CreateRole(ctx context.Context, req CreateRoleRequest) (*model.Role, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperr.Validation(apperr.Fields{"name": []string{"role name is required"}})
	}

	existing, err := s.roleRepo.GetRoleByName(ctx, name)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		return nil, apperr.Conflict("Role %q already exists", name)
	}

	role := &model.Role{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		IsSystem:    false,
	}

	if err := s.roleRepo.CreateRole(ctx, role, req.PermissionIDs); err != nil {
		return nil, apperr.Internal(err)
	}

	return s.roleRepo.GetRoleByID(ctx, role.ID)
}

func (s *rbacService) UpdateRole(ctx context.Context, id uuid.UUID, req UpdateRoleRequest) (*model.Role, error) {
	role, err := s.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if role == nil {
		return nil, apperr.NotFound("role")
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, apperr.Validation(apperr.Fields{"name": []string{"role name cannot be empty"}})
		}
		role.Name = name
	}
	if req.Description != nil {
		role.Description = strings.TrimSpace(*req.Description)
	}

	if err := s.roleRepo.UpdateRole(ctx, role, req.PermissionIDs); err != nil {
		return nil, apperr.Internal(err)
	}

	return s.roleRepo.GetRoleByID(ctx, role.ID)
}

func (s *rbacService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	role, err := s.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		return apperr.Internal(err)
	}
	if role == nil {
		return apperr.NotFound("role")
	}
	if role.IsSystem {
		return apperr.Forbidden("System roles cannot be deleted")
	}
	if role.UserCount > 0 {
		return apperr.Conflict("Cannot delete role assigned to %d active users", role.UserCount)
	}

	if err := s.roleRepo.DeleteRole(ctx, id); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (s *rbacService) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	perms, err := s.roleRepo.ListPermissions(ctx)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return perms, nil
}
