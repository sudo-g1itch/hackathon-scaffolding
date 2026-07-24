// Package handler holds the HTTP handlers.
//
// Handlers are thin: bind + validate input, call a service, write the
// response envelope. No business logic, no direct DB access.
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/database"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/response"
	"gorm.io/gorm"
)

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
