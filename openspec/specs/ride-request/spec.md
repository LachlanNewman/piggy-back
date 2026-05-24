## Purpose

The ride-request capability provides the API and frontend for riders to submit ride requests directly to the backend and poll for driver acceptance. Requests are written directly to PostgreSQL via the backend API — no message queue is involved.

## Requirements

### Requirement: Rider can submit a ride request
The system SHALL accept `POST /api/v1/ride-requests` with a JSON body containing `pickup_address` (string, required), `dropoff_address` (string, required), and a `?sub=` query parameter identifying the rider. On success it SHALL insert a row into `ride_requests` with `status=pending` and return `201` with `{"id": "<uuid>"}`.

#### Scenario: Valid submission
- **WHEN** an authenticated rider POSTs `{"pickup_address": "123 Main St", "dropoff_address": "456 Oak Ave"}` with `?sub=<their-sub>`
- **THEN** a row is inserted into `ride_requests` with `status=pending`, `rider_id=<sub>`, and the response is `201 {"id": "<uuid>"}`

#### Scenario: Missing pickup_address
- **WHEN** the request body omits `pickup_address`
- **THEN** the response is `400 {"error": "pickup_address is required"}`

#### Scenario: Missing dropoff_address
- **WHEN** the request body omits `dropoff_address`
- **THEN** the response is `400 {"error": "dropoff_address is required"}`

#### Scenario: Missing sub query param
- **WHEN** the request is made without a `?sub=` query parameter
- **THEN** the response is `400 {"error": "sub is required"}`

### Requirement: Rider can poll for request status
The system SHALL accept `GET /api/v1/ride-requests/:id` and return the current status of the ride request. The response SHALL include `id`, `status`, `pickup_address`, `dropoff_address`, and `requested_at`.

#### Scenario: Request exists and is pending
- **WHEN** a GET is made for a request with `status=pending`
- **THEN** the response is `200 {"id": "...", "status": "pending", "pickup_address": "...", "dropoff_address": "...", "requested_at": "..."}`

#### Scenario: Request exists and is accepted
- **WHEN** a GET is made for a request with `status=accepted`
- **THEN** the response is `200` with `"status": "accepted"`

#### Scenario: Request not found
- **WHEN** a GET is made for an ID that does not exist
- **THEN** the response is `404 {"error": "not found"}`

### Requirement: ride_requests table has a status column
The `ride_requests` table SHALL have a `status` column of type `TEXT NOT NULL DEFAULT 'pending'` with a CHECK constraint allowing only `pending` and `accepted`. The `pickup_lat`, `pickup_lng`, `dropoff_lat`, and `dropoff_lng` columns SHALL be made nullable.

#### Scenario: New ride request row has status pending
- **WHEN** a ride request is inserted without specifying status
- **THEN** the row has `status = 'pending'`

#### Scenario: Status constraint enforced
- **WHEN** an insert or update attempts to set `status` to a value other than `pending` or `accepted`
- **THEN** the database rejects the operation with a constraint violation

### Requirement: Frontend ride request form
The frontend SHALL render a ride request form when the rider is authenticated and their profile is complete. The form SHALL have a pickup address text field, a dropoff address text field, and a submit button. On submit, the frontend SHALL POST to `/api/v1/ride-requests` and transition to a polling state.

#### Scenario: Successful form submission
- **WHEN** a rider fills in both fields and clicks submit
- **THEN** the frontend POSTs the request and displays a "waiting for driver" message

#### Scenario: Empty field on submit
- **WHEN** a rider submits the form with one or both fields empty
- **THEN** the form does not submit and indicates the missing field

### Requirement: Frontend polls for driver acceptance
After submitting a ride request, the frontend SHALL poll `GET /api/v1/ride-requests/:id` every 3 seconds. When the status becomes `accepted`, the frontend SHALL stop polling and display a confirmation to the rider.

#### Scenario: Driver accepts while rider is waiting
- **WHEN** the ride request status changes to `accepted`
- **THEN** the frontend stops polling and shows "Your driver is on the way!"

#### Scenario: Polling continues while pending
- **WHEN** the ride request status is still `pending`
- **THEN** the frontend continues polling every 3 seconds
