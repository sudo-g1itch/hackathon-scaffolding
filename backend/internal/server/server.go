// Package server builds the Gin engine, installs middleware, registers routes,
// and runs the HTTP server with graceful shutdown.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/handler"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/middleware"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
)

const APIVersion = "v1"

// Handlers collects everything the router needs.
type Handlers struct {
	Health    *handler.HealthHandler
	Auth      *handler.AuthHandler
	User      *handler.UserHandler
	Role      *handler.RoleHandler
	RecoverAI *handler.RecoverAIHandler
	Goal      *handler.GoalHandler
	Care      *handler.CareHandler
	AuthSvc   service.AuthService
}

// Server owns the HTTP listener and its lifecycle.
type Server struct {
	http *http.Server
	log  *zap.Logger
	cfg  config.HTTP
}

// New builds the engine and wraps it in an http.Server.
func New(cfg *config.Config, log *zap.Logger, h Handlers) (*Server, error) {
	engine, err := newEngine(cfg, log, h)
	if err != nil {
		return nil, err
	}

	return &Server{
		http: &http.Server{
			Addr:              cfg.HTTP.Addr(),
			Handler:           engine,
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
			IdleTimeout:       cfg.HTTP.IdleTimeout,
			ReadHeaderTimeout: headerTimeout,
		},
		log: log,
		cfg: cfg.HTTP,
	}, nil
}

const headerTimeout = 5 * time.Second

func (s *Server) Handler() http.Handler { return s.http.Handler }

func newEngine(cfg *config.Config, log *zap.Logger, h Handlers) (*gin.Engine, error) {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	if err := engine.SetTrustedProxies(cfg.HTTP.TrustedProxies); err != nil {
		return nil, fmt.Errorf("server: configuring trusted proxies: %w", err)
	}

	engine.RedirectTrailingSlash = false
	engine.HandleMethodNotAllowed = true

	engine.Use(
		middleware.Recovery(log),
		middleware.RequestID(),
		middleware.CORS(cfg.CORS),
		middleware.Logger(log),
	)

	engine.NoRoute(middleware.NotFound())
	engine.NoMethod(middleware.MethodNotAllowed())

	registerRoutes(engine, h)
	return engine, nil
}

// registerRoutes is the single source of truth for the API's URL surface.
func registerRoutes(engine *gin.Engine, h Handlers) {
	// Public probes, unversioned.
	engine.GET("/healthz", h.Health.Live)
	engine.GET("/readyz", h.Health.Ready)

	// Versioned API group.
	api := engine.Group("/api/" + APIVersion)

	// Public Auth routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", h.Auth.Login)
		authGroup.POST("/register", h.Auth.Register)
	}

	// Protected routes (Require Authentication)
	protected := api.Group("")
	protected.Use(middleware.Authenticate(h.AuthSvc))
	{
		protected.GET("/auth/me", h.Auth.Me)

		// AnchorOne recovery routes.
		recoverai := protected.Group("")
		{
			// Which optional integrations (Gemini, Deepgram) are configured.
			recoverai.GET("/capabilities", h.RecoverAI.Capabilities)

			// Check-in: voice (audio upload) or typed.
			recoverai.POST("/checkin", h.RecoverAI.Checkin)
			recoverai.POST("/risk", h.RecoverAI.Risk)

			// Raw voice services.
			recoverai.POST("/voice/transcribe", h.RecoverAI.Transcribe)
			recoverai.POST("/voice/speak", h.RecoverAI.Speak)

			recoverai.GET("/dashboard", h.RecoverAI.Dashboard)
			recoverai.GET("/timeline", h.RecoverAI.Timeline)
			recoverai.POST("/emergency", h.RecoverAI.Emergency)

			recoverai.POST("/coach/chat", h.RecoverAI.CoachChat)
			recoverai.GET("/coach/history", h.RecoverAI.CoachHistory)
			recoverai.POST("/education", h.RecoverAI.Education)

			// Recovery profile (goal, substance, emergency contacts).
			recoverai.GET("/profile", h.RecoverAI.GetProfile)
			recoverai.PUT("/profile", h.RecoverAI.UpdateProfile)
			recoverai.GET("/caregivers", h.RecoverAI.GetCaregivers)
			recoverai.PUT("/profile/caregiver", h.RecoverAI.SetCaregiver)
		}

		// The recovery plan: many goals per person, each with a progress log.
		// These paths act on the caller's OWN plan.
		goals := protected.Group("/goals")
		{
			goals.GET("", h.Goal.List)
			goals.POST("", h.Goal.Create)
			goals.GET("/summary", h.Goal.Summary)
			goals.GET("/:goalID", h.Goal.Get)
			goals.PUT("/:goalID", h.Goal.Update)
			goals.DELETE("/:goalID", h.Goal.Delete)
			goals.POST("/:goalID/progress", h.Goal.LogProgress)
		}

		// Patient-scoped routes. There is no role guard here on purpose: who may
		// read a given person's record depends on whether that person linked
		// this caregiver, which only the service layer can answer. A role check
		// would either be too loose (any caregiver sees anyone) or redundant.
		patients := protected.Group("/patients/:patientID")
		{
			patients.GET("", h.Care.PatientOverview)
			patients.GET("/goals", h.Goal.ListForPatient)
			patients.POST("/goals", h.Goal.CreateForPatient)

			// The patient <-> caregiver conversation, shared by both sides.
			patients.GET("/messages", h.Care.Thread)
			patients.POST("/messages", h.Care.SendMessage)
			patients.POST("/messages/read", h.Care.MarkRead)
		}

		// Unread badge, for whichever side of a conversation the caller is on.
		protected.GET("/messages/unread", h.Care.UnreadCount)

		// The caregiver's list of the people who chose them. Unlike the routes
		// above this one is not about a specific person, so a role guard is the
		// right check: a plain user has no list to show.
		caregiverOnly := protected.Group("")
		caregiverOnly.Use(middleware.RequireRole(model.RoleCaregiver, model.RoleAdmin))
		{
			caregiverOnly.GET("/caregiver", h.Care.ListPatients)
		}

		// Admin-only routes (RBAC check)
		adminOnly := protected.Group("")
		adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
		{
			// Users CRUD
			adminOnly.GET("/users", h.User.List)
			adminOnly.POST("/users", h.User.Create)
			adminOnly.GET("/users/:id", h.User.GetByID)
			adminOnly.PUT("/users/:id", h.User.Update)
			adminOnly.DELETE("/users/:id", h.User.Delete)

			// Roles & Permissions
			adminOnly.GET("/roles", h.Role.ListRoles)
			adminOnly.POST("/roles", h.Role.CreateRole)
			adminOnly.GET("/roles/:id", h.Role.GetRole)
			adminOnly.PUT("/roles/:id", h.Role.UpdateRole)
			adminOnly.DELETE("/roles/:id", h.Role.DeleteRole)
			adminOnly.GET("/permissions", h.Role.ListPermissions)
		}
	}
}

// Run starts the server and blocks until ctx is cancelled, then drains.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server listening", zap.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server: listening on %s: %w", s.http.Addr, err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.shutdown()
	}
}

func (s *Server) shutdown() error {
	s.log.Info("shutting down http server", zap.Duration("timeout", s.cfg.ShutdownTimeout))

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("server: graceful shutdown failed: %w", err)
	}
	s.log.Info("http server stopped cleanly")
	return nil
}
