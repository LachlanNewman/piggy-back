## REMOVED Requirements

### Requirement: Worker runs as a standalone service
**Reason**: The worker architecture is retired. Ride requests are now written directly to the database via the backend API. There is no queue to consume.
**Migration**: Delete `backend/consumer/` and `backend/cmd/worker/`. Remove the RabbitMQ service from `docker-compose.yml`.

### Requirement: Worker connects to RabbitMQ on startup
**Reason**: RabbitMQ is removed from the stack. The backend API writes directly to PostgreSQL.
**Migration**: Remove `RABBITMQ_URL` from all environment configurations.

### Requirement: Worker declares a durable ride-requests queue
**Reason**: No queue exists in the new architecture.
**Migration**: No migration needed — queue was never persisted to SQL schema.

### Requirement: Worker processes valid ride request messages
**Reason**: Messages are no longer published to a queue. The `POST /api/v1/ride-requests` endpoint handles validation and inserts directly.
**Migration**: See `ride-request` capability spec for the new API contract.

### Requirement: Worker rejects malformed messages without requeue
**Reason**: Validation now happens in the HTTP handler, not a queue consumer.
**Migration**: None.

### Requirement: Worker requeues messages on transient database errors
**Reason**: HTTP requests return errors synchronously. Retry is the client's responsibility.
**Migration**: None.

### Requirement: Worker shuts down gracefully
**Reason**: Worker binary is deleted.
**Migration**: None.

### Requirement: ride_requests table stores ride request records
**Reason**: Table is retained but ownership moves to the backend API. Schema is extended (status column, nullable lat/lng) via migration 004.
**Migration**: See `ride-request` capability spec for updated schema requirements.
