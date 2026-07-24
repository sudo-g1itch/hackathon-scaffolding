// Package ctxkey owns every value the middleware chain puts on the request
// context, and the only accessors for reading them back.
package ctxkey

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	keyRequestID = "ctx.request_id"
	keyLogger    = "ctx.logger"
	keyUserID    = "ctx.user_id"
	keyUserRole  = "ctx.user_role"
)

func SetRequestID(c *gin.Context, id string) { c.Set(keyRequestID, id) }

func RequestID(c *gin.Context) string {
	v, ok := c.Get(keyRequestID)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func SetLogger(c *gin.Context, l *zap.Logger) { c.Set(keyLogger, l) }

func Logger(c *gin.Context) *zap.Logger {
	v, ok := c.Get(keyLogger)
	if !ok {
		return zap.NewNop()
	}
	l, ok := v.(*zap.Logger)
	if !ok || l == nil {
		return zap.NewNop()
	}
	return l
}

func SetUserID(c *gin.Context, id uuid.UUID) { c.Set(keyUserID, id) }

func UserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(keyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok && id != uuid.Nil
}

func SetUserRole(c *gin.Context, role string) { c.Set(keyUserRole, role) }

func UserRole(c *gin.Context) string {
	v, ok := c.Get(keyUserRole)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

