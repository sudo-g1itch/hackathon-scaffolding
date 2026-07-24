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
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/server"
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
	// TODO: Wire your repositories → services → handlers here.

	handlers := server.Handlers{
		Health: handler.NewHealthHandler(db, version),
	}

	srv, err := server.New(cfg, log, handlers)
	if err != nil {
		return err
	}

	return srv.Run(ctx)
}
