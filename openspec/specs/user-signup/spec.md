# Capability: user-signup

## Purpose

Enables new users to register an account by providing personal details. The system stores user records in PostgreSQL and exposes a REST endpoint for account creation. A frontend signup form collects the required fields and communicates with the backend API.

---

## Requirements

### Requirement: Users table exists in the database
The system SHALL have a `users` table in PostgreSQL with columns: `id` (serial primary key), `first_name` (text, not null), `last_name` (text, not null), `email` (text, not null, unique), `date_of_birth` (date, not null), `weight` (numeric(5,2), not null), `gender` (gender_enum, not null — one of `male`, `female`, `unknown`), `created_at` (timestamptz, default now(), not null), and `updated_at` (timestamptz, default now(), not null).

#### Scenario: Table is created on first startup
- **WHEN** the backend starts and the `users` table does not exist
- **THEN** the table SHALL be created automatically before the server begins accepting requests

#### Scenario: Startup is idempotent
- **WHEN** the backend starts and the `users` table already exists
- **THEN** startup SHALL succeed without error and the table SHALL remain unchanged

---

### Requirement: Create user via POST /api/v1/users
The backend SHALL expose a `POST /api/v1/users` endpoint that accepts a JSON body with `first_name` (string), `last_name` (string), `email` (string), `date_of_birth` (string, ISO 8601 date `YYYY-MM-DD`), `weight` (number), and `gender` (string, one of `male`, `female`, `unknown`), inserts a new row into the `users` table, and returns the created user.

#### Scenario: Successful user creation
- **WHEN** a client sends `POST /api/v1/users` with a valid JSON body containing all required fields
- **THEN** the system SHALL insert a row into `users` and return HTTP 201 with JSON `{ "id": <int>, "first_name": <string>, "last_name": <string>, "email": <string>, "date_of_birth": <string>, "weight": <number>, "gender": <string>, "created_at": <string>, "updated_at": <string> }`

#### Scenario: Missing required field
- **WHEN** a client sends `POST /api/v1/users` with one or more of `first_name`, `last_name`, `email`, `date_of_birth`, `weight`, or `gender` missing or empty
- **THEN** the system SHALL return HTTP 400 with JSON `{ "error": "<field> is required" }`

#### Scenario: Duplicate email
- **WHEN** a client sends `POST /api/v1/users` with an `email` that already exists in the `users` table
- **THEN** the system SHALL return HTTP 409 with JSON `{ "error": "email already registered" }`

#### Scenario: Invalid gender value
- **WHEN** a client sends `POST /api/v1/users` with a `gender` value that is not one of `male`, `female`, or `unknown`
- **THEN** the system SHALL return HTTP 400 with JSON `{ "error": "gender must be one of: male, female, unknown" }`

#### Scenario: Invalid HTTP method
- **WHEN** a client sends a request to `/api/v1/users` with a method other than POST or OPTIONS
- **THEN** the system SHALL return HTTP 405

---

### Requirement: Signup form in the frontend
The frontend SHALL render a signup form that collects `first_name`, `last_name`, `email`, `date_of_birth`, `weight`, and `gender`, submits them to `POST /api/v1/users`, and displays a result message.

#### Scenario: Successful signup
- **WHEN** the user fills in all fields with valid data and submits the form
- **THEN** the form SHALL POST to `/api/v1/users`, and on a 201 response display a success message (e.g. "Account created!") and clear the form fields

#### Scenario: Duplicate email error
- **WHEN** the backend returns a 409 response
- **THEN** the form SHALL display the `error` field from the response body without clearing the form fields

#### Scenario: Validation error from backend
- **WHEN** the backend returns a 400 response
- **THEN** the form SHALL display the `error` field from the response body without clearing the form fields

#### Scenario: Network or server error
- **WHEN** the request fails due to a network error or the backend returns a 5xx response
- **THEN** the form SHALL display a generic error message (e.g. "Something went wrong. Please try again.")

#### Scenario: Submit while request is in flight
- **WHEN** the user clicks submit and the request has not yet completed
- **THEN** the submit button SHALL be disabled until the request resolves
