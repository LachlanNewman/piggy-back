# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Stack

Full-stack web app: React 18 + Vite (frontend), Go 1.23 standard library (backend), PostgreSQL 17 with pgx/v5, Keycloak 26 (OIDC auth).

## Commands

### Full stack (Docker)
```bash
docker compose up           # Start all services
docker compose up --build   # Rebuild images then start
docker compose watch        # Start with live reload (backend rebuild on change, frontend sync)
```

### Frontend (`frontend/`)
```bash
npm run dev      # Vite dev server on port 5173
npm run build    # Production build
npm run preview  # Preview production build
```

Requires a `frontend/.env` file — see `frontend/.env.example` for required vars (`VITE_OIDC_AUTHORITY`, `VITE_OIDC_CLIENT_ID`, `VITE_OIDC_REDIRECT_URI`).

### Backend (`backend/`)
```bash
go run ./cmd/api   # Requires DATABASE_URL env var pointing to postgres
go test ./...      # Run all tests (unit tests mock the DB; no DATABASE_URL required)
```

## Architecture

### Services (Docker Compose)
| Service           | Port | Description                         |
|-------------------|------|-------------------------------------|
| postgres          | 5432 | App database                        |
| backend           | 8080 | Go HTTP API                         |
| frontend          | 5173 | Vite dev server                     |
| keycloak-postgres | —    | Keycloak's own database (internal)  |
| keycloak          | 8180 | OIDC identity provider              |

### Request flow
Frontend (React) → `/api/*` → Backend (Go HTTP) → PostgreSQL (pgx pool)

Auth flow: Frontend → Keycloak (OIDC code flow, port 8180) → tokens stored via `SplitTokenStore` (cookie + sessionStorage split) → `react-oidc-context` provides auth state.

### Backend (`backend/`)
Entry point: `cmd/api/main.go` — CORS middleware, route registration, HTTP server on `:8080`.

Package layout:
- `cmd/api/main.go` — wires config, DB pool, handlers, and starts server
- `config/config.go` — env-parsed config struct (uses `caarlos0/env`)
- `db/db.go` — `New()` creates pgxpool and runs goose migrations on startup; `NewDB()` returns `*DB` querier
- `db/migrations/` — SQL migration files (goose, embedded at build time)
- `db/users.go` — `DB.CreateUser` (upsert), `DB.GetUserBySubject`; defines `querier` interface for mocking
- `db/ride_requests.go` — ride request DB operations
- `handlers/users.go` — `CreateUser` handler (`POST /api/v1/users`)
- `handlers/users_me.go` — `GetUserMe` handler (`GET /api/v1/users/me?sub=...`)
- `handlers/ride_requests.go` — ride request handlers

No framework (no Gin/Echo/Chi); uses `net/http` only.

### API routes
| Method | Path                        | Handler            |
|--------|-----------------------------|--------------------|
| GET    | `/api/health`               | inline health check |
| GET    | `/api/hello`                | inline hello        |
| POST   | `/api/v1/users`             | CreateUser          |
| GET    | `/api/v1/users/me`          | GetUserMe           |
| POST   | `/api/v1/ride-requests`     | CreateRideRequest   |
| GET    | `/api/v1/ride-requests/{id}`| GetRideRequest      |

### Frontend (`frontend/src/`)
- `main.jsx` — React root; wraps app in `AuthProvider` (react-oidc-context) + `BrowserRouter`; routes `/callback` and `/*`
- `App.jsx` — auth gate, profile-completion gate, main app UI with `RideRequestForm`
- `ProfileCompletionForm.jsx` — shown when `profile_complete` is false; POSTs to `/api/v1/users`
- `RideRequestForm.jsx` — ride request submission form
- `pages/CallbackPage.jsx` — handles OIDC redirect callback, then navigates to `/`
- `auth/splitTokenStore.js` — custom OIDC token store: access/id tokens in sessionStorage, refresh token in httpOnly-style cookie

### Auth / profile gate
After login, `App.jsx` calls `GET /api/v1/users/me?sub=<oidc_sub>`. A 404 or `profile_complete: false` response renders `ProfileCompletionForm` instead of the main UI. On completion, the form POSTs to `POST /api/v1/users` (upsert), then re-checks the profile status.

### Database migrations
Goose manages migrations via embedded SQL in `db/migrations/`. Migrations run automatically when `db.New()` is called at startup. File naming: `NNN_description.sql`.

### Testing
- Handler tests: `handlers/*_test.go` — mock the repository interface, no real DB
- DB tests: `db/*_test.go` — mock the `querier` interface (`pgx.Row`), no `DATABASE_URL` required
- Validator uses json field names (registered via `RegisterTagNameFunc`)

## OpenSpec Workflow

This project uses the OpenSpec artifact workflow for structured development. Changes live in `openspec/changes/<change-name>/` with sequential artifacts: `proposal.md` → `design.md` → `tasks.md` → `specs/<name>/spec.md`.

Main specs are in `openspec/specs/`. Archived changes are in `openspec/changes/archive/`. Use the `/opsx:*` skills to navigate the workflow (e.g. `/opsx:apply` to implement tasks, `/opsx:verify` before archiving).

No active changes currently — all changes are archived.
