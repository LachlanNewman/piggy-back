## Why

Users currently can submit ride requests into a void — there is no way to find or connect with other users who could give or receive a ride. This feature closes that gap by letting users share their real-time location, discover nearby users, and send directed ride requests with a configurable expiry.

## What Changes

- Users can push their current GPS location to the backend via periodic HTTP polling
- Users can retrieve a list of nearby users (name only, no coordinates exposed) within a configurable radius
- Ride requests become directed — a rider selects a specific nearby user to request a ride from
- A rider may only have one active (pending) request at a time
- A driver can accept or explicitly decline an incoming request; requests also expire automatically after a configurable TTL
- ride_requests gains two new statuses: `declined` and `expired`

## Capabilities

### New Capabilities

- `user-location`: User pushes GPS coordinates; backend stores with a timestamp. Stale entries (older than 2× poll interval) are excluded from nearby queries.
- `nearby-users`: Returns a list of users (id, first_name, last_name) within a configurable radius who have a fresh location. No coordinates are exposed to clients.
- `directed-ride-request`: A rider selects a specific nearby user and sends a ride request. Enforces one active request per rider. Driver can accept or decline; request auto-expires at TTL.

### Modified Capabilities

- `ride-request`: Adds `driver_id` (the target user's auth_subject), `expires_at`, and two new status values (`declined`, `expired`). Enforces the one-active-request constraint.

## Impact

- **New table**: `user_locations` (user_id, lat, lng, updated_at)
- **Altered table**: `ride_requests` — add `driver_id TEXT`, `expires_at TIMESTAMPTZ`; extend status check constraint to include `declined` and `expired`
- **New config env vars**: `LOCATION_POLL_INTERVAL_SECONDS`, `NEARBY_RADIUS_KM` (server-side cap enforced), `RIDE_REQUEST_TTL_MINUTES`
- **New backend handlers**: location push, nearby users list, incoming requests, accept, decline
- **New frontend components**: nearby users list view, incoming request notification panel
- **No new dependencies** — Haversine distance calculated in plain SQL; no PostGIS required
