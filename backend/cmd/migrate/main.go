// Command migrate applies, rolls back, and reports on database migrations.
//
// Usage:
//
//	migrate up        # apply every pending migration (default)
//	migrate down      # roll back the most recently applied migration
//	migrate status    # list every migration and whether it has been applied
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/database"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/logging"
	"github.com/sudo-g1itch/hackathon-scaffolding/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	envFile := flag.String("env", ".env", "path to an optional .env file")
	flag.Parse()

	command := flag.Arg(0)
	if command == "" {
		command = "up"
	}

	cfg, err := config.Load(*envFile)
	if err != nil {
		return err
	}

	log, err := logging.New(cfg.Log, cfg.App)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	db, err := database.Open(cfg.Database, log)
	if err != nil {
		return err
	}
	defer func() { _ = database.Close(db) }()

	switch command {
	case "up":
		return migrations.Up(db, log)
	case "down":
		return migrations.Down(db, log)
	case "status":
		steps, err := migrations.Status(db)
		if err != nil {
			return err
		}
		for _, s := range steps {
			mark := "pending"
			if s.Applied {
				mark = "applied"
			}
			fmt.Printf("%-10s %s\n", mark, s.ID)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q: expected up, down, or status", command)
	}
}
