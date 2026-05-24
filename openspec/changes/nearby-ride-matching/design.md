## Context

The backend is a Go 1.23 `net/http` service with no framework. Data lives in PostgreSQL 17 accessed via `pgx/v5`. Config is env-parsed via `caarlos0/env`. The frontend is React 18 + Vite using `react-oidc-context` for auth. All communication is REST over HTTP — no WebSockets or message broker.

Currently `ride_requests` are directionless: `rider_id` is set but there is no target driver, no expiry, and no way to find who to request from. Users have no stored location.

## Goals / Non-Goals

**Goals:**
- Let users push their GPS location via periodic HTTP polling
- Let users discover nearby users (names only) within a configurable radius
- Let a rider send a directed ride request to one specific nearby user
- Let a driver accept or decline; requests also expire after a configurable TTL
- Enforce one active request per rider at a time

**Non-Goals:**
- Real-time push / WebSockets / SSE
- Route matching (A→B overlap)
- Driver/rider role distinction — all users can do both
- Map UI or displaying other users' coordinates to clients
- PostGIS — plain Haversine SQL is sufficient

## Decisions

### Location storage: separate `user_locations` table

Storing live location in a dedicated table (rather than columns on `users`) keeps ephemeral data separate from the persistent user profile. It also makes it easy to drop or truncate stale rows later without touching user records.

```
user_locations
  user_id     TEXT NOT NULL UNIQUE  ← auth_subject (matches users.auth_subject)
  lat         DOUBLE PRECISION NOT NULL
  lng         DOUBLE PRECISION NOT NULL
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
```

`UNIQUE` on `user_id` means each user has exactly one live location row, updated via `ON CONFLICT DO UPDATE`.

### Proximity query: Haversine in plain SQL, no PostGIS

At the scale of a local ride-sharing app, a Haversine formula embedded in a WHERE clause performs adequately. PostGIS adds an extension dependency and operational complexity not warranted here.

The stale threshold is derived — not configured separately — as `2 × LOCATION_POLL_INTERVAL_SECONDS`. This means if a user stops polling (closes the app) they automatically disappear from nearby results within two poll cycles.

### Directed requests: `driver_id` + `expires_at` on `ride_requests`

Rather than a separate table, the existing `ride_requests` table is extended with:
- `driver_id TEXT NOT NULL` — auth_subject of the target user
- `expires_at TIMESTAMPTZ NOT NULL` — set at insert time as `now() + TTL interval`

Status check constraint is extended to `('pending', 'accepted', 'declined', 'expired')`.

Expiry is enforced on read (filter `WHERE expires_at > now() AND status = 'pending'`), not by a background job. A cleanup migration can be added later if table size becomes a concern.

### One active request per rider: application-layer check

Before inserting a new ride request, the handler queries for any existing row where `rider_id = $1 AND status = 'pending' AND expires_at > now()`. If one exists, return HTTP 409. This is simpler than a partial unique index and easier to reason about.

### Config: three new env vars

Follows the existing `caarlos0/env` pattern in `config/config.go`:

```
LOCATION_POLL_INTERVAL_SECONDS  int  default: 30
NEARBY_RADIUS_KM                float64  default: 5  (server enforces max: 20)
RIDE_REQUEST_TTL_MINUTES        int  default: 15
```

`NEARBY_RADIUS_KM` is capped server-side at 20 km regardless of what the client sends, preventing abuse.

### Frontend: polling, not push

The client polls `GET /api/v1/ride-requests/incoming` on the same interval as the location push. No persistent connection needed. The nearby list is fetched on demand (user action), not continuously.

## Risks / Trade-offs

- **Location lag**: With a 30s poll interval, a user's position is up to 30s stale. Acceptable for a local ride-share; would not suit a taxi-hailing app. → No mitigation needed at this stage.
- **Expiry on read only**: Expired requests remain in the DB until queried. Under low traffic this is fine; under high traffic the table can grow. → Add a cleanup job or DB-level TTL later if needed.
- **No driver notification push**: Driver must be actively polling to see incoming requests. If the driver closes the app, requests silently expire. → Acceptable given the polling-first approach; push notifications are a future concern.
- **Haversine accuracy**: Haversine slightly underestimates distance near the poles and over large distances. At ≤20 km in inhabited areas the error is negligible. → No mitigation needed.
- **One request at a time per rider**: If the chosen driver is unavailable, the rider must wait for TTL expiry before trying another. → TTL should be kept short (default 15 min). A future improvement could allow cancel-and-retry.

## Migration Plan

1. Deploy migration `005_add_user_locations.sql` — creates `user_locations` table.
2. Deploy migration `006_extend_ride_requests_for_directed.sql` — adds `driver_id`, `expires_at`, extends status constraint. Existing rows get `expires_at = now()` (immediately expired) and `driver_id = ''` as safe defaults.
3. Deploy backend with new handlers and updated config.
4. Deploy frontend with location polling and nearby-users UI.

Rollback: drop the two new migrations and revert backend/frontend. Existing `ride_requests` rows with the old status values remain valid after rollback since the constraint is additive.

## Open Questions

- Should riders be able to cancel a pending request explicitly, or only wait for TTL?
- What should the frontend show while waiting for a driver to respond?
