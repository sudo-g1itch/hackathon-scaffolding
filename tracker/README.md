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
