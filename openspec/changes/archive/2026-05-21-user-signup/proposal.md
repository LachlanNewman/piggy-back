## Why

The app has no way to create or persist users — the backend currently only serves static health/hello endpoints. A signup flow is the foundational step needed before any user-specific features can be built.

## What Changes

- Add a `POST /api/users` endpoint in the Go backend that accepts user details and inserts a new user row into PostgreSQL.
- Add a `users` table to the database (via migration) with columns for first name, last name, email (unique), date of birth, weight, gender, created at, and updated at.
- Add a signup form in the React frontend that collects first name, last name, email, date of birth, weight, and gender, then POSTs to the backend.
- Display a success or error message after submission.

## Capabilities

### New Capabilities

- `user-signup`: End-to-end flow for creating a new user — signup form in the frontend, REST endpoint in the backend, and a `users` table in PostgreSQL.

### Modified Capabilities

## Impact

- **Backend**: New `POST /api/users` handler and user model; new `users` table migration.
- **Frontend**: New `SignupForm` component; `App.jsx` updated to render it.
- **Database**: New `users` table (id, first_name, last_name, email unique, date_of_birth, weight, gender, created_at, updated_at).
- **No breaking changes** to existing `/api/health` or `/api/hello` endpoints.
