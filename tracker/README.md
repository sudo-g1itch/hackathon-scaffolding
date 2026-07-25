# Project Implementation Tracker

This directory tracks all tasks, features, architectural decisions, and session logs implemented across the project lifecycle.

---

## Port Allocation Matrix (Range: 20000 - 21000)

| Service | Dev Port | Docker / Prod Port | Notes |
|---|---|---|---|
| **Frontend (Next.js)** | `20000` | `20000` | Standardized in `package.json`, `docker-compose.yml`, and `CORS` |
| **Backend API (Go Gin)** | `20080` | `20080` | Standardized in `config.go`, `docker-compose.yml`, and `axios.ts` |
| **PostgreSQL Database** | `20543` | `20543` | Published host port in `docker-compose.override.yml` |

---

## Session Implementation History

### Session 1: Project Scaffolding & Shared DRY Architecture
- **Infrastructure:**
  - Configured Go Gin monolith, Viper config loader, Zap logger, GORM ORM, and Gormigrate versioned migrations.
  - Setup Next.js 16 (React 19) Materialize MUI 7 frontend starter kit.
  - Formulated strict rules in `GEMINI.md`.
- **Reusable Frontend UI Component Library (Ported from RCM Platform):**
  - `DataTable` (`@components/DataTable`)
  - `DataTableFilters` (`@components/DataTableFilters`)
  - `StatCard` (`@components/StatCard`)
  - `PanelCard` (`@components/PanelCard`)
  - `StatusChip` (`@components/StatusChip`)
  - `ConfirmDialog` (`@components/ConfirmDialog`)
  - `EmptyState` (`@components/EmptyState`)
  - `useServerTable` (`@/hooks/useServerTable`) & `usePagination` (`@/hooks/usePagination`)
  - `handleApiError` (`@/utils/handleApiError`) & `axios` instance (`@/libs/axios`)

---

### Session 2: User Authentication & Role-Based Access Control (RBAC)
- **Backend Auth Architecture:**
  - `User` model (`internal/model/user.go`) with bcrypt password hashing & UUIDv7 PK.
  - Migration `0002_create_users_table.go` with default admin seed (`admin@hackathon.local` / `Admin123!`).
  - `UserRepository` (`internal/repository/user_repository.go`) with pagination whitelist and full-text search.
  - `AuthService` (`internal/service/auth_service.go`) handling login, registration, JWT validation, and user profile retrieval.
  - Middleware (`internal/middleware/auth.go`) providing `Authenticate` (JWT token check) and `RequireRole` (RBAC authorization).
  - Routes registered in `internal/server/server.go`: `POST /api/v1/auth/login`, `POST /api/v1/auth/register`, `GET /api/v1/auth/me`, `GET /api/v1/users` (Admin only).
- **Frontend Auth & RBAC Architecture:**
  - `AuthContext` (`src/contexts/AuthContext.tsx`) managing identity state, local storage tokens, login/register/logout actions, and `hasRole` helper.
  - `AuthGuard` (`src/components/auth/AuthGuard.tsx`) for client-side route protection.
  - `RoleGuard` (`src/components/auth/RoleGuard.tsx`) for component-level RBAC authorization.
  - Updated `Login.tsx` view and `UserDropdown.tsx` with live user details and role status chip.

---

### Session 3: Port Standardization & Strict Rule Enforcement
- Standardized project stack ports to range `20000 - 21000`.
- Codified absolute rules in `GEMINI.md`:
  - Always use existing packages/libraries instead of custom implementations.
  - Maintain session tracking in `tracker/`.
  - Maintain `GEMINI.md` and `tracker/` updated on every change.

---

### Session 4: Docker Stack Setup & Build Pipeline Fixes
- **Docker Compose & Container Build Pipeline:**
  - Created `frontend/Dockerfile` (multi-stage Next.js production build) and `frontend/.dockerignore`.
  - Configured `ENV GOTOOLCHAIN=auto` in `backend/Dockerfile` for Go toolchain downloads.
