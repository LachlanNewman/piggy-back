## Context

The backend is an HTTP API service — embedding a long-running consumer in it couples two unrelated responsibilities and complicates scaling and deployment. Running the worker as a separate container lets each service scale, restart, and deploy independently. RabbitMQ is already in Docker Compose; this adds a `worker` service alongside `backend`, `postgres`, and `rabbitmq`.

## Goals / Non-Goals

**Goals:**
- Standalone `worker/` Go module with its own `main.go` and Dockerfile
- Connects to RabbitMQ (`RABBITMQ_URL`) and PostgreSQL (`DATABASE_URL`) on startup
- Declares and consumes from a durable `ride-requests` queue
- Persists ride requests to a `ride_requests` table; acks on success, nacks on error
- Applies the `ride_requests` migration on startup
- Graceful shutdown on SIGINT/SIGTERM
- Added as a `worker` service in Docker Compose

**Non-Goals:**
- Publishing messages (producer is out of scope)
- Dead-letter queue setup
- Retry policies beyond single requeue-on-error
- Sharing code with the `backend/` module (separate binaries, separate go.mod)

## Decisions

**Shared Go module with `cmd/` layout**
The worker lives inside the existing `backend/` module at `backend/cmd/worker/main.go`. The API server moves to `backend/cmd/api/main.go`. Both binaries import `backend/db`, `backend/config`, and (for the worker) a new `backend/consumer` package. One `go.mod`, one `go.sum`, one Dockerfile build context — the target binary is selected by the Docker build arg. This avoids duplicating DB setup, config parsing, and dependency management across two modules.

**Backend runs all migrations; worker connects only**
The backend already applies migrations on startup via `db/db.go`. The `ride_requests` table migration is added there alongside the existing `users` migration. The worker connects to Postgres and uses the pool but does not run migrations — it relies on the backend having already applied the schema. Docker Compose `depends_on` ordering ensures the backend starts (and migrates) before the worker.

**Use `amqp091-go` directly**
No framework — consistent with the backend's stdlib-only philosophy. Keeps the worker small and predictable.

**Ack/Nack strategy: ack on success, nack+requeue on DB error, nack+no-requeue on parse error**
Malformed messages are dropped to prevent infinite loops. DB errors are transient and worth retrying via requeue.

**Graceful shutdown via `os/signal` + `context.WithCancel`**
A single cancel context drives both the AMQP consumer loop and any cleanup. On signal, the consumer exits its delivery loop, closes the AMQP connection, and the process exits cleanly.

## Risks / Trade-offs

- **RabbitMQ or Postgres unavailable at startup** → Process exits with a fatal log. Docker Compose `depends_on` + healthchecks handle ordering in dev. Production should use a retry/backoff loop — left for a follow-up.
- **Module restructure requires moving `backend/main.go`** → Small one-time refactor; all existing packages and imports are unchanged. The Dockerfile needs updating to target `./cmd/api` instead of `.`.
- **Worker depends on backend for schema** → If the worker starts before the backend has applied migrations, inserts will fail. Compose `depends_on: backend (healthy)` mitigates this in dev.
- **Single consumer, prefetch=1** → Limits throughput but ensures ordered, safe processing for the initial implementation.

## Migration Plan

1. Move `backend/main.go` → `backend/cmd/api/main.go`; update `backend/Dockerfile` to build `./cmd/api`
2. Add `ride_requests` migration to `backend/db/migrations/`
3. Add `backend/consumer/` package and `backend/cmd/worker/main.go`
4. Add `worker` service to `docker-compose.yml` (same build context as backend, targets `./cmd/worker`)
5. `docker-compose up --build` to verify all services start and the worker consumes messages

Rollback: remove the `worker` service from `docker-compose.yml`. No changes to the backend or existing migrations.

## Open Questions

- Queue name hardcoded as `ride-requests` for now; can be made configurable via env var later.
