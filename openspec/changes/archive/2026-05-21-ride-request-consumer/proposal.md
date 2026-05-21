## Why

The backend currently has no way to receive ride requests asynchronously. A dedicated worker service reads from a RabbitMQ queue and persists ride requests to PostgreSQL, keeping message consumption decoupled from the HTTP API server.

## What Changes

- Add a `worker` binary inside the existing `backend/` Go module, sharing the `db` and `config` packages
- The worker reads messages from a durable `ride-requests` queue, persists each to the `ride_requests` table, and acks/nacks appropriately
- Define a message schema for ride request payloads: `request_id` (UUID, idempotency key), `rider_id`, `pickup` and `dropoff` objects (each with `lat`, `lng`, `address`), `requested_at` (RFC3339)
- Add the worker as a new service in Docker Compose, depending on both `rabbitmq` and `postgres`
- Add a `ride_requests` table migration, applied by the worker on startup

## Capabilities

### New Capabilities

- `ride-request-consumer`: Standalone worker service that reads ride request messages from RabbitMQ and persists them to PostgreSQL

### Modified Capabilities

<!-- none -->

## Impact

- **Module restructure**: `backend/main.go` → `backend/cmd/api/main.go`; worker at `backend/cmd/worker/main.go`; shared packages (`db/`, `config/`) unchanged
- **New package**: `backend/consumer/` — AMQP consumer logic, imported by the worker binary
- **Database**: New `ride_requests` migration added to `backend/db/migrations/`, applied by the API on startup
- **Dependencies**: `github.com/rabbitmq/amqp091-go` added to the existing `backend/go.mod`
- **Docker Compose**: New `worker` service built from the same `backend/` context targeting `./cmd/worker`; no new `go.mod`
