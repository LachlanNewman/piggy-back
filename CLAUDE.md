# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Stack

Full-stack web app: Next.js 16 (App Router) + React 19 (frontend), Go 1.23 standard library (backend), PostgreSQL 17 with pgx/v5, Keycloak 26 (OIDC auth).

The product is a piggyback rideshare: every signed-in user is both a rider and a carrier. Riders find people nearby with a free back, send a pickup/drop-off request, and the carrier accepts or declines.

## Commands

### Full stack (Docker)
```bash
docker compose up           # Start all services
docker compose up --build   # Rebuild images then start
docker compose watch        # Start with live reload (backend rebuild on change, frontend sync)
```

### Frontend (`frontend/`)
```bash
npm run dev        # Next dev server on port 3000
npm run build      # Production build
npm run start      # Serve the production build
npm run typecheck  # tsc --noEmit
```

Requires a `frontend/.env.local` file — see `frontend/.env.example` for required vars (`NEXT_PUBLIC_OIDC_AUTHORITY`, `NEXT_PUBLIC_OIDC_CLIENT_ID`, `NEXT_PUBLIC_OIDC_REDIRECT_URI`, `BACKEND_URL`).

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
| frontend          | 3000 | Next.js dev server                  |
| keycloak-postgres | —    | Keycloak's own database (internal)  |
| keycloak          | 8180 | OIDC identity provider              |

### Request flow
Browser → Next.js (rewrites `/api/:path*` to `BACKEND_URL`) → Backend (Go HTTP) → PostgreSQL (pgx pool)

The Next app has no API routes of its own; `next.config.ts` rewrites `/api/:path*` to the Go backend so the browser stays same-origin.

### Realtime updates
Ride changes reach the other party over Server-Sent Events rather than polling.
SSE over WebSockets because the traffic is one-directional (every client action
is already a REST call), `EventSource` reconnects on its own, and it needs
nothing outside `net/http`.

- `events/hub.go` — in-process pub/sub keyed by auth subject. `Publish` is
  non-blocking, so a stalled client can never hold up an HTTP handler.
- `handlers/events.go` — `GET /api/v1/events?sub=...` holds the stream open,
  sends a `connected` handshake and a heartbeat comment every
  `EVENT_HEARTBEAT_SECONDS`.
- Events are change *signals* (`{type, id, status}`), not data. The client
  refetches from the REST API on receipt, so response shapes have one source
  of truth.
- `CreateRideRequest` notifies the carrier; accept/decline notify the rider.

**`Cache-Control: no-cache, no-transform` on the stream is load-bearing.** Next
compresses proxied responses, and a gzipping intermediary buffers the stream
until its window fills — events then arrive minutes late or not at all. The
`no-transform` directive is what stops it. `X-Accel-Buffering: no` covers nginx.
`statusRecorder` in `main.go` must also keep its `Flush` method (the SSE handler
asserts `http.Flusher` through it) and its capped log buffer (an unbounded one
would grow for the life of every stream).

The hub is in-process: it fans out only to subscribers on this backend
instance. More than one replica would need a shared broker — Postgres
LISTEN/NOTIFY behind the same `Publish`/`Subscribe` surface.

On the client, `lib/realtime.ts` owns the `EventSource` and exposes the
connection state; components poll **only** while the stream is not live
(`connection !== 'live'`), so polling is a fallback rather than the mechanism.
Ride expiry is settled client-side against `expires_at` (`lib/time.ts`) — the
server only marks it lazily on read, so the countdown is what ends a request on
time. Browser notifications (`lib/notifications.ts`) are opt-in, requested from
a click, and suppressed while the tab is visible.

Auth flow: Frontend → Keycloak (OIDC code flow, port 8180) → tokens stored via `SplitTokenStore` (refresh token in a cookie, access/id tokens in memory) → `react-oidc-context` provides auth state. Auth is client-only: `app/providers.tsx` mounts `AuthProvider` after hydration because `oidc-client-ts` touches browser storage on construction.

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
- `handlers/events.go` — SSE stream endpoint
- `events/hub.go` — in-process pub/sub fanning changes out to open streams

