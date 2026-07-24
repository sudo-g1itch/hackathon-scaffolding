package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/pagination"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

type UserHandler struct {
	authSvc service.AuthService
	rbacSvc service.RBACService
}

func NewUserHandler(authSvc service.AuthService, rbacSvc service.RBACService) *UserHandler {
	return &UserHandler{
		authSvc: authSvc,
		rbacSvc: rbacSvc,
	}
}

func (h *UserHandler) List(c *gin.Context) {
	var raw pagination.Raw
	if !response.BindQuery(c, &raw) {
		return
	}

	params, err := pagination.Resolve(raw, repository.UserSortable, "created_at")
	if err != nil {
		response.Error(c, err)
		return
	}

	users, total, err := h.authSvc.ListUsers(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, pagination.NewPage(users, params, total))
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.Validation(apperr.Fields{"id": []string{"invalid user ID"}}))
		return
	}

	user, err := h.rbacSvc.GetUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if !response.BindJSON(c, &req) {
		return
	}

	user, err := h.rbacSvc.CreateUser(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.Validation(apperr.Fields{"id": []string{"invalid user ID"}}))
		return
	}

	var req service.UpdateUserRequest
	if !response.BindJSON(c, &req) {
		return
	}

	user, err := h.rbacSvc.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.Validation(apperr.Fields{"id": []string{"invalid user ID"}}))
		return
	}

	if err := h.rbacSvc.DeleteUser(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, gin.H{"deleted": true})
}
