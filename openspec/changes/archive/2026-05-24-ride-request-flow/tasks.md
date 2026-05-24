## 1. Remove Worker Infrastructure

- [x] 1.1 Delete `backend/consumer/` directory
- [x] 1.2 Delete `backend/cmd/worker/` directory
- [x] 1.3 Remove RabbitMQ service from `docker-compose.yml`
- [x] 1.4 Remove any RabbitMQ-related environment variables from `docker-compose.yml`

## 2. Database Migration

- [x] 2.1 Create `backend/db/migrations/004_alter_ride_requests_status.sql` — adds `status TEXT NOT NULL DEFAULT 'pending'` with CHECK constraint `('pending', 'accepted')`, makes `pickup_lat`, `pickup_lng`, `dropoff_lat`, `dropoff_lng` nullable

## 3. Backend — DB Layer

- [x] 3.1 Add `CreateRideRequest(ctx, params)` method to the DB layer — inserts a row with generated `request_id` UUID, `rider_id`, `pickup_address`, `dropoff_address`, `requested_at=now()`, `status=pending`; returns the inserted `id`
- [x] 3.2 Add `GetRideRequestByID(ctx, id)` method — returns `id`, `status`, `pickup_address`, `dropoff_address`, `requested_at` or a not-found sentinel

## 4. Backend — Handlers

- [x] 4.1 Create `backend/handlers/ride_requests.go` with `CreateRideRequest` handler — validates `pickup_address`, `dropoff_address` in body and `sub` query param; calls DB layer; returns `201 {"id": "..."}` on success
- [x] 4.2 Create `GetRideRequest` handler in the same file — parses `:id` from path; calls DB layer; returns `200` with status fields or `404`
- [x] 4.3 Register both routes in `backend/cmd/api/main.go`: `POST /api/v1/ride-requests` and `GET /api/v1/ride-requests/{id}`

## 5. Frontend — Ride Request Form

- [x] 5.1 Create `frontend/src/RideRequestForm.jsx` — pickup address text input, dropoff address text input, submit button; client-side validation that both fields are non-empty
- [x] 5.2 On submit, POST to `/api/v1/ride-requests?sub=<user.profile.sub>` with the address fields; on success, transition to polling state with the returned `id`

## 6. Frontend — Polling

- [x] 6.1 After successful submission, poll `GET /api/v1/ride-requests/:id` every 3 seconds
- [x] 6.2 While polling, show "Waiting for a driver..." message
- [x] 6.3 When status becomes `accepted`, stop polling and show "Your driver is on the way!"

## 7. Frontend — App Integration

- [x] 7.1 Render `RideRequestForm` in `App.jsx` when the rider is authenticated and `profileStatus === 'complete'`
