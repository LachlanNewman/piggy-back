## 1. Dependencies

- [x] 1.1 Add `github.com/go-playground/validator/v10` to `backend/go.mod` via `go get`

## 2. Database Migration

- [x] 2.1 Create `backend/db/migrations/001_create_users.sql`: define `CREATE TYPE gender_enum AS ENUM ('male','female','unknown')` and `CREATE TABLE IF NOT EXISTS users` (id, first_name, last_name, email unique, date_of_birth, weight, gender gender_enum, created_at, updated_at)
- [x] 2.2 Update `backend/db/db.go` (or `main.go`) to read and execute the migration SQL on startup before the server starts listening

## 3. Backend — User Handler

- [x] 3.1 Define a `CreateUserRequest` struct in `backend/handlers/users.go` with `validate` struct tags: `required,email` for email; `required,gt=0` for weight; `required,oneof=male female unknown` for gender; `required` for remaining fields
- [x] 3.2 In the `CreateUser` handler, decode the JSON body then run `validator.Validate.Struct()` — on failure return HTTP 400 with the first validation error as `{ "error": "<field>: <reason>" }`
- [x] 3.3 Add the INSERT query using `pgxpool.Pool.QueryRow`, returning the new row's `id`, `created_at`, and `updated_at`
- [x] 3.4 Return HTTP 409 for duplicate email and HTTP 201 with the created user JSON on success
- [x] 3.5 Register `POST /api/v1/users` in `main.go`, passing the db pool to the handler

## 4. Frontend — Signup Form

- [x] 4.1 Create `frontend/src/SignupForm.jsx` with controlled inputs for `first_name`, `last_name`, `email`, `date_of_birth` (date input), `weight` (number input), and `gender` (select with options: Male, Female, Unknown)
- [x] 4.2 Implement form submit: POST to `/api/v1/users` with JSON body, disable submit button while request is in flight
- [x] 4.3 On 201 response: display success message and clear form fields
- [x] 4.4 On 400 or 409 response: display the `error` field from the response body without clearing fields
- [x] 4.5 On network error or 5xx: display generic error message

## 5. Frontend — Integration

- [x] 5.1 Import and render `SignupForm` in `frontend/src/App.jsx`

## 6. End-to-End Verification

- [x] 6.1 Start the stack (`docker compose up`) and submit the form with valid data — confirm the user row appears in the database with all columns populated
- [x] 6.2 Submit the same email a second time — confirm the 409 error message appears in the UI
- [x] 6.3 Submit with a missing field — confirm the 400 error message appears in the UI
- [x] 6.4 Submit with an invalid email format — confirm the 400 validation error appears in the UI
- [x] 6.5 Send a raw request with an invalid gender value — confirm the 400 error `gender must be one of: male, female, unknown` is returned
- [x] 6.6 Restart the backend — confirm startup is idempotent and the table is not dropped or altered
