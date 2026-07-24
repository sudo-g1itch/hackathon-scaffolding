# Hackathon Project — common commands.
#
# `make help` lists every target.

.DEFAULT_GOAL := help
.PHONY: help dev api web check build vet lint fmt tidy \
        migrate-up migrate-down migrate-status db-shell db-reset \
        up down logs ps clean

BACKEND  := backend
FRONTEND := frontend
COMPOSE  := docker compose
VERSION  := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

## help: list the available targets
help:
	@echo "Hackathon Project — make targets\n"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | sort
	@echo ""

# ---------------------------------------------------------------- development

## dev: start Postgres in Docker, then run the API locally with hot reload
dev:
	$(COMPOSE) up -d postgres
	cd $(BACKEND) && go run ./cmd/api

## api: run the API against an already-running database
api:
	cd $(BACKEND) && go run ./cmd/api

## web: run the Next.js dev server
web:
	cd $(FRONTEND) && npm run dev

# -------------------------------------------------------------------- quality

## check: everything CI should run — build, vet, lint
check: build vet lint

## build: compile the backend
build:
	cd $(BACKEND) && go build ./...

## vet: run go vet over the backend
vet:
	cd $(BACKEND) && go vet ./...

## lint: lint the frontend (strict TypeScript + ESLint)
lint:
	cd $(FRONTEND) && npm run lint

## fmt: format Go and frontend sources
fmt:
	cd $(BACKEND) && gofmt -w .
	cd $(FRONTEND) && npm run format

## tidy: prune and verify go.mod/go.sum
tidy:
	cd $(BACKEND) && go mod tidy

# ------------------------------------------------------------------- database

## migrate-up: apply all pending migrations
migrate-up:
	cd $(BACKEND) && go run ./cmd/migrate up

## migrate-down: roll back the most recent migration
migrate-down:
	cd $(BACKEND) && go run ./cmd/migrate down

## migrate-status: show which migrations have been applied
migrate-status:
	cd $(BACKEND) && go run ./cmd/migrate status

## db-shell: open psql against the compose database
db-shell:
	$(COMPOSE) exec postgres psql -U $${DATABASE_USER:-hackathon} -d $${DATABASE_NAME:-hackathon}

## db-reset: DESTROY the database volume and recreate from migrations
db-reset:
	$(COMPOSE) down -v
	$(COMPOSE) up -d postgres
	@sleep 3
	cd $(BACKEND) && go run ./cmd/migrate up

# --------------------------------------------------------------------- docker

## up: build and start the whole stack in Docker
up:
	VERSION=$(VERSION) $(COMPOSE) up -d --build

## down: stop the stack (the database volume survives)
down:
	$(COMPOSE) down

## logs: follow the logs of every service
logs:
	$(COMPOSE) logs -f

## ps: show the status of each service
ps:
	$(COMPOSE) ps

# ---------------------------------------------------------------------- other

## clean: remove build output and coverage files
clean:
	rm -rf $(BACKEND)/bin $(BACKEND)/coverage.out $(FRONTEND)/.next
