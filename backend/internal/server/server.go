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
)

const APIVersion = "v1"

// Handlers collects everything the router needs.
type Handlers struct {
	Health *handler.HealthHandler
	// TODO: Add your domain handlers here as the project grows.
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
	_ = engine.Group("/api/" + APIVersion)

	// TODO: Register your domain routes here. Example:
	// api.POST("/users", h.User.Create)
	// api.GET("/users", h.User.List)
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
