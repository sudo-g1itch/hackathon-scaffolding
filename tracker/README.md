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
