// Command api is the Hackathon HTTP server.
//
// This file is the composition root: the one place that constructs concrete
// dependencies and injects them downward
// (config → logger → database → repositories → services → handlers → router).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/database"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/handler"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/logging"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/server"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/service"
	"github.com/sudo-g1itch/hackathon-scaffolding/migrations"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	envFile := flag.String("env", ".env", "path to an optional .env file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(*envFile)
	if err != nil {
		return err
	}

	log, err := logging.New(cfg.Log, cfg.App)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting hackathon api", zap.String("version", version))

	db, err := database.Open(cfg.Database, log)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Error("closing database", zap.Error(err))
		}
	}()

	// Development convenience: auto-migrate on boot.
	if cfg.App.AutoMigrate {
		if err := migrations.Up(db, log); err != nil {
			return err
		}
	}

	// --- dependency graph, bottom up ---
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	recoveraiRepo := repository.NewRecoverAIRepository(db)
	goalRepo := repository.NewGoalRepository(db)
	supportRepo := repository.NewSupportRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg.JWT)
	rbacSvc := service.NewRBACService(userRepo, roleRepo)
	aiSvc := service.NewAIService(cfg.AI, log)
	voiceSvc := service.NewVoiceService(cfg.AI, log)

	// The recovery plan and the caregiver conversation are built before the
	// dashboard that summarises them, and before the caregiver views that
	// compose both.
	goalSvc := service.NewGoalService(goalRepo, recoveraiRepo, userRepo, log)
	supportSvc := service.NewSupportService(supportRepo, recoveraiRepo, userRepo, log)
	careSvc := service.NewCareService(recoveraiRepo, goalSvc, supportSvc, log)
	recoveraiSvc := service.NewRecoverAIService(
		recoveraiRepo, userRepo, goalSvc, supportSvc, aiSvc, voiceSvc, log,
	)

	// Both AI integrations are optional. Say plainly at boot which ones are
	// inert, so a demo failure is never a mystery.
	log.Info("ai integrations",
		zap.Bool("gemini_enabled", aiSvc.Available()),
		zap.String("gemini_model", cfg.AI.GeminiModel),
		zap.Bool("deepgram_enabled", voiceSvc.Available()),
	)
	if !aiSvc.Available() {
		log.Warn("AI_GEMINI_API_KEY is unset — check-in analysis, coach, emergency and education are disabled")
	}
	if !voiceSvc.Available() {
		log.Warn("AI_DEEPGRAM_API_KEY is unset — voice transcription and playback are disabled")
	}

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(authSvc, rbacSvc)
	roleHandler := handler.NewRoleHandler(rbacSvc)
	recoveraiHandler := handler.NewRecoverAIHandler(recoveraiSvc, cfg.AI.MaxAudioBytes)
	goalHandler := handler.NewGoalHandler(goalSvc)
	careHandler := handler.NewCareHandler(careSvc, supportSvc)

	handlers := server.Handlers{
		Health:    handler.NewHealthHandler(db, version),
		Auth:      authHandler,
		User:      userHandler,
		Role:      roleHandler,
		RecoverAI: recoveraiHandler,
		Goal:      goalHandler,
		Care:      careHandler,
		AuthSvc:   authSvc,
	}

	srv, err := server.New(cfg, log, handlers)
	if err != nil {
		return err
	}

	return srv.Run(ctx)
}