No framework (no Gin/Echo/Chi); uses `net/http` only.

### API routes
| Method | Path                        | Handler            |
|--------|-----------------------------|--------------------|
| GET    | `/api/health`               | inline health check |
| GET    | `/api/hello`                | inline hello        |
| GET    | `/api/v1/events`            | Events (SSE stream) |
| POST   | `/api/v1/users`             | CreateUser          |
| GET    | `/api/v1/users/me`          | GetUserMe           |
| GET    | `/api/v1/users/nearby`      | GetNearbyUsers      |
| POST   | `/api/v1/location`          | PushLocation        |
| POST   | `/api/v1/ride-requests`     | CreateRideRequest   |
| GET    | `/api/v1/ride-requests/incoming` | GetIncomingRequests |
| GET    | `/api/v1/ride-requests/{id}`| GetRideRequest      |
| PATCH  | `/api/v1/ride-requests/{id}/accept`  | AcceptRideRequest  |
| PATCH  | `/api/v1/ride-requests/{id}/decline` | DeclineRideRequest |

### Frontend (`frontend/src/`)
- `app/layout.tsx` — root layout; metadata, `globals.css`, wraps children in `Providers`
- `app/providers.tsx` — client component; mounts `AuthProvider` (react-oidc-context) after hydration
- `app/page.tsx` — client component; auth gate → profile-completion gate → ride app (location push, incoming requests, nearby carriers, ride flow)
- `app/callback/page.tsx` — OIDC redirect callback, then `router.replace('/')`
- `app/globals.css` — the design system (tokens, cards, buttons, forms, light/dark)
- `components/ProfileCompletionForm.tsx` — shown when `profile_complete` is false; POSTs to `/api/v1/users`
- `components/NearbyCarriers.tsx` — nearby users available to carry; exports the `Carrier` type
- `components/RideRequestFlow.tsx` — pickup/drop-off form, then follows the request live until accepted/declined/expired
- `components/IncomingRequests.tsx` — requests addressed to you; accept/decline, live via SSE
- `lib/realtime.ts` — `EventSource` lifecycle and connection state
- `lib/notifications.ts` — opt-in browser notifications
- `lib/time.ts` — shared clock, countdowns, client-side expiry
- `lib/api/client.ts` — typed backend client; zod-validated responses, `ApiError` for non-2xx
- `lib/auth/splitTokenStore.ts` — custom OIDC token store: access/id tokens in memory, refresh token in a cookie
- `lib/format.ts` — small display helpers

Routing is the App Router (no react-router). All product UI is client-rendered — the Go backend is the only data source.

### Auth / profile gate
After login, `app/page.tsx` calls `GET /api/v1/users/me?sub=<oidc_sub>`. A 404 or `profile_complete: false` response renders `ProfileCompletionForm` instead of the main UI. On completion, the form POSTs to `POST /api/v1/users` (upsert), then re-checks the profile status.

### Database migrations
Goose manages migrations via embedded SQL in `db/migrations/`. Migrations run automatically when `db.New()` is called at startup. File naming: `NNN_description.sql`.

### Testing
- Handler tests: `handlers/*_test.go` — mock the repository interface, no real DB
- Realtime: `events/hub_test.go` (concurrency, run with `-race`),
  `handlers/events_test.go` (stream framing and headers), and
  `handlers/realtime_integration_test.go` (create → notify → accept → notify
  over a real `httptest` server)
- DB tests: `db/*_test.go` — mock the `querier` interface (`pgx.Row`), no `DATABASE_URL` required
- Validator uses json field names (registered via `RegisterTagNameFunc`)

## OpenSpec Workflow

This project uses the OpenSpec artifact workflow for structured development. Changes live in `openspec/changes/<change-name>/` with sequential artifacts: `proposal.md` → `design.md` → `tasks.md` → `specs/<name>/spec.md`.

Main specs are in `openspec/specs/`. Archived changes are in `openspec/changes/archive/`. Use the `/opsx:*` skills to navigate the workflow (e.g. `/opsx:apply` to implement tasks, `/opsx:verify` before archiving).

No active changes currently — all changes are archived.
