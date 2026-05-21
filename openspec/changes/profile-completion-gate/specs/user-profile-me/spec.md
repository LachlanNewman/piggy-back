## ADDED Requirements

### Requirement: Get user profile by sub via GET /api/v1/users/me
The backend SHALL expose a `GET /api/v1/users/me` endpoint that accepts a `sub` query parameter and returns the matching user's profile and completion status.

#### Scenario: User exists with complete profile
- **WHEN** a client sends `GET /api/v1/users/me?sub=<sub>` and a row with that `auth_subject` exists and `profile_complete` is `true`
- **THEN** the system SHALL return HTTP 200 with JSON `{ "id": <int>, "auth_subject": <string>, "first_name": <string>, "last_name": <string>, "email": <string>, "profile_complete": true }`

#### Scenario: User exists with incomplete profile
- **WHEN** a client sends `GET /api/v1/users/me?sub=<sub>` and a row with that `auth_subject` exists and `profile_complete` is `false`
- **THEN** the system SHALL return HTTP 200 with JSON `{ "id": <int>, "auth_subject": <string>, "first_name": <string>, "last_name": <string>, "email": <string>, "profile_complete": false }`

#### Scenario: User does not exist
- **WHEN** a client sends `GET /api/v1/users/me?sub=<sub>` and no row with that `auth_subject` exists
- **THEN** the system SHALL return HTTP 404 with JSON `{ "error": "user not found" }`

#### Scenario: Missing sub query parameter
- **WHEN** a client sends `GET /api/v1/users/me` without a `sub` query parameter
- **THEN** the system SHALL return HTTP 400 with JSON `{ "error": "sub is required" }`

#### Scenario: Invalid HTTP method
- **WHEN** a client sends a request to `/api/v1/users/me` with a method other than GET or OPTIONS
- **THEN** the system SHALL return HTTP 405
