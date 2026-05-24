## ADDED Requirements

### Requirement: User can push current GPS location
The system SHALL accept `POST /api/v1/location` with a JSON body containing `lat` (float, required) and `lng` (float, required), and a `?sub=` query parameter identifying the user. On success it SHALL upsert a row in `user_locations` (keyed on `user_id`) and return `204 No Content`.

#### Scenario: Valid location push
- **WHEN** an authenticated user POSTs `{"lat": -33.8688, "lng": 151.2093}` with `?sub=<their-sub>`
- **THEN** the system upserts the row in `user_locations` and returns `204`

#### Scenario: Repeated location push updates the row
- **WHEN** the same user pushes a new location
- **THEN** the existing `user_locations` row is updated in place (upsert on `user_id`) and `updated_at` reflects the new time

#### Scenario: Missing sub query param
- **WHEN** the request is made without a `?sub=` query parameter
- **THEN** the response is `400 {"error": "sub is required"}`

#### Scenario: Missing lat or lng
- **WHEN** the request body omits `lat` or `lng`
- **THEN** the response is `400 {"error": "lat is required"}` or `400 {"error": "lng is required"}`

#### Scenario: Invalid method
- **WHEN** a GET is made to `/api/v1/location`
- **THEN** the response is `405 Method Not Allowed`

### Requirement: user_locations table stores one live location per user
The system SHALL maintain a `user_locations` table with columns `user_id TEXT NOT NULL UNIQUE`, `lat DOUBLE PRECISION NOT NULL`, `lng DOUBLE PRECISION NOT NULL`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`. The `user_id` column stores the user's `auth_subject`.

#### Scenario: New user location row is created
- **WHEN** a user pushes their location for the first time
- **THEN** a new row is inserted into `user_locations`

#### Scenario: Existing row is updated, not duplicated
- **WHEN** the same user pushes their location again
- **THEN** exactly one row exists in `user_locations` for that `user_id` with updated coordinates and `updated_at`

### Requirement: Stale locations are excluded from queries
The system SHALL treat a location as stale if `updated_at < now() - (2 × LOCATION_POLL_INTERVAL_SECONDS)`. Stale locations SHALL be excluded from all nearby-user queries. The stale threshold is derived from `LOCATION_POLL_INTERVAL_SECONDS` config; it is not a separate config value.

#### Scenario: Fresh location is included in nearby queries
- **WHEN** a user's `updated_at` is within the stale threshold
- **THEN** that user's location is eligible to appear in nearby-user results

#### Scenario: Stale location is excluded from nearby queries
- **WHEN** a user's `updated_at` is older than 2 × poll interval
- **THEN** that user does not appear in nearby-user results