- **Frontend Type Safety & Linting:**
  - Resolved `noUncheckedIndexedAccess` and `noImplicitReturns` TypeScript errors across `@core/components/customizer`, `settingsContext`, `useMediaQuery`, `bundle-icons-css`, and menu components.
  - Passed `npm run type-check` and `npm run lint` cleanly.
- **Stack Verification:**
  - Successfully deployed full local stack (`Postgres`, `Migrate`, `API`, `Web`) via `make up`.
  - Verified API health endpoint (`GET http://localhost:20080/healthz` -> 200 OK) and Web frontend running on port 20000.

---

### Session 5: Complete User Management & RBAC Module Implementation
- **Backend Architecture:**
  - GORM Models: `Permission`, `Role`, and `RolePermission` (`internal/model/role.go`).
  - Migration `0003_create_roles_permissions.go` seeding default permissions and roles (`admin`, `manager`, `user`).
  - Repositories & Services: `RoleRepository` (`internal/repository/role_repository.go`), `UserRepository` extension, and `RBACService` (`internal/service/rbac_service.go`).
  - REST API Handlers: `UserHandler` CRUD endpoints and `RoleHandler` CRUD & permission endpoints (`internal/handler/`).
  - Composition Root Wiring: Registered in `cmd/api/main.go` and protected with `RequireRole("admin")` middleware in `internal/server/server.go`.
- **Frontend UI Screens & Navigation:**
  - TypeScript Types (`src/types/rbacTypes.ts`) & Service API wrapper (`src/services/rbacService.ts`).
  - `/admin/users`: Server-side paginated user management data table with KPI stat cards, role/status filters, user creation/editing modal, and confirmation dialogs.
  - `/admin/roles`: Roles management dashboard with interactive Permission Matrix dialog grouped by system module.
  - `/admin/permissions`: Permissions reference catalog page.
  - Navigation: Integrated Admin & RBAC section into `VerticalMenu.tsx`.
- **Quality & Stack Verification:**
  - `npm run type-check && npm run lint` — Passed with 0 errors.
  - `make up` — Docker stack built and started all containers cleanly.

---

### Session 6: Next.js Security Patch (RCE Fix)
- **Frontend Dependency Update:**
  - Upgraded Next.js from `16.0.4` to `latest` in `frontend/package.json` to mitigate a Remote Code Execution (RCE) vulnerability.
  - Re-ran `npm install` to update lockfiles.

---

### Session 7: AnchorOne Recovery Platform — PRD Implementation & Remediation

Audited the existing AnchorOne feature set against the PRD and fixed the reasons
it did not work end to end, then completed the missing features.

**Blocking defects found and fixed**
- **Gemini model 404'd.** `gemini-1.5-flash` is no longer served for this key, so
  *every* AI feature failed. Now `gemini-3.5-flash` (configurable via
  `AI_GEMINI_MODEL`), verified live.
- **Untyped AI output.** Gemini answered `craving` as prose ("Moderate-to-High")
  and `risk` in mixed case, which failed to unmarshal into `Craving int` and
  broke every `risk === 'HIGH'` comparison. Fixed with Gemini **response
  schemas** (`INTEGER` + risk `enum`), plus `model.NormalizeRisk` /
  `model.ClampCraving` as a safety net.
- **Frontend called the wrong URLs.** `anchorOneService` prefixed `/api/v1`,
  which the axios `baseURL` already carries — every request resolved to
  `/api/v1/api/v1/...` and 404'd. Paths are now relative like every other service.
- **Production deploy could not boot.** `config.Validate()` hard-required both AI
  keys while `deploy.yml` never wrote them. Keys are now **optional** (rule 3.8);
  `deploy.yml` passes `AI_GEMINI_API_KEY` / `AI_DEEPGRAM_API_KEY` secrets.
