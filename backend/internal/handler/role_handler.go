package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

type RoleHandler struct {
	rbacSvc service.RBACService
}

func NewRoleHandler(rbacSvc service.RBACService) *RoleHandler {
	return &RoleHandler{rbacSvc: rbacSvc}
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.rbacSvc.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, roles)
}

func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.Validation(apperr.Fields{"id": []string{"invalid role ID"}}))
		return
	}

	role, err := h.rbacSvc.GetRole(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, role)
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req service.CreateRoleRequest
	if !response.BindJSON(c, &req) {
		return
	}

	role, err := h.rbacSvc.CreateRole(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, role)
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.Validation(apperr.Fields{"id": []string{"invalid role ID"}}))
		return
	}

	var req service.UpdateRoleRequest
	if !response.BindJSON(c, &req) {
		return
	}

	role, err := h.rbacSvc.UpdateRole(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, role)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.Validation(apperr.Fields{"id": []string{"invalid role ID"}}))
		return
	}

	if err := h.rbacSvc.DeleteRole(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, gin.H{"deleted": true})
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.rbacSvc.ListPermissions(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, perms)
}
