package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/ctxkey"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

type AuthHandler struct {
	authSvc service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if !response.BindJSON(c, &req) {
		return
	}

	res, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, res)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if !response.BindJSON(c, &req) {
		return
	}

	res, err := h.authSvc.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, res)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("User not authenticated"))
		return
	}

	user, err := h.authSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, user)
}
