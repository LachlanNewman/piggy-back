### Requirement: Worker runs as a standalone service
The worker SHALL be a separate Go binary in a `worker/` directory with its own `go.mod` and Dockerfile, independent of the `backend/` module.

#### Scenario: Worker starts independently
- **WHEN** the worker container starts
- **THEN** it connects to RabbitMQ and PostgreSQL without requiring the backend service to be running

### Requirement: Worker connects to RabbitMQ on startup
The worker SHALL connect to RabbitMQ using the `RABBITMQ_URL` environment variable. If the connection fails or the variable is missing, the process SHALL exit with a fatal log message.

#### Scenario: Successful connection
- **WHEN** `RABBITMQ_URL` is set to a valid RabbitMQ address
- **THEN** the worker connects and begins consuming from the `ride-requests` queue

#### Scenario: Missing RABBITMQ_URL
- **WHEN** `RABBITMQ_URL` is not set
- **THEN** the process exits with a fatal error indicating the missing variable

### Requirement: Worker declares a durable ride-requests queue
The worker SHALL declare a durable queue named `ride-requests` before consuming. Declaring a queue that already exists with the same parameters SHALL be a no-op.

#### Scenario: Queue does not exist
- **WHEN** the worker starts and the queue does not exist
- **THEN** the queue is created as durable

#### Scenario: Queue already exists
- **WHEN** the worker starts and the queue already exists as durable
- **THEN** the worker proceeds without error

### Requirement: Worker processes valid ride request messages
The worker SHALL parse each message body as JSON with the following schema:
- `request_id` (UUID string, required) — publisher-generated idempotency key
- `rider_id` (UUID string, required)
- `pickup` (object, required) — `lat` (float64), `lng` (float64), `address` (string)
- `dropoff` (object, required) — `lat` (float64), `lng` (float64), `address` (string)
- `requested_at` (RFC3339 string, required)

On success, the record SHALL be inserted into `ride_requests` and the message SHALL be acknowledged. If a record with the same `request_id` already exists, the insert SHALL be a no-op and the message SHALL still be acknowledged (idempotent delivery).

#### Scenario: Valid message received
- **WHEN** a valid JSON ride request message arrives on the queue
- **THEN** the record is inserted into `ride_requests` and the message is acked

#### Scenario: Duplicate request_id
- **WHEN** a message arrives with a `request_id` that already exists in `ride_requests`
- **THEN** the insert is skipped, the message is acked, and no error is logged

### Requirement: Worker rejects malformed messages without requeue
The worker SHALL nack a message without requeue if the body cannot be parsed as a valid ride request JSON payload or any required field is missing.

#### Scenario: Invalid JSON
- **WHEN** a message body is not valid JSON
- **THEN** the message is nacked with requeue=false and a log entry is written

#### Scenario: Missing required fields
- **WHEN** a message is valid JSON but missing one or more required fields (including nested `lat`, `lng`, or `address`)
- **THEN** the message is nacked with requeue=false and a log entry is written

### Requirement: Worker requeues messages on transient database errors
The worker SHALL nack a message with requeue=true if the database insert fails.

#### Scenario: Database insert fails
- **WHEN** a valid message is received but the database insert returns an error
- **THEN** the message is nacked with requeue=true and a log entry is written

### Requirement: Worker shuts down gracefully
The worker SHALL stop consuming and close the AMQP connection when the process receives SIGINT or SIGTERM.

#### Scenario: SIGINT received
- **WHEN** the process receives SIGINT
- **THEN** the consumer stops, the AMQP connection is closed, and the process exits cleanly

### Requirement: ride_requests table stores ride request records
The database SHALL have a `ride_requests` table with the following columns:
- `id` (UUID, primary key, default gen_random_uuid())
- `request_id` (UUID, not null, unique) — idempotency key from the message
- `rider_id` (text, not null)
- `pickup_lat` (double precision, not null)
- `pickup_lng` (double precision, not null)
- `pickup_address` (text, not null)
- `dropoff_lat` (double precision, not null)
- `dropoff_lng` (double precision, not null)
- `dropoff_address` (text, not null)
- `requested_at` (timestamptz, not null)
- `created_at` (timestamptz, not null, default now())

This migration SHALL be applied by the backend on startup (added to `backend/db/migrations/`). The worker SHALL NOT run migrations.

#### Scenario: Record inserted
- **WHEN** a valid ride request is processed
- **THEN** a row exists in `ride_requests` with all fields populated and `created_at` set to the insert time

#### Scenario: Unique constraint on request_id
- **WHEN** an insert is attempted with a `request_id` that already exists
- **THEN** the insert is rejected by the unique constraint and the worker treats it as a no-op
