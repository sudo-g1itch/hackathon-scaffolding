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
├── tracker/               ← session implementation log & task tracker
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

### 3.13 Use Existing Packages Over Custom Implementations

- **Always prefer existing packages/libraries** (Go modules & npm packages) over reinventing features from scratch to maximize hackathon development speed.
- If a tested third-party package or internal utility exists, import and reuse it.

### 3.14 Task & Session Tracker (`tracker/`)

- Maintain a dedicated `tracker/` folder at the repo root (`tracker/README.md`) containing an up-to-date record of implemented tasks, features, and session history.
- **`tracker/` and `GEMINI.md` MUST be updated on every non-trivial change.**

### 3.15 Dedicated Port Range Allocation (20000 - 21000)

Because multiple projects run concurrently on this host system, this application stack MUST use ports strictly in the **20000 - 21000** range:
- **Frontend (Web):** `20000` (Dev & Prod)
- **Backend (API):** `20080` (Dev & Prod)
- **PostgreSQL Database:** `20543` (Host DB port)

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
- **Auth & RBAC Context**: `import { useAuth } from '@/contexts/AuthContext'` — access `user`, `login`, `logout`, `register`, `hasRole(role)`.
- **Auth Guard**: `import AuthGuard from '@components/auth/AuthGuard'` — client-side route authentication wrapper.
- **Role Guard**: `import RoleGuard from '@components/auth/RoleGuard'` — RBAC wrapper checking `roles={['admin']}` before rendering protected UI.

---

## 4.5 AnchorOne domain (the product)

AnchorOne is an AI recovery companion: voice check-in → risk detection →
intervention, around a recovery plan the person and their caregiver both work on.

**Two people use it, and they see different apps.**

| | Person in recovery (`user`) | Caregiver (`caregiver`) |
|---|---|---|
| Sees | Dashboard, check-ins, AI coach, timeline, their goals, their caregiver | The people who chose them, each one's signals + plan, the conversation |
| Does | Checks in, logs goal progress, messages their caregiver | Watches risk, suggests goals, encourages, messages |
| Never sees | The caregiver's other people | Transcripts; summaries unless the user consents |

