# GEMINI.md — Hackathon Project

Operating guide for AI assistants (Gemini, Claude Code, Cursor, etc.) and humans
working in this repository. **Read this before any non-trivial change.**

---

## 1. What this is

A **hackathon project** with a layered full-stack architecture:

- **Backend:** a single **Go monolith** — **Gin** (HTTP) + **GORM** (ORM) on
  **PostgreSQL**. Not microservices.
- **Frontend:** **Next.js 16 / React 19 / MUI 7** (Materialize admin template —
  the starter kit).
- **Infrastructure:** **Docker Compose** (Postgres + API + web), **Makefile**
  for all commands.
- **Remote:** `github.com/sudo-g1itch/hackathon-scaffolding` — **main branch only**.

---

## 2. Repository map

```
hackathon/
├── GEMINI.md              ← you are here — THE RULES
├── Makefile               ← all commands: make help
├── docker-compose.yml     ← full stack
├── docker-compose.override.yml ← dev overrides
├── .env.example           ← stack config
├── .gitignore
├── backend/               ← Go monolith (Gin + GORM + Postgres)
│   ├── cmd/api/           ← API entrypoint (composition root)
│   ├── cmd/migrate/       ← migration CLI
│   ├── internal/
│   │   ├── apperr/        ← error codes & constructors
│   │   ├── config/        ← Viper-based config (single source of env)
│   │   ├── ctxkey/        ← typed context keys
│   │   ├── database/      ← GORM connection + pool + zap logger
│   │   ├── handler/       ← HTTP handlers (thin: bind → service → respond)
│   │   ├── logging/       ← zap logger builder
│   │   ├── middleware/     ← Recovery, RequestID, CORS, Logger
│   │   ├── model/         ← GORM structs (BaseModel with UUIDv7)
│   │   ├── pagination/    ← universal list-endpoint params + sort whitelist
│   │   ├── repository/    ← all GORM/SQL queries
│   │   ├── response/      ← JSON envelope helpers (OK, Created, Paginated, Error)
│   │   ├── server/        ← Gin engine + route table + graceful shutdown
│   │   └── service/       ← business logic (no gin, no GORM)
│   ├── migrations/        ← gormigrate steps (append-only)
│   ├── Dockerfile
│   ├── .env.example
│   └── go.mod
└── frontend/              ← Next.js app (Materialize MUI starter-kit)
    ├── src/
    ├── .eslintrc.js
    ├── tsconfig.json
    └── package.json
```

---

## 3. ABSOLUTE RULES — VIOLATIONS ARE REJECTED

### 3.1 DRY / Single Source of Truth

- **Before writing ANY new code**, search for an existing helper, type,
  constant, or pattern and **reuse or extend** it.
- Each business rule, status map, error code, and type is defined **once** and
  **imported everywhere** it is needed.
- **Never duplicate** logic between handler ↔ service, or frontend ↔ backend
  types.

### 3.2 Imports Over Inline

- **Use imports for everything.** If a utility exists in a shared package,
  import it. Do not copy-paste.
- Go: import from `internal/` packages. Frontend: use path aliases (`@/*`,
  `@core/*`, `@components/*`, etc.).

### 3.3 Layered Architecture (Backend)

A request flows **one way only**:

```
router → middleware → handler → service → repository → GORM → PostgreSQL
```

| Layer | Package | Rules |
|---|---|---|
| **Handlers** | `internal/handler` | **Thin:** bind + validate → call service → write envelope. **No business logic. No DB access.** |
| **Services** | `internal/service` | **All business logic.** Depend on repository interfaces. **Never import `gin`.** |
| **Repositories** | `internal/repository` | **All GORM/SQL.** Nothing above writes queries. |
| **Models** | `internal/model` | GORM structs. Embed `BaseModel` (UUIDv7 PK + timestamps + soft delete). |
| **Config** | `internal/config` | **Only** place that reads env vars. Everything else gets typed config via constructor injection. |
| **Response** | `internal/response` | **Only** place that builds JSON envelopes. Never hand-assemble JSON. |

### 3.4 Response Envelope (Every Endpoint)

```jsonc
// success
{ "success": true, "data": { … }, "meta": { "pagination": { … } } }
// error
{ "success": false, "error": { "code": "…", "message": "…", "fields": { … } } }
```