- **Emergency Mode 500'd for new users.** It required a pre-existing
  `recovery_profiles` row, and no endpoint could create one. Profiles are now
  created on demand and there is full profile CRUD.
- **Emergency Mode lied.** The UI claimed "Your caregiver has been notified" when
  nothing was sent, and never rendered `emergency_sms` at all.
- **Coach double-sent each message.** History was read *after* saving the new
  turn, so the model saw it twice. History is now read first.
- **`GetLastCheckin` returned a zero-valued struct** with its error, so a user
  with no check-ins looked like one with an empty check-in on the caregiver
  dashboard. Lookups now return `(nil, nil)` when absent.
- **AI output was silently dropped:** `recommended_actions` (check-ins) and
  `grounding_exercise` / `encouraging_message` (emergencies) were generated,
  shown once, then discarded. Persisted via migration `0006`.
- **Recovery streak was hardcoded to `0`.** Now computed from distinct check-in
  days (`CheckinDays` + `calculateStreak`).
- No timeouts on any Gemini/Deepgram call (`http.DefaultClient`); unbounded audio
  upload; the API key was passed in the URL query string.

**Features completed**
- `POST /voice/transcribe`, `POST /voice/speak` (Deepgram TTS was dead code — no
  route existed), `POST /risk` (typed check-in), `GET|PUT /profile`,
  `GET /coach/history`, `GET /capabilities`.
- Timeline is now one merged reverse-chronological feed, not two lists.
- Emergency Mode renders the caregiver script with Copy / Speak / SMS / Call, and
  states plainly that nothing has been sent yet.
- Voice check-in narrates record → transcribe → analyse and shows the result
  (risk, craving, triggers, actions), with a typed fallback.
- New `/anchor-one/profile` screen: goal, substance, caregiver + emergency
  contacts — the personalisation every prompt depends on.
- Spoken playback (`SpeakButton`) on plans, coach replies and education answers.

**Architecture & privacy**
- Replaced `map[string]interface{}` payloads with typed DTOs mirrored one-for-one
  by `frontend/src/types/anchorOneTypes.ts`.
- Caregiver dashboard restricted to `caregiver`/`admin` via `RequireRole`, and
  its DTO carries **no transcript, summary or triggers** — the PRD's privacy
  guarantee is enforced in the type, not by the UI choosing what to render.
- Added `apperr.CodeUnavailable` → HTTP 503 so a missing/failing integration is
  explained rather than surfacing as a generic 500; `response.Error` now
  sanitizes only `CodeInternal`.
- Migration `0006_extend_recoverai_tables` (idempotent — step 0004 AutoMigrates
  live model structs) plus composite indexes for the dashboard/streak queries.

**Verification**
- `go build ./... && go vet ./...` — clean.
- `npm run lint && npm run type-check` — 0 errors (also fixed 27 pre-existing
  lint errors).
- End-to-end against **real Gemini + Deepgram**: profile → caregiver link →
  text check-in (`risk: HIGH`, `craving: 9` as an int) → emergency plan
  (persisted) → 2-turn coach (no duplication) → education → dashboard → merged
  timeline → caregiver view (403 for non-caregivers, no private data) → TTS
  (`audio/mpeg`) → **full voice pipeline** (TTS mp3 → Deepgram → Gemini) →
  silence/validation guards (422).
- Full stack rebuilt via `docker compose up -d --build`; all five AnchorOne
  screens plus the emergency and check-in flows driven in headless Chromium with
  zero console errors.

---

### Session 5: Recovery Plan (Multi-Goal), Caregiver Workflow & Role-Gated UI

**The gaps this session closed**
- The "recovery plan" was a single free-text sentence on the profile. There was
  no way to state more than one commitment, no target, no progress, no history.
- The caregiver role had a read-only risk table and nothing else: no way to see
  a person's plan or check-in history, and no way to say anything to them.
- Every signed-in account saw every screen. A caregiver was shown a personal
  check-in dashboard they have no data for; a user was shown Admin & RBAC.