The feature code lives in `internal/{model,repository,service,handler}/` under
`recoverai*` (the user's own recovery), `goal*` (the plan), `support*` (the
conversation), `care*` (the caregiver's views and shared access rules), and in
`frontend/src/{app/(dashboard)/anchor-one,components/anchor-one}`.

### API surface (all under `/api/v1`, all authenticated)

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/capabilities` | Which AI integrations have keys — the UI disables features accordingly |
| `POST` | `/checkin` | Voice check-in: multipart `audio` → Deepgram → Gemini → stored `Checkin` |
| `POST` | `/risk` | Typed check-in — same analysis, no microphone |
| `POST` | `/voice/transcribe` | Speech → text only |
| `POST` | `/voice/speak` | Text → MP3 (**answers `audio/mpeg`, not the JSON envelope**) |
| `GET` | `/dashboard` | Mood, risk, craving, streak, counts, profile, capabilities |
| `GET` | `/timeline` | Check-ins + emergencies merged, reverse-chronological |
| `POST` | `/emergency` | Crisis plan: actions, caregiver SMS, grounding, encouragement |
| `POST`/`GET` | `/coach/chat`, `/coach/history` | Recovery coach conversation |
| `POST` | `/education` | Plain-English explainer |
| `GET`/`PUT` | `/profile` | Recovery goal, substance, caregiver + emergency contacts, sharing consent |
| `GET` | `/caregivers`, `PUT` `/profile/caregiver` | List / link (or unlink with `null`) a caregiver |
| `GET` | `/caregiver` | The caregiver's list of everyone who chose them — **`caregiver` or `admin` role only** |

#### Recovery plan — the caller's own goals

| Method | Path | Purpose |
|---|---|---|
| `GET`/`POST` | `/goals` | List / add goals |
| `GET` | `/goals/summary` | Counts + average progress + next goal (the dashboard roll-up) |
| `GET`/`PUT`/`DELETE` | `/goals/:goalID` | A goal and its progress feed |
| `POST` | `/goals/:goalID/progress` | Log a step, a note, or (from a caregiver) encouragement |

#### Patient-scoped — the caller's own record *or* someone they support

**No route-level role guard, on purpose.** Who may read a given person's record
depends on whether that person linked this caregiver, which is a data question,
not a role question. `service.careAccess` answers it (see rule 8 below).

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/patients/:patientID` | Signals + check-in history + goals + plan summary |
| `GET`/`POST` | `/patients/:patientID/goals` | Read the plan / a caregiver suggesting a goal |
| `GET`/`POST` | `/patients/:patientID/messages` | The shared conversation (reading marks it read) |
| `POST` | `/patients/:patientID/messages/read` | Clear the badge without loading the thread |
| `GET` | `/messages/unread` | One number for the navigation badge |

### Rules specific to this domain

1. **AI integrations are optional.** A missing `AI_GEMINI_API_KEY` or
   `AI_DEEPGRAM_API_KEY` disables that feature (503 with an explanation via
   `apperr.Unavailable`) and **must never stop the API from booting.** Never add
   an AI key to `config.Validate()` as a required field.
2. **Never trust model output's shape.** Every JSON-returning Gemini call sends a
   `responseSchema` (see `checkinAnalysisSchema`, `emergencyPlanSchema`). Without
   it Gemini returns `craving` as prose and `risk` in mixed case, which breaks
   unmarshalling and every risk comparison. `model.NormalizeRisk` and
   `model.ClampCraving` are the second line of defence — keep using them.
3. **Risk levels and coach roles are constants** (`model.RiskLow/Medium/High`,
   `model.CoachRoleUser/AI`). Never write the string literals inline.
4. **Caregiver privacy is consent-gated, and the default is private.**
   `service.CaregiverPatient` (in `care_service.go`) still carries no transcript,
   summary or triggers. A check-in's *narrative* reaches a caregiver through
   `PatientOverview` only when the user has switched on
   `RecoveryProfile.ShareCheckinDetails`, which defaults to `false` and is
   theirs alone to change. **The raw transcript is never projected at any
   setting** — see `toPatientCheckin`. Widening what a caregiver sees means
   adding another consent flag, not removing this one.
5. **Never claim an action the app did not take.** Emergency Mode hands the user
   a ready-to-send message; it does not send it, and the UI says so. Telling
   someone in crisis that help is coming when it is not is a safety bug.
6. **Frontend service paths are relative** (`/checkin`, not `/api/v1/checkin`) —
   the axios `baseURL` already carries `/api/v1`.
7. **DTOs are mirrored by hand.** `internal/service/{recoverai,goal,support,care}_service.go`
   and `frontend/src/types/anchorOneTypes.ts` are one contract; change both together.
8. **Access to a patient's record is link-based, not role-based.**
   `service.careAccess` is the only place that decides it, returning one
   `CareRelation` (`self` / `caregiver` / `admin`) that goals, the overview and
   messaging all check. Holding a `caregiver` account grants nothing until
   somebody chooses you. `RelationAdmin` is deliberately separate: an admin may
   read an overview but is **not** a party to the private conversation. Never
   re-derive this rule with an ad-hoc role check in a handler.
9. **A caregiver supports; they do not self-report on someone's behalf.** They
   may suggest a goal and leave encouragement. Only the person in recovery moves
   their own numbers (`GoalService.LogProgress`). Enforce this in the service —
   the UI hiding a button is a courtesy, not the rule.
10. **Goal progress has one definition.** `model.RecoveryGoal.ProgressPercent()`,
    clamped 0–100. Never recompute a percentage in a handler or a React
    component, and never write a status string inline — use
    `model.GoalStatus*` / `model.GoalCategory*`.
11. **`frontend/src/configs/navigation.ts` is the single source for who sees what.**
    The sidebar, the horizontal menu, `RouteGuard` and the post-login landing
    path all read it. Adding a screen means adding it there — otherwise it is
    either invisible or reachable by anyone with the URL.

### AI configuration

| Env | Default | Notes |
|---|---|---|
| `AI_GEMINI_API_KEY` | *(empty)* | Unset ⇒ analysis, coach, emergency, education disabled |
| `AI_GEMINI_MODEL` | `gemini-3.5-flash` | `gemini-1.5-flash` is **retired** and 404s |
| `AI_DEEPGRAM_API_KEY` | *(empty)* | Unset ⇒ transcription and playback disabled |
| `AI_DEEPGRAM_STT_MODEL` | `nova-2` | |
| `AI_DEEPGRAM_TTS_MODEL` | `aura-asteria-en` | |
| `AI_REQUEST_TIMEOUT` | `60s` | Applied to the provider HTTP clients |
| `AI_MAX_AUDIO_BYTES` | `10485760` | Upload cap for a check-in recording |

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

---

## 9. Deployment (CI/CD)

Deployment is **fully automated via GitHub Actions on a self-hosted runner** that
lives on the deploy server itself. **To deploy, you just push to `main`** — the
runner rebuilds and restarts the stack in place. There is nothing to run by hand
in the normal flow.

### 9.1 How it works

```
push to main ──► GitHub Actions ──► self-hosted runner (ON the server)
                                     │
                                     ├─ write production .env (secrets + host)
                                     ├─ docker compose build
                                     ├─ run DB migrations (explicit)
                                     ├─ docker compose up -d
                                     └─ health-check API /healthz
```

| Piece | Location | Purpose |
|---|---|---|
| **CI** | `.github/workflows/ci.yml` | On PRs to `main` + non-`main` pushes: Go `build`/`vet`/`test` + frontend `lint`/`type-check`/`build`. |
| **CD** | `.github/workflows/deploy.yml` | On **push to `main`**: build images, migrate, `up -d`, health-check. |
| **Prod override** | `docker-compose.prod.yml` | Production-only compose layer (see 9.3). |
| **Runner** | server, systemd service | Executes the jobs. Runs as `root`. |

Both workflows run on `runs-on: [self-hosted, hackathon]`. The repo is **private**,
so all jobs use the self-hosted runner (no GitHub-hosted minutes).

### 9.2 The server & runner

- **Server:** Ubuntu 24.04, public IP `69.164.244.74`. Also hosts other unrelated
  runners — **only touch the hackathon one.**
- **Runner:** name `hackathon-interserver-dev`, dir `/opt/actions-runner-hackathon`,
  systemd unit `actions.runner.sudo-g1itch-hackathon-scaffolding.hackathon-interserver-dev.service`.
- Check / restart it:
  ```bash
  systemctl status  actions.runner.sudo-g1itch-hackathon-scaffolding.hackathon-interserver-dev.service
  systemctl restart actions.runner.sudo-g1itch-hackathon-scaffolding.hackathon-interserver-dev.service
  ```
- **Public URL: `https://anchorone.dcix.in`** — this is the address to use.
  The direct ports (`http://69.164.244.74:20000` / `:20080`) still answer, but
  the browser app is built for the HTTPS origin.

### 9.2.1 TLS & the public domain

nginx on the same server terminates TLS and proxies to the containers. The vhost
is **its own file**, `/etc/nginx/conf.d/anchorone.conf` — every other site on this
box lives in the shared `default.conf`, so keep AnchorOne out of it.

```
https://anchorone.dcix.in/       → 127.0.0.1:20000   (Next.js)
https://anchorone.dcix.in/api/   → 127.0.0.1:20080   (Go API)
http://anchorone.dcix.in/        → 301 to https
```

- **Certificate:** Let's Encrypt, issued and auto-renewed by certbot using the
  webroot shared with the other vhosts:
  ```bash
  certbot certonly --webroot -w /var/www/letsencrypt -d anchorone.dcix.in
  ```
- The `/api/` block sets `client_max_body_size 12m` (a voice check-in uploads up
  to `AI_MAX_AUDIO_BYTES`, and the proxy's 1 MB default would reject it) and
  120s read/send timeouts (Gemini and Deepgram calls outlast the 60s default).
- **`NEXT_PUBLIC_API_URL` must stay `https://`.** A page served over HTTPS may
  not call an `http://` API — the browser blocks it as mixed content — and the
  value is baked in at *build* time, so changing it needs a rebuild, not a
  restart.

### 9.3 Production vs. development compose

- **Dev** (`docker compose up`, `make up`): auto-loads `docker-compose.override.yml`.
- **Prod** (what CD runs): explicit files, so the dev override is **NOT** loaded:
  ```bash
  docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
  ```
  `docker-compose.prod.yml` does two things production requires:
  1. Passes `NEXT_PUBLIC_API_URL` as a **build arg** (Next.js inlines `NEXT_PUBLIC_*`
     at build time — a runtime env var is ignored).
  2. Sets `APP_AUTO_MIGRATE=false` on the `migrate` service.

### 9.4 Production config guards (IMPORTANT)

When `APP_ENV=production`, `internal/config` **rejects** the app on startup unless:

| Guard | Required value | Why the deploy sets it |
|---|---|---|
| `app.auto_migrate` | `false` | Migrations run as an explicit step, never on boot. Set via `APP_AUTO_MIGRATE=false`. |
| `database.ssl_mode` | **not** `disable` | Deploy uses `DATABASE_SSL_MODE=prefer` — TLS with plaintext fallback, since the bundled Postgres has no TLS (`require` would fail to connect over the internal network). |

If you add new production guards in `internal/config`, **update the `.env` block in
`deploy.yml` accordingly** or the deploy will fail.

### 9.5 Secrets & config

- Repo secret **`DATABASE_PASSWORD`** (GitHub → Settings → Secrets) is injected into
  the production `.env` at deploy time. Never hard-code it.
- Repo secrets **`AI_GEMINI_API_KEY`** and **`AI_DEEPGRAM_API_KEY`** are injected the
  same way. They are **optional**: if unset the deploy still succeeds and the API
  boots with those features disabled (it logs a warning at startup naming each
  missing key). Add them under GitHub → Settings → Secrets to enable the AI demo
  in production.
- The production `.env` is generated fresh each deploy inside `deploy.yml`
  (`git clean` wipes it between runs) — do **not** rely on a hand-placed `.env` on
  the server.
- Ports stay in the **20000–21000** range (see 3.15).

### 9.6 Manual deploy / rollback (only if Actions is unavailable)

SSH to the server, then in the runner's checkout
(`/opt/actions-runner-hackathon/_work/hackathon-scaffolding/hackathon-scaffolding`):

```bash
# redeploy current main
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# apply migrations explicitly
docker compose -f docker-compose.yml -f docker-compose.prod.yml run --rm migrate

# rollback: check out a known-good commit, then redeploy the same way
```

### 9.7 Deploy troubleshooting

- **Watch a run:** `gh run list --workflow deploy.yml` then
  `gh run view <id> --log-failed`.
- **App up but frontend calls the wrong API:** `NEXT_PUBLIC_API_URL` is baked at
  **build** time — a rebuild (not just restart) is required after changing it.
- **Migration/boot fails with "invalid configuration":** a production guard (9.4)
  is unsatisfied — check the generated `.env` values in `deploy.yml`.