- Produced by `response.OK()`, `response.Created()`, `response.Paginated()`,
  `response.Error()` — **never hand-build.**
- **5xx messages are sanitized.** The real error is logged server-side with the
  request id. Never leak GORM/Postgres errors to clients.

### 3.5 Error Handling

- Define **sentinel errors** in the service layer using `internal/apperr`.
- Map them to HTTP status in the handler via `response.Error()`.
- Mapping: validation → 422, not found → 404, conflict → 409, unauthorized →
  401, forbidden → 403.

### 3.6 Dependency Injection

- **Manual constructor injection** wired top-to-bottom in `cmd/api/main.go`.
- **No DI framework. No globals. No `init()` functions.**
- If you want to know what the app is made of, read `main.go` top to bottom.

### 3.7 Migrations (gormigrate)

- Each migration has a stable `ID`, a `Migrate`, and a `Rollback`.
- **Append only** — never edit a migration that has been applied. Add a new one.
- **No bare `AutoMigrate`.** Every schema change is a gormigrate step.
- Register new steps in `migrations.All()`.

### 3.8 Config & Secrets

- All configuration comes from **environment variables** (via Viper in
  `internal/config`), each with a sane default.
- Commit `.env.example`; **never commit** `.env`, secrets, or keys.
- Optional integrations stay inert when their env is unset.

### 3.9 No Test Files (Backend)

- **Do not create `*_test.go` files** for the backend during this hackathon.
  We skip tests to save time. Quality is enforced by `go build`, `go vet`,
  and the frontend linter instead.

### 3.10 Frontend — Strict Linting & Type Checks

- `tsconfig.json` has **`strict: true`** plus `noUncheckedIndexedAccess`,
  `noImplicitReturns`, `noFallthroughCasesInSwitch`.
- ESLint is configured with **`@typescript-eslint/no-explicit-any: warn`**
  and **`no-console: warn`**.
- **Every PR must pass:** `npm run lint` and `npm run type-check`.
- **No `// @ts-ignore` or `// @ts-nocheck`** unless there is an inline comment
  explaining why.

### 3.11 Pagination

- Every list endpoint supports: `page`, `page_size`, `q`, `sort_by`,
  `sort_order`.
- `sort_by` is validated against a **column whitelist** (`pagination.Sortable`)
  — never interpolate raw input into SQL.

### 3.12 Git Hygiene

- **Single branch:** `main`.
- Keep commits focused and descriptive.
- Never commit `node_modules/`, `.next/`, `.env`, `*.pem`, `*.key`, or
  build artifacts.

---

## 4. Tech Stack Reference

### Backend

| Concern | Choice | Import path |
|---|---|---|
| HTTP framework | `gin-gonic/gin` | Use `gin.New()` + explicit middleware |
| ORM | `gorm.io/gorm` + `gorm.io/driver/postgres` | Pooled `*sql.DB` |
| Migrations | `go-gormigrate/gormigrate/v2` | Versioned, append-only |
| Config | `spf13/viper` | Env + `.env` + typed defaults |
| Logging | `go.uber.org/zap` | Structured; request-scoped via ctxkey |
| Validation | `go-playground/validator/v10` | Via Gin `binding` tags |
| IDs | `google/uuid` (UUID v7) | Time-ordered, non-enumerable |

### Frontend

| Concern | Choice |
|---|---|
| Framework | Next.js 16 / React 19 |
| UI Library | MUI 7 (Materialize template) |
| Type Safety | TypeScript (strict mode) |
| Data Fetching | Axios (`@/libs/axios`) + `@tanstack/react-table` |
| Linting | ESLint + Prettier |
| Path Aliases | `@/*`, `@core/*`, `@layouts/*`, `@menu/*`, `@components/*` |

#### Shared Frontend Architecture (STRICT DRY)

Before building any UI screen, **reuse these canonical helpers**:

- **HTTP Client**: `import axios from '@/libs/axios'` — preconfigured with `NEXT_PUBLIC_API_URL`, bearer token injection, and 401 interceptors.
- **API Types**: `import type { StandardResponse, ListQueryParams, PaginationMetadata, SortOrder } from '@/types/apiTypes'`
- **Error Handling**: `import { handleApiError, getApiErrorMessage } from '@/utils/handleApiError'` — maps backend field validation errors (`error.fields`) onto `react-hook-form` fields.
- **Server Table State**: `import { useServerTable } from '@/hooks/useServerTable'` — single source of truth for `page`, `pageSize`, `search`, `sortBy`, `sortOrder`, and `params`.
- **Data Table**: `import DataTable from '@components/DataTable'` — TanStack table wrapper with server-side pagination, server sorting headers, skeleton loaders, and empty state.
- **Table Filters**: `import DataTableFilters from '@components/DataTableFilters'` — debounced search input, filter chips, action button, and reset toolbar.
- **Status Chips**: `import StatusChip from '@components/StatusChip'` — tonal status chip with leading dot indicator, resolving colors via `@/configs/statusColors`.
- **Stat Cards**: `import StatCard from '@components/StatCard'` — KPI stat card with count-up animation (`useCountUp`), accent bar, and trend indicator.
- **Panel Cards**: `import PanelCard from '@components/PanelCard'` — outlined panel with icon badge, title, count chip, action slot, and collapsible body.
- **Empty States**: `import EmptyState from '@components/EmptyState'` — friendly empty state with animated SVG ECG trace backdrop.
- **Confirm Dialogs**: `import ConfirmDialog from '@components/ConfirmDialog'` — standard confirmation modal.

---

## 5. Running the Project

### Full stack (Docker)

```bash
cp .env.example .env
docker compose up -d
```

Compose starts Postgres → runs migrations → starts API → starts web.

### Hot-reload development

```bash
make dev     # Postgres + API with auto-migrate
make web     # Next.js dev server (separate terminal)
```

### Common commands (via Makefile)

```bash
make help           # list all targets
make build          # compile backend
make vet            # go vet
make lint           # frontend ESLint
make fmt            # format everything
make tidy           # go mod tidy
make migrate-up     # apply pending migrations
make migrate-down   # roll back one migration
make migrate-status # show migration status
make db-reset       # destroy DB and recreate
make up             # docker compose up -d --build
make down           # docker compose down
make logs           # docker compose logs -f
```

---

## 6. Adding a New Feature (Checklist)

1. **Model** — add GORM struct in `internal/model/` embedding `BaseModel`.
2. **Migration** — add a gormigrate step in `migrations/` and register in
   `migrations.All()`.
3. **Repository** — add query methods in `internal/repository/`.
4. **Service** — add business logic in `internal/service/`. Use `apperr` for
   domain errors.
5. **Handler** — add thin HTTP handlers in `internal/handler/`. Use
   `response.*` for envelopes, `response.BindJSON` for validation.
6. **Routes** — register in `internal/server/server.go` `registerRoutes()`.
7. **Wire** — connect in `cmd/api/main.go` (repo → service → handler).
8. **Frontend** — add pages/components using MUI + path aliases. Keep types
   strict.

---

## 7. Definition of Done

A change is not complete until:

1. `cd backend && go build ./... && go vet ./...` — clean.
2. Any DB change ships as a new gormigrate step.
3. Frontend: `npm run lint && npm run type-check` — clean.
4. Old code is gone (clean replacement) — grep for old symbols, confirm zero.
5. `docker compose build` still works.
6. Committed to `main` and pushed.

---

## 8. What NOT to Do

- ❌ Do NOT create test files (`*_test.go`) for the backend.
- ❌ Do NOT use `AutoMigrate` directly. Use gormigrate steps.
- ❌ Do NOT write queries outside `internal/repository/`.
- ❌ Do NOT put business logic in handlers.
- ❌ Do NOT import `gin` from a service.
- ❌ Do NOT hand-build JSON responses — use `response.*` helpers.
- ❌ Do NOT commit `.env`, secrets, or `node_modules/`.
- ❌ Do NOT use `// @ts-ignore` without an explanation comment.
- ❌ Do NOT duplicate code — import shared helpers.
- ❌ Do NOT use `any` in TypeScript unless absolutely necessary (warned by ESLint).
- ❌ Do NOT use `console.log` in production code (warned by ESLint).
