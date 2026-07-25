// Package handler holds the HTTP handlers.
//
// Handlers are thin: bind + validate input, call a service, write the
// response envelope. No business logic, no direct DB access.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/ctxkey"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/database"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

// actorFrom builds the acting user from the request context, writing the 401
// itself so every caller is a two-line guard rather than a repeated block.
//
// Handlers pass the whole Actor down, never just an id: authorisation for the
// patient-scoped endpoints depends on the caller's role as well, and that
// decision belongs to the service layer.
func actorFrom(c *gin.Context) (service.Actor, bool) {
	userID, ok := ctxkey.UserID(c)
	if !ok {
		response.Error(c, apperr.Unauthorized("unauthorized"))
		return service.Actor{}, false
	}

	return service.Actor{ID: userID, Role: ctxkey.UserRole(c)}, true
}

// uuidParam parses a path parameter as a UUID, answering 422 with the field
// name when it is malformed.
func uuidParam(c *gin.Context, name string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.Error(c, apperr.Validation(apperr.Fields{name: {"must be a valid UUID"}}))
		return uuid.Nil, false
	}
	return parsed, true
}

// HealthHandler serves the liveness and readiness probes.
type HealthHandler struct {
	db      *gorm.DB
	version string
}

func NewHealthHandler(db *gorm.DB, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

// Live is GET /healthz — always returns 200.
func (h *HealthHandler) Live(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "ok",
		"version": h.version,
	})
}

// Ready is GET /readyz — verifies the database is reachable.
func (h *HealthHandler) Ready(c *gin.Context) {
	if err := database.Ping(c.Request.Context(), h.db); err != nil {
		response.Error(c, err)
		return
	}
	stats, _ := database.Stats(h.db)
	response.OK(c, gin.H{
		"status":   "ok",
		"database": stats,
	})
}

// TODO: Add your domain handlers here. Each handler receives a service
// through its constructor — never access the DB directly from a handler.