**Recovery plan — many goals per person**
- `model.RecoveryGoal` — title, description, category, status, `current/target`
  value + unit, optional target date. `ProgressPercent()` is the single
  definition of "how far along", clamped 0–100 so a lowered target cannot report
  340%.
- `model.GoalUpdate` — one chronological feed carrying progress steps, the
  user's own notes and a caregiver's encouragement, each stamped with its author
  and role.
- Status vocabulary (`active`/`completed`/`paused`/`archived`) and categories are
  constants with `ValidGoalStatus` / `ValidGoalCategory` whitelists.
- A goal auto-completes on reaching its target and reopens if the target is
  raised (`settleCompletion`), so a plan can never show "done" at 60%.
- Open goals are preloaded into `RecoveryProfile.Goals` and rendered into every
  Gemini prompt by `goalContext()` — the coach now knows what the person is
  actually working on, not just which substance they named.

**Caregiver workflow**
- `service.CareService` — `ListPatients` and `GetPatientOverview` (signals +
  check-in history + goals + plan summary), composing `GoalService` and
  `SupportService` so progress is computed in exactly one place.
- `model.SupportMessage` — the private two-person conversation, keyed by the
  `(patient, caregiver)` pair. Both sides read and write the same thread;
  opening it marks the other side's messages read, which drives the unread badge
  on both dashboards.
- A caregiver may **suggest** a goal and leave encouragement, but may not move
  someone else's numbers — enforced in `GoalService.LogProgress`, not the UI.
- `CaregiverPatient` gained plan progress and unread counts; it still carries no
  transcript, summary or trigger.

**Privacy — consent replaces a blanket rule**
- New `recovery_profiles.share_checkin_details` (defaults **false**). A caregiver
  always sees risk, mood, craving and streak; they see a check-in's *summary and
  triggers only if the person switched sharing on themselves*, from their
  recovery plan screen. The raw transcript is never projected at any setting
  (`toPatientCheckin`).
- This is the deliberate decision GEMINI.md rule 4.5.4 asks for: the caregiver
  view widened, but only by the user's own choice.

**Authorisation**
- `service.careAccess` resolves one `CareRelation` (`self` / `caregiver` /
  `admin`) used by goals, the overview and messaging alike. The check is
  **link-based, not role-based**: holding a caregiver account grants nothing
  until someone chooses you.
- `/patients/:patientID/*` therefore carries no route-level role guard — a role
  check would be either too loose or redundant. An admin may read an overview
  but is refused the private thread.

**Role-gated UI**
- `frontend/src/configs/navigation.ts` is the single source for who sees what.
  The sidebar, the horizontal menu, the new `RouteGuard` and the post-login
  landing path all read it, so a hidden screen is also an unreachable one.
- `user`/`manager` → recovery suite; `caregiver` → People I Support + messages;
  `admin` → administration + caregiver oversight; Education is open to all.
- Deleted the stale `data/navigation/*MenuData.tsx` (they still pointed at the
  removed `/home` and `/about` pages).

**New surface**
- API: `GET|POST /goals`, `GET /goals/summary`, `GET|PUT|DELETE /goals/:goalID`,
  `POST /goals/:goalID/progress`, `GET /patients/:patientID`,
  `GET|POST /patients/:patientID/goals`,
  `GET|POST /patients/:patientID/messages`,
  `POST /patients/:patientID/messages/read`, `GET /messages/unread`.
  `GET /caregiver` moved from `RecoverAIHandler` to `CareHandler`.
- Screens: `/anchor-one/goals`, `/anchor-one/messages`,
  `/anchor-one/caregiver/[patientId]`.
- Components: `GoalCard`, `GoalFormDialog`, `GoalDetailDialog`, `SupportChat`,
  `RouteGuard`, `useUnreadMessages`.
