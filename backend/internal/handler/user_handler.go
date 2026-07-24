package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/pagination"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

type UserHandler struct {
	authSvc service.AuthService
}

func NewUserHandler(authSvc service.AuthService) *UserHandler {
	return &UserHandler{authSvc: authSvc}
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
