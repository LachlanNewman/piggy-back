# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Stack

Full-stack web app: React 18 + Vite (frontend), Go 1.23 standard library (backend), PostgreSQL 17 with pgx/v5.

## Commands

### Full stack (Docker)
```bash
docker-compose up           # Start all services (postgres:5432, backend:8080, frontend:5173)
docker-compose up --build   # Rebuild images then start
```

### Frontend (`frontend/`)
```bash
npm run dev      # Vite dev server on port 5173
npm run build    # Production build
npm run preview  # Preview production build
```

### Backend (`backend/`)
The backend has no standalone run script — it's built via Docker multi-stage. To run locally outside Docker:
```bash
go run main.go   # Requires DATABASE_URL env var pointing to postgres
```

## Architecture

### Request flow
Frontend (React) → `/api/*` → Backend (Go HTTP) → PostgreSQL (pgx pool)

In development via Docker Compose, all three services share a network. The frontend calls the backend directly at `localhost:8080`.

### Backend (`backend/`)
- `main.go` — Single file: CORS middleware, route registration, HTTP server on `:8080`
- `db/db.go` — Exports `New()` which creates a `pgxpool.Pool` from `DATABASE_URL`. This is the only database abstraction layer.
- No framework (no Gin/Echo/Chi); uses `net/http` only. All handlers are registered in `main.go`.

### Frontend (`frontend/src/`)
- `main.jsx` — React DOM root
- `App.jsx` — Main component; fetches `/api/health` and `/api/hello` on mount

### Infrastructure
- Docker Compose orchestrates all three services. Backend waits for postgres health check before starting.
- PostgreSQL credentials and `DATABASE_URL` are set in `docker-compose.yml`.

## OpenSpec Workflow

This project uses the OpenSpec artifact workflow for structured development. Changes live in `openspec/changes/<change-name>/` with sequential artifacts: `proposal.md` → `design.md` → `tasks.md` → `specs/<name>/spec.md`.

Main specs are in `openspec/specs/`. Use the `/opsx:*` skills to navigate the workflow (e.g. `/opsx:apply` to implement tasks, `/opsx:verify` before archiving).

**Active change**: `openspec/changes/user-signup/` — adding `POST /api/users` endpoint, a `users` table, and a `SignupForm` React component.
