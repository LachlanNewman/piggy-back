# Capability: user-signup

## Purpose

Enables new users to register an account by providing personal details. The system stores user records in PostgreSQL and exposes a REST endpoint for account creation. A frontend signup form collects the required fields and communicates with the backend API.

---
## Requirements
### Requirement: Users table exists in the database
The system SHALL have a `users` table in PostgreSQL with columns: `id` (serial primary key), `auth_subject` (text, not null, unique — the OIDC `sub` claim), `first_name` (text, not null), `last_name` (text, not null), `email` (text, not null, unique), `date_of_birth` (date, nullable), `weight` (numeric(5,2), nullable), `gender` (gender_enum, nullable — one of `male`, `female`, `unknown`), `profile_complete` (boolean, not null, default false), `created_at` (timestamptz, default now(), not null), and `updated_at` (timestamptz, default now(), not null).

#### Scenario: Table is created on first startup
- **WHEN** the backend starts and the `users` table does not exist
- **THEN** the table SHALL be created automatically before the server begins accepting requests

#### Scenario: Startup is idempotent
- **WHEN** the backend starts and the `users` table already exists
- **THEN** startup SHALL succeed without error and the table SHALL remain unchanged

### Requirement: Create user via POST /api/v1/users
The backend SHALL expose a `POST /api/v1/users` endpoint that accepts a JSON body with `auth_subject` (string), `first_name` (string), `last_name` (string), `email` (string), `date_of_birth` (string, ISO 8601 `YYYY-MM-DD`), `weight` (number), and `gender` (string, one of `male`, `female`, `unknown`), inserts or updates a row in the `users` table, sets `profile_complete = true`, and returns the created user's ID.

#### Scenario: Successful user creation
- **WHEN** a client sends `POST /api/v1/users` with a valid JSON body containing all required fields
- **THEN** the system SHALL insert a row into `users` with `profile_complete = true` and return HTTP 201 with JSON `{ "id": <int> }`

#### Scenario: Missing required field
- **WHEN** a client sends `POST /api/v1/users` with one or more required fields missing or empty
- **THEN** the system SHALL return HTTP 400 with JSON `{ "error": "<field> is required" }`

#### Scenario: Duplicate auth_subject (retry / race)
- **WHEN** a client sends `POST /api/v1/users` with an `auth_subject` that already exists in the `users` table
- **THEN** the system SHALL update the biometric fields (`date_of_birth`, `weight`, `gender`) and set `profile_complete = true`, returning HTTP 200 with JSON `{ "id": <int> }`

#### Scenario: Invalid gender value
- **WHEN** a client sends `POST /api/v1/users` with a `gender` value that is not one of `male`, `female`, or `unknown`
- **THEN** the system SHALL return HTTP 400 with JSON `{ "error": "gender must be one of: male, female, unknown" }`

#### Scenario: Invalid HTTP method
- **WHEN** a client sends a request to `/api/v1/users` with a method other than POST or OPTIONS
- **THEN** the system SHALL return HTTP 405

