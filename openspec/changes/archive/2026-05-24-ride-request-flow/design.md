## Context

The project has a `ride_requests` table (migration 002) and a worker binary that consumed ride request messages from RabbitMQ and inserted them. The worker architecture assumed automated processing, but the actual flow requires a human driver to accept each request. This design replaces the queue path with a direct API write and adds the frontend rider experience.

## Goals / Non-Goals

**Goals:**
- Rider can submit a ride request (pickup + dropoff addresses) from the frontend
- Backend writes the request directly to `ride_requests` with `status=pending`
- Rider frontend polls for status change to `accepted`
- Remove the RabbitMQ worker — nothing in the system publishes to or consumes from the queue

**Non-Goals:**
- Driver acceptance UI (separate change)
- Geocoding or map-based location input
- Push notifications or WebSockets (polling is sufficient for now)
- Ride cancellation

## Decisions

### Direct DB write over queue
**Decision**: `POST /api/v1/ride-requests` inserts directly into `ride_requests`, no RabbitMQ.

**Rationale**: A queue is appropriate for automated processing pipelines. Here, the record just needs to sit in the DB until a human driver acts on it. Direct writes are simpler, synchronous, and give the rider instant confirmation the request was saved.

**Alternative considered**: Keep the queue path, add a status column, have the worker insert with `status=pending`. Rejected — adds latency before the record exists (making polling awkward) and keeps an infrastructure dependency (RabbitMQ) that serves no purpose.

### lat/lng columns made nullable
**Decision**: Migration 004 makes `pickup_lat`, `pickup_lng`, `dropoff_lat`, `dropoff_lng` nullable.

**Rationale**: The form collects text addresses only. Keeping these NOT NULL would require dummy 0-values which are worse than NULL. Geocoding can be added later without schema changes.

### rider_id from OIDC sub query param
**Decision**: The POST endpoint reads `rider_id` from a `?sub=` query parameter, consistent with the existing `GET /api/v1/users/me?sub=` pattern.

**Rationale**: Auth header validation isn't implemented anywhere in the backend yet. Using the same `?sub=` convention keeps the auth pattern consistent until proper token validation is added.

### Polling interval: 3 seconds
**Decision**: Frontend polls `GET /api/v1/ride-requests/:id` every 3 seconds.

**Rationale**: Fast enough to feel responsive when a driver accepts, slow enough to not hammer the DB. No need for WebSockets given current scale.

### request_id generated server-side
**Decision**: The backend generates `request_id` (UUID) on insert, not the client.

**Rationale**: The client doesn't need to supply an idempotency key for a synchronous API — if the POST fails, the user submits again. Server-side generation is simpler.

## Risks / Trade-offs

- **No auth validation on rider_id** → A caller could supply any sub. Acceptable until token validation middleware is added; matches existing API surface.
- **Polling creates steady DB reads** → At low scale this is fine. If ride volume grows, this needs an index on `id` (already the PK) and potentially a status index.
- **lat/lng nullable** → Existing ride_requests rows (if any) have 0,0 coordinates from the worker. New rows will have NULL. Code reading these columns must handle NULL.

## Migration Plan

1. Deploy migration 004: adds `status` column, makes lat/lng nullable
2. Remove `consumer/`, `cmd/worker/` from the backend module
3. Remove RabbitMQ service from `docker-compose.yml`
4. Deploy new backend with ride request endpoints
5. Deploy frontend with ride request form

Rollback: migration 004 Down drops the status column and restores NOT NULL on lat/lng (safe if no rows have NULL lat/lng, i.e. fresh dev environment).
