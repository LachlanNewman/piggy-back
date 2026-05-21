## Context

The project is a React (Vite) + Go + PostgreSQL stack. The backend uses Go's standard `net/http` with `pgx/v5` for Postgres. There is currently no user model, no database schema, and no frontend form — only health/hello endpoints. This design covers adding the first real data-write flow.

## Goals / Non-Goals

**Goals:**
- Define the `users` table schema and the migration to create it.
- Expose `POST /api/v1/users` to insert a new user row.
- Build a `SignupForm` React component that posts user data and shows feedback.

**Non-Goals:**
- Authentication or sessions (login, JWTs, cookies).
- Password storage or hashing.
- Email format validation beyond non-empty check.
- Input sanitization beyond basic null/empty checks.
- Pagination or listing of users.

## Decisions

### Database migration via raw SQL file

**Decision**: Use a plain SQL migration file (`backend/db/migrations/001_create_users.sql`) executed at startup via `pgx`.

**Rationale**: The project has no migration tool set up yet. Adding one (golang-migrate, goose) would be premature for a single table. A startup-time `EXEC` of a `CREATE TABLE IF NOT EXISTS` statement is simple, safe, and easy to replace later.

**Alternative considered**: golang-migrate — adds tooling overhead and a separate migration binary; deferred until the schema is more complex.

---

### Single POST endpoint, no versioning

**Decision**: `POST /api/v1/users` with a flat JSON body `{ first_name, last_name, email, date_of_birth, weight, gender }`.

**Rationale**: The app has no API versioning strategy yet. Introducing `/api/v1/` now would be speculative. A plain path is consistent with existing endpoints and easy to change later.

---

### Input validation via go-playground/validator

**Decision**: Use `github.com/go-playground/validator/v10` to validate the request body struct before inserting into the database.

**Rationale**: Provides declarative, struct-tag-based rules (`required`, `email`, `gt=0`, `oneof=male female unknown`) and returns structured field-level errors that can be mapped to a consistent `{ "error": "<field>: <reason>" }` response. Keeps validation out of the handler logic and makes rules self-documenting on the struct.

**Alternative considered**: Manual field checks — already partially in place for the empty-string guard, but doesn't handle type constraints (e.g. valid email format, weight > 0) without growing into a bespoke validation layer.

---

### No ORM — direct pgx queries

**Decision**: Write the INSERT directly using `pgxpool.Pool.Exec`.

**Rationale**: There is one query. Importing an ORM (sqlc, gorm) for a single insert adds unnecessary complexity at this stage.

---

### Frontend: inline form in App.jsx → extracted component

**Decision**: Add a `SignupForm.jsx` component, imported into `App.jsx`.

**Rationale**: Keeps `App.jsx` clean and makes the form independently testable without creating a routing layer that doesn't exist yet.

## Data Shape

### Request payload (`POST /api/v1/users`)

```json
{
  "first_name": "Jane",
  "last_name": "Doe",
  "email": "jane@example.com",
  "date_of_birth": "1995-06-15",
  "weight": 68.5,
  "gender": "female"
}
```

### Response body (HTTP 201)

```json
{
  "id": 1,
  "first_name": "Jane",
  "last_name": "Doe",
  "email": "jane@example.com",
  "date_of_birth": "1995-06-15",
  "weight": 68.5,
  "gender": "female",
  "created_at": "2026-05-21T10:34:00Z",
  "updated_at": "2026-05-21T10:34:00Z"
}
```

### Database row (`users` table)

| column          | type              | example value              |
|-----------------|-------------------|----------------------------|
| `id`            | serial (PK)       | `1`                        |
| `first_name`    | text not null     | `Jane`                     |
| `last_name`     | text not null     | `Doe`                      |
| `email`         | text not null unique | `jane@example.com`      |
| `date_of_birth` | date not null     | `1995-06-15`               |
| `weight`        | numeric(5,2) not null | `68.50`               |
| `gender`        | gender_enum not null (`male`,`female`,`unknown`) | `female` |
| `created_at`    | timestamptz not null | `2026-05-21 10:34:00+00` |
| `updated_at`    | timestamptz not null | `2026-05-21 10:34:00+00` |

## Risks / Trade-offs

- **Email as the uniqueness key** → A unique constraint on `email` prevents duplicate accounts. If a duplicate is submitted, the backend returns 409; the frontend surfaces the error without clearing the form.
- **Startup migration on every boot** → `CREATE TABLE IF NOT EXISTS` is idempotent, so this is safe. The risk is migration drift if the table is later altered manually — mitigated by moving to a proper migration tool before adding more tables.

## Migration Plan

1. Add `backend/db/migrations/001_create_users.sql`.
2. Update `db.New` (or `main.go`) to run the migration on startup.
3. Add `POST /api/v1/users` handler and wire it in `main.go`.
4. Add `SignupForm.jsx` to the frontend.
5. Update `App.jsx` to render `SignupForm`.
6. Test end-to-end via the running Docker Compose stack.

Rollback: drop the `users` table and remove the handler. No data migration needed at this stage.
