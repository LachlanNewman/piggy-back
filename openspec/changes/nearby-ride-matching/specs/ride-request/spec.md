## MODIFIED Requirements

### Requirement: Rider can submit a ride request
The system SHALL accept `POST /api/v1/ride-requests` with a JSON body containing `pickup_address` (string, required), `dropoff_address` (string, required), `driver_id` (string, required — the `auth_subject` of the target driver), and a `?sub=` query parameter identifying the rider. On success it SHALL insert a row into `ride_requests` with `status=pending`, `driver_id` set, and `expires_at = now() + RIDE_REQUEST_TTL_MINUTES`, and return `201` with `{"id": "<uuid>"}`. If the rider already has a pending non-expired request, the system SHALL return `409 {"error": "you already have an active request"}`.

#### Scenario: Valid submission
- **WHEN** an authenticated rider POSTs `{"pickup_address": "123 Main St", "dropoff_address": "456 Oak Ave", "driver_id": "<target-sub>"}` with `?sub=<their-sub>` and no active pending request exists
- **THEN** a row is inserted with `status=pending`, `driver_id=<target-sub>`, `expires_at = now() + TTL`, and the response is `201 {"id": "<uuid>"}`

#### Scenario: Missing pickup_address
- **WHEN** the request body omits `pickup_address`
- **THEN** the response is `400 {"error": "pickup_address is required"}`

#### Scenario: Missing dropoff_address
- **WHEN** the request body omits `dropoff_address`
- **THEN** the response is `400 {"error": "dropoff_address is required"}`

#### Scenario: Missing driver_id
- **WHEN** the request body omits `driver_id`
- **THEN** the response is `400 {"error": "driver_id is required"}`

#### Scenario: Missing sub query param
- **WHEN** the request is made without a `?sub=` query parameter
- **THEN** the response is `400 {"error": "sub is required"}`

#### Scenario: Rider already has an active request
- **WHEN** the rider already has a `pending` request with `expires_at > now()`
- **THEN** the response is `409 {"error": "you already have an active request"}`

### Requirement: Rider can poll for request status
The system SHALL accept `GET /api/v1/ride-requests/:id` and return the current status of the ride request. The response SHALL include `id`, `status`, `pickup_address`, `dropoff_address`, `requested_at`, and `expires_at`. The status field SHALL reflect `expired` when `expires_at` is in the past and the status is still `pending`.

#### Scenario: Request exists and is pending and not expired
- **WHEN** a GET is made for a request with `status=pending` and `expires_at > now()`
- **THEN** the response is `200` with `"status": "pending"` and all required fields

#### Scenario: Request exists and is accepted
- **WHEN** a GET is made for a request with `status=accepted`
- **THEN** the response is `200` with `"status": "accepted"`

#### Scenario: Request exists and is declined
- **WHEN** a GET is made for a request with `status=declined`
- **THEN** the response is `200` with `"status": "declined"`

#### Scenario: Request has expired (TTL passed, still pending in DB)
- **WHEN** a GET is made and `expires_at < now()` and `status=pending`
- **THEN** the response is `200` with `"status": "expired"`

#### Scenario: Request not found
- **WHEN** a GET is made for an ID that does not exist
- **THEN** the response is `404 {"error": "not found"}`

### Requirement: ride_requests table supports directed requests with expiry
The `ride_requests` table SHALL have `driver_id TEXT NOT NULL` and `expires_at TIMESTAMPTZ NOT NULL` columns added. The `status` CHECK constraint SHALL be extended to allow `pending`, `accepted`, `declined`, and `expired`. Existing rows SHALL be migrated with `driver_id = ''` and `expires_at = now()` (treating them as immediately expired).

#### Scenario: New ride request row has required fields
- **WHEN** a ride request is inserted
- **THEN** the row has `status = 'pending'`, a non-null `driver_id`, and a non-null `expires_at` in the future

#### Scenario: Extended status constraint enforced
- **WHEN** an insert or update attempts to set `status` to a value outside `pending`, `accepted`, `declined`, `expired`
- **THEN** the database rejects the operation with a constraint violation
