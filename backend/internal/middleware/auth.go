package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/ctxkey"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

func Authenticate(authSvc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, apperr.Unauthorized("Authorization header missing"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, apperr.Unauthorized("Authorization header must be Bearer <token>"))
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := authSvc.ValidateToken(tokenStr)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		ctxkey.SetUserID(c, claims.UserID)
		ctxkey.SetUserRole(c, claims.Role)

		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleMap := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleMap[r] = true
	}

	return func(c *gin.Context) {
		userRole := ctxkey.UserRole(c)
		if userRole == "" || !roleMap[userRole] {
			response.Error(c, apperr.Forbidden("Access denied: insufficient permissions"))
			c.Abort()
			return
		}

		c.Next()
	}
}
