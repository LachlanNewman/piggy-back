## Why

Riders need a way to submit ride requests from the frontend, and drivers need to accept them. The existing RabbitMQ/worker architecture was designed for automated message processing, which doesn't fit a flow that requires a human driver to manually accept each request.

## What Changes

- **New**: `POST /api/v1/ride-requests` — rider submits a ride request (pickup address, dropoff address); backend writes directly to `ride_requests` with `status=pending`
- **New**: `GET /api/v1/ride-requests/:id` — returns current status of a ride request; used by rider frontend to poll for acceptance
- **New**: Migration adding a `status` column (`pending`, `accepted`) to the `ride_requests` table
- **New**: Frontend ride request form (pickup + dropoff as text addresses) and polling loop that shows the rider when their request is accepted
- **BREAKING**: Remove the RabbitMQ worker architecture — `consumer/`, `cmd/worker/`, and RabbitMQ from `docker-compose.yml` are deleted

## Capabilities

### New Capabilities
- `ride-request`: Rider submits a ride request via a frontend form; backend writes directly to the database; rider polls for driver acceptance

### Modified Capabilities
- `ride-request-consumer`: Retired — the worker and queue-based processing requirements no longer apply; replaced by the direct-write `ride-request` capability

## Impact

- `backend/consumer/` — deleted
- `backend/cmd/worker/` — deleted
- `backend/cmd/api/main.go` — two new route registrations
- `backend/db/migrations/` — new migration for `status` column
- `docker-compose.yml` — RabbitMQ service removed
- `frontend/src/` — new ride request form component and polling logic