- Migration `0007_create_recovery_plan_tables` (idempotent; drops GORM's
  no-cascade has-many FK so the explicit `ON DELETE CASCADE` is the only rule).

**Deployment**
- `anchorone.dcix.in` now serves the stack over HTTPS. nginx vhost
  `/etc/nginx/conf.d/anchorone.conf` on `69.164.244.74` terminates TLS
  (Let's Encrypt via certbot webroot) and proxies `/` → `:20000`,
  `/api/` → `:20080`, with a 12 MB body cap for check-in audio and 120s
  upstream timeouts for Gemini/Deepgram.
- `deploy.yml` now bakes `NEXT_PUBLIC_API_URL=https://anchorone.dcix.in/api/v1`
  — an https page may not call an http API, so the old `IP:20080` value would be
  blocked as mixed content.

**Verification**
- `go build ./... && go vet ./...` — clean.
- `npm run lint && npm run type-check && npm run build` — clean; all 15 routes
  compiled.
- Migration `0007` applied, rolled back and re-applied against Postgres; schema,
  partial unread index and the single cascade FK confirmed.
- End-to-end against the running stack with a real user + caregiver pair:
  link → two goals → progress `12/90 (13%)` → caregiver list shows plan progress
  → overview → two-way messaging with unread counts clearing correctly →
  caregiver encouragement recorded as `encouragement`/`caregiver` → caregiver
  progress write refused (403) → typed check-in analysed `MEDIUM` → **summary
  absent without consent, present after the user enabled it, transcript absent
  either way** → unlinked caregiver refused (403) → plain user refused the
  caregiver list (403) → admin allowed the overview but refused the thread →
  goal auto-completed at target with overshoot clamped.

---

### Session 6: Emergency Mode — a message that actually arrives

**The gap**
Emergency Mode produced a plan and a script, then stopped. The script could only
be copied, opened in the OS SMS composer, or dialled — none of which the app can
observe, so it could never honestly say anything had happened. There was also no
way to say *where* you are, and no way to speak instead of type at the moment
speaking is easiest.

**Sending is now real**
- `POST /emergency/:logID/alert` delivers the chosen script into the caregiver's
  **support thread** — the one channel the app controls and the caregiver
  already watches. `SupportMessage` gained `kind` + `emergency_id`, so an alert
  sits in the conversation but renders as an alert.
- `POST /emergency/:logID/acknowledge` (caregiver only, and only the linked one)
  stamps `acknowledged_at`. This is the **only** thing that lets the UI say
  "they have seen it" — before that it says "sent", and before sending it says
  nothing has been sent.
- With no caregiver linked, sending is refused (`422`) with an explanation and
  a pointer to the recovery plan screen. An alert is never recorded as sent with
  nobody to send it to.

**Personalised scripts**
- `model.EmergencyScriptPresets(senderName, caregiverName)` renders five presets
  ("I need someone", "Strong craving", "I'm not safe", "I slipped", "Just need
  to talk") already addressed and signed, so a person in crisis never faces a
  blank box. The AI draft is offered alongside and pre-filled when available;
  the presets work with no AI key at all.
- The user can edit any of it before sending; `sent_message` records what was
  actually said, separately from the AI's `generated_script`.

**Location**
- A per-alert toggle, not a stored setting. When on, the browser's geolocation
  is read at send time and rendered through `model.EmergencyLog.LocationURL()` —
  one definition of the Google Maps link, used by the thread, the caregiver's
  alert list and the timeline. A refused or unavailable location degrades to
  sending the message without it, and says so.

**Voice note**
- Record → Deepgram → **only the transcript is stored**; the audio is not kept.
  The caregiver reads it inline on the alert.
- This is consistent with the check-in privacy rule rather than an exception to
  it: a check-in is the user talking to the app, while this note was recorded
  specifically in order to be sent to that person.

**SMS and calls removed**
- No `sms:` links, no `tel:` links, no share sheet, and the caregiver phone
  field is gone from the recovery plan screen. The app cannot place a call or
  send a text; a button that hands off to the OS and hopes looks like delivery
  without being it. The `caregiver_phone` column stays for existing data.

**New surface**
- API: `POST /emergency/:logID/note`, `POST /emergency/:logID/alert`,
  `POST /emergency/:logID/acknowledge`. `POST /emergency` now also returns
  `presets`, `caregiver_linked` and `caregiver_name`.
- `PatientOverview.emergencies` lists only alerts that were actually sent — a
  plan the person worked through alone was never addressed to anyone.
- Migration `0008_emergency_alerts` (idempotent).

**Verification**
- `go build ./... && go vet ./...`, `npm run lint && type-check && build` — clean.
- Migration `0008` applied via the containerised migrate service.
- End-to-end: trigger → 5 personalised presets + AI draft → send with location
  (`share_location: true`, maps URL correct) → caregiver reads it in the thread
  as `kind: emergency` with the link → second alert with a **real Deepgram
  round-trip** voice note (TTS mp3 → transcribe → `"Kim, I am outside the bar on
  Jumeirah Road…"`) → sent with location off (no coordinates stored) →
  caregiver overview lists both → acknowledge → sender sees `acknowledged_at` →
  unrelated caregiver acknowledging refused (403) → user with no caregiver gets
  personalised presets but a 422 on send → another user touching someone else's
  emergency gets 404.

---

### Session 6b: Fixing the deploy — two migration bugs only a fresh database exposes

The Session 6 deploy failed at migration `0005`, and the investigation turned up
a second, worse bug behind it. Both come from the same root cause.

**Root cause: step `0004` AutoMigrates the LIVE model structs**, not a snapshot
of how they looked when it was written. On a database created today it therefore
emits columns and constraints that later steps also expect to create. Existing
databases never hit this, because those steps ran long ago — which is exactly
why it stayed hidden until production came up on a fresh volume
(`already_applied: 3`).

**Bug 1 — `0005` could not run at all on a new database.**
`0004` now creates `caregiver_id` *and* a `fk_recovery_profiles_caregiver`
constraint (the model gained the `Caregiver` belongs-to since). `0005`'s bare
`ADD CONSTRAINT` then died with SQLSTATE 42710, blocking every later step.

Fixed by making `0005` `DROP CONSTRAINT IF EXISTS` before adding — which also
matters on its own merits, since AutoMigrate's version has no `ON DELETE` clause
and would block deleting a caregiver account instead of unlinking the people
they support. This edits an applied migration, which rule 3.7 forbids; it is the
one case the rule cannot cover, because the step is unrunnable on a fresh
database and no later step can repair it. The edit is a no-op wherever `0005`
already succeeded.

**Bug 2 — a new user could not create their first goal.**
`RecoveryProfile.Goals` (added so open goals can be preloaded into AI prompts)
made GORM emit `fk_recovery_profiles_goals: recovery_goals.user_id ->
recovery_profiles.user_id`. As a constraint that asserts something the app does
not mean: *a goal may only exist if its owner already has a profile row.* Anyone
who opened "My Goals" before touching their recovery plan would have their first
goal rejected by Postgres.

The Session 5 end-to-end run missed it because that user linked a caregiver
first, which creates a profile row as a side effect. Migration `0009` drops the
accidental constraint; `fk_recovery_goals_user` (→ `users.id`, `ON DELETE
CASCADE`) was already the correct rule and remains the only one.

**Verification**
- All 9 migrations applied against a **virgin Postgres** (`applied_now: 9`), and
  every foreign key in the resulting schema audited by hand — one bogus
  constraint found, none left.
- The exact failing case reproduced before the fix and re-run after: a
  brand-new account with no profile row now creates a goal and loads its
  dashboard through the real API.
- Local stack rebuilt and healthy on `20000`/`20080`; certbot renewal dry-run
  for `anchorone.dcix.in` passes, timer active.
