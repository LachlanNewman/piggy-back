## ADDED Requirements

### Requirement: User can retrieve a list of nearby users
The system SHALL accept `GET /api/v1/users/nearby?sub=<auth_subject>` and return a JSON array of users who have a fresh location within the configured radius of the requesting user's last known location. Each item SHALL contain `id` (int), `first_name` (string), and `last_name` (string). The requesting user SHALL NOT appear in their own results. No coordinates SHALL be included in the response.

#### Scenario: Users within radius with fresh locations
- **WHEN** a user with a fresh location calls the endpoint and other users with fresh locations exist within `NEARBY_RADIUS_KM`
- **THEN** the response is `200` with a JSON array of those users' id, first_name, last_name

#### Scenario: No nearby users
- **WHEN** no other users have fresh locations within the radius
- **THEN** the response is `200` with an empty array `[]`

#### Scenario: Requesting user has no location on record
- **WHEN** the `sub` has no row in `user_locations`
- **THEN** the response is `404 {"error": "location not found — push your location first"}`

#### Scenario: Requesting user not included in results
- **WHEN** the calling user has a fresh location
- **THEN** their own profile does not appear in the returned list

#### Scenario: Missing sub query param
- **WHEN** the request is made without `?sub=`
- **THEN** the response is `400 {"error": "sub is required"}`

### Requirement: Radius is configurable but server-capped
The system SHALL use `NEARBY_RADIUS_KM` (default: 5) as the search radius. The server SHALL enforce a maximum of 20 km regardless of the configured value. The radius is not client-controlled.

#### Scenario: Configured radius is respected
- **WHEN** `NEARBY_RADIUS_KM=10` and a user is 8 km away with a fresh location
- **THEN** that user appears in the nearby list

#### Scenario: Server cap enforced
- **WHEN** `NEARBY_RADIUS_KM=50`
- **THEN** the system treats it as 20 km and only returns users within 20 km

### Requirement: Frontend displays nearby users list
The frontend SHALL fetch `GET /api/v1/users/nearby` on demand (user-triggered) and display a list of nearby users by name. Each entry SHALL have a "Request ride" button that initiates a directed ride request to that user.

#### Scenario: Nearby users returned
- **WHEN** the API returns a non-empty list
- **THEN** the frontend shows each user's full name with a "Request ride" button

#### Scenario: No nearby users
- **WHEN** the API returns an empty array
- **THEN** the frontend shows a "No nearby users found" message

#### Scenario: User has no location yet
- **WHEN** the API returns 404 (no location on record)
- **THEN** the frontend prompts the user to enable location sharing
