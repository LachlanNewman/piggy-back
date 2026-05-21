## 1. Module Restructure

- [x] 1.1 Move `backend/main.go` to `backend/cmd/api/main.go` (package main, all imports unchanged)
- [x] 1.2 Update `backend/Dockerfile` to build `./cmd/api` instead of `.`
- [x] 1.3 Add `github.com/rabbitmq/amqp091-go` to `backend/go.mod` and run `go mod tidy`

## 2. Database Migration (backend)

- [x] 2.1 Create `backend/db/migrations/002_create_ride_requests.sql` with the `ride_requests` table: `id` (UUID PK), `request_id` (UUID unique not null), `rider_id` (text not null), `pickup_lat`/`pickup_lng` (double precision not null), `pickup_address` (text not null), `dropoff_lat`/`dropoff_lng` (double precision not null), `dropoff_address` (text not null), `requested_at` (timestamptz not null), `created_at` (timestamptz not null default now())
- [x] 2.2 Embed and apply the new migration in `backend/db/db.go` alongside `001_create_users.sql`

## 3. Consumer Package

- [x] 3.1 Create `backend/consumer/consumer.go` with a `Consumer` struct holding the AMQP connection, channel, and pgx pool
- [x] 3.2 Implement `New(amqpURL string, pool *pgxpool.Pool)` that dials RabbitMQ, opens a channel, and declares the durable `ride-requests` queue
- [x] 3.3 Define `RideRequestMessage` struct with `RequestID` (string), `RiderID` (string), `Pickup`/`Dropoff` (nested struct with `Lat`, `Lng`, `Address`), `RequestedAt` (string); validate all required fields are non-zero after unmarshal
- [x] 3.4 Implement `Consume(ctx context.Context)` that ranges over deliveries, parses each as `RideRequestMessage`, inserts into `ride_requests` using `INSERT ... ON CONFLICT (request_id) DO NOTHING`, and acks/nacks per spec
- [x] 3.5 Implement `Close()` that closes the AMQP channel and connection

## 4. Worker Entrypoint

- [x] 4.1 Create `backend/cmd/worker/main.go` that initialises the DB pool via `db.New(ctx)` (fatal on error), reads `RABBITMQ_URL` from env (fatal if missing), initialises `consumer.New(...)`, sets up OS signal handling (SIGINT/SIGTERM) with `context.WithCancel`, starts `Consume` in a goroutine, and calls `Close()` on shutdown

## 5. Docker Compose

- [x] 5.1 Add a healthcheck to the `backend` service in `docker-compose.yml` (curl or wget `http://localhost:8080/api/health`)
- [x] 5.2 Add a `worker` service with `build: ./backend`, a `command` or build arg targeting `./cmd/worker`, `DATABASE_URL`, `RABBITMQ_URL`, and `depends_on: backend (healthy) + rabbitmq (healthy)`

## 6. Verification

- [ ] 6.1 Run `docker-compose up --build` and confirm all services start without errors
- [ ] 6.2 Publish a valid test message to `ride-requests` via the RabbitMQ management UI and confirm a row appears in `ride_requests`
- [ ] 6.3 Publish a malformed message and confirm it is nacked without requeue (queue depth stays 0)
