## 1. Config

- [x] 1.1 Add `LOCATION_POLL_INTERVAL_SECONDS` (int, default 30), `NEARBY_RADIUS_KM` (float64, default 5), and `RIDE_REQUEST_TTL_MINUTES` (int, default 15) to `backend/config/config.go`

## 2. Database Migrations

- [x] 2.1 Create `backend/db/migrations/005_add_user_locations.sql` — `user_locations` table with `user_id TEXT NOT NULL UNIQUE`, `lat`, `lng`, `updated_at`
- [x] 2.2 Create `backend/db/migrations/006_extend_ride_requests_for_directed.sql` — add `driver_id TEXT NOT NULL DEFAULT ''`, `expires_at TIMESTAMPTZ NOT NULL DEFAULT now()`; extend status CHECK constraint to include `declined` and `expired`

## 3. DB Layer — user_locations

- [x] 3.1 Add `UpsertUserLocation(ctx, sub string, lat, lng float64) error` to `backend/db/users.go` (upsert on `user_id`)
- [x] 3.2 Add `GetNearbyUsers(ctx, sub string, lat, lng, radiusKm float64, staleThreshold time.Duration) ([]NearbyUser, error)` using Haversine SQL; exclude requesting user; exclude stale rows
- [x] 3.3 Write unit tests for `UpsertUserLocation` and `GetNearbyUsers` using the mock querier pattern

## 4. DB Layer — directed ride requests

- [x] 4.1 Update `CreateRideRequest` in `backend/db/ride_requests.go` to accept `DriverID string` and `ExpiresAt time.Time` in `CreateRideRequestParams`
- [x] 4.2 Add `HasActivePendingRequest(ctx, riderID string) (bool, error)` — returns true if a non-expired pending request exists for the rider
- [x] 4.3 Add `GetIncomingRequests(ctx, driverID string) ([]RideRequest, error)` — pending, non-expired requests where `driver_id = driverID`
- [x] 4.4 Add `AcceptRideRequest(ctx, id, driverID string) error` — validates ownership, status, expiry; sets `status=accepted`
- [x] 4.5 Add `DeclineRideRequest(ctx, id, driverID string) error` — same pre-conditions; sets `status=declined`
- [x] 4.6 Update `GetRideRequestByID` to include `expires_at` in the returned struct; return virtual status `expired` when `expires_at < now()` and status is `pending`
- [x] 4.7 Write unit tests for all new/updated DB methods

## 5. Backend Handlers

- [x] 5.1 Create `backend/handlers/location.go` — `PushLocation` handler for `POST /api/v1/location`; validate `sub`, `lat`, `lng`; call `UpsertUserLocation`; return 204
- [x] 5.2 Create `backend/handlers/nearby_users.go` — `GetNearbyUsers` handler for `GET /api/v1/users/nearby`; validate `sub`; look up requester's location; call `GetNearbyUsers`; cap radius at 20 km; return JSON array
- [x] 5.3 Update `backend/handlers/ride_requests.go` — `CreateRideRequest` handler: add `driver_id` to request body; check `HasActivePendingRequest` (return 409 if true); pass `ExpiresAt` and `DriverID` to DB layer
- [x] 5.4 Create `backend/handlers/ride_requests_incoming.go` — `GetIncomingRequests` handler for `GET /api/v1/ride-requests/incoming`
- [x] 5.5 Add `AcceptRideRequest` handler for `PATCH /api/v1/ride-requests/{id}/accept`
- [x] 5.6 Add `DeclineRideRequest` handler for `PATCH /api/v1/ride-requests/{id}/decline`
- [x] 5.7 Write handler unit tests for all new handlers (mock interfaces, no DB)

## 6. Route Registration

- [x] 6.1 Register all new routes in `backend/cmd/api/main.go`: `POST /api/v1/location`, `GET /api/v1/users/nearby`, `GET /api/v1/ride-requests/incoming`, `PATCH /api/v1/ride-requests/{id}/accept`, `PATCH /api/v1/ride-requests/{id}/decline`

## 7. Frontend — Location Polling

- [x] 7.1 Add a `useLocationPoller` hook (or effect in `App.jsx`) that requests browser GPS permission, then POSTs to `POST /api/v1/location` every `LOCATION_POLL_INTERVAL_SECONDS` seconds while the user is authenticated and profile is complete
- [x] 7.2 Handle GPS permission denied gracefully — show a prompt explaining location is required for nearby matching

## 8. Frontend — Nearby Users

- [x] 8.1 Create `frontend/src/NearbyUsersList.jsx` — fetches `GET /api/v1/users/nearby` on mount and on a "Refresh" button; displays list of names with "Request ride" button per user; handles empty state and 404 (no location yet)
- [x] 8.2 Wire "Request ride" button to pre-fill `driver_id` in the ride request flow and submit `POST /api/v1/ride-requests`
- [x] 8.3 Show 409 error ("you already have an active request") inline if rider tries to request while one is pending

## 9. Frontend — Incoming Requests (Driver Side)

- [x] 9.1 Create `frontend/src/IncomingRequests.jsx` — polls `GET /api/v1/ride-requests/incoming` on the same interval as location; displays pending requests with Accept / Decline buttons
- [x] 9.2 Accept button calls `PATCH /api/v1/ride-requests/{id}/accept`; on success removes from panel
- [x] 9.3 Decline button calls `PATCH /api/v1/ride-requests/{id}/decline`; on success removes from panel

## 10. Frontend — Rider Status Polling

- [x] 10.1 After a ride request is submitted, poll `GET /api/v1/ride-requests/{id}` on the same interval until status is `accepted`, `declined`, or `expired`
- [x] 10.2 Show appropriate terminal message for each final status and stop polling

## 11. Integration

- [x] 11.1 Wire `NearbyUsersList` and `IncomingRequests` into `App.jsx` so both are visible when profile is complete
- [ ] 11.2 Manual end-to-end test: two browser sessions — one rider, one driver — verify full request/accept flow
- [ ] 11.3 Manual test: verify request expires correctly after TTL
- [ ] 11.4 Manual test: verify declined request shows correct message to rider
