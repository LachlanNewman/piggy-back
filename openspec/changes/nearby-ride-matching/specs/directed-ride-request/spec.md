## ADDED Requirements

### Requirement: Driver can poll for incoming ride requests
The system SHALL accept `GET /api/v1/ride-requests/incoming?sub=<auth_subject>` and return all pending, non-expired ride requests where `driver_id = sub`. The response SHALL be a JSON array; each item SHALL include `id`, `rider_id`, `pickup_address`, `dropoff_address`, `requested_at`, and `expires_at`.

#### Scenario: Pending requests exist for driver
- **WHEN** a driver polls and has one or more pending, non-expired requests directed at them
- **THEN** the response is `200` with an array of those requests

#### Scenario: No incoming requests
- **WHEN** no pending non-expired requests are directed at the driver
- **THEN** the response is `200 []`

#### Scenario: Missing sub query param
- **WHEN** the request is made without `?sub=`
- **THEN** the response is `400 {"error": "sub is required"}`

### Requirement: Driver can accept a ride request
The system SHALL accept `PATCH /api/v1/ride-requests/{id}/accept?sub=<driver_sub>`. The handler SHALL verify that the request exists, is in `pending` status, has not expired, and that `driver_id` matches `sub`. On success it SHALL update `status` to `accepted` and return `200 {"id": "..."}`.

#### Scenario: Valid accept
- **WHEN** a driver PATCHes accept on a pending non-expired request directed at them
- **THEN** the status becomes `accepted` and the response is `200 {"id": "..."}`

#### Scenario: Request already accepted
- **WHEN** the request status is already `accepted`
- **THEN** the response is `409 {"error": "request already accepted"}`

#### Scenario: Request expired
- **WHEN** `expires_at` is in the past
- **THEN** the response is `410 {"error": "request has expired"}`

#### Scenario: Wrong driver
- **WHEN** `sub` does not match `driver_id`
- **THEN** the response is `403 {"error": "forbidden"}`

### Requirement: Driver can decline a ride request
The system SHALL accept `PATCH /api/v1/ride-requests/{id}/decline?sub=<driver_sub>`. The handler SHALL apply the same pre-conditions as accept. On success it SHALL update `status` to `declined` and return `200 {"id": "..."}`.

#### Scenario: Valid decline
- **WHEN** a driver PATCHes decline on a pending non-expired request directed at them
- **THEN** the status becomes `declined` and the response is `200 {"id": "..."}`

#### Scenario: Request already declined
- **WHEN** the request status is already `declined`
- **THEN** the response is `409 {"error": "request already declined"}`

#### Scenario: Request expired before decline
- **WHEN** `expires_at` is in the past
- **THEN** the response is `410 {"error": "request has expired"}`

### Requirement: Frontend polls for incoming requests and shows accept/decline UI
The frontend SHALL poll `GET /api/v1/ride-requests/incoming` on the same interval as the location push (`LOCATION_POLL_INTERVAL_SECONDS`). When one or more pending requests are present, the frontend SHALL display each with the rider's name, pickup/dropoff addresses, and Accept / Decline buttons.

#### Scenario: Incoming request appears
- **WHEN** the poll returns one or more pending requests
- **THEN** the frontend displays a notification panel with Accept and Decline buttons for each

#### Scenario: Request accepted by driver
- **WHEN** driver clicks Accept
- **THEN** frontend PATCHes accept and hides the request from the panel

#### Scenario: Request declined by driver
- **WHEN** driver clicks Decline
- **THEN** frontend PATCHes decline and hides the request from the panel

### Requirement: Frontend shows request status to rider while waiting
After submitting a directed ride request, the frontend SHALL poll `GET /api/v1/ride-requests/{id}` on the same interval. It SHALL display the current status and stop polling when status is `accepted`, `declined`, or `expired`.

#### Scenario: Driver accepts
- **WHEN** status becomes `accepted`
- **THEN** frontend stops polling and shows "Your driver is on the way!"

#### Scenario: Driver declines
- **WHEN** status becomes `declined`
- **THEN** frontend stops polling and shows "Your request was declined. Try another nearby user."

#### Scenario: Request expires
- **WHEN** status becomes `expired` or `expires_at` is past with status still `pending`
- **THEN** frontend stops polling and shows "Your request timed out. Try again."
