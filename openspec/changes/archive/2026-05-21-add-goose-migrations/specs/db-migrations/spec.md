## ADDED Requirements

### Requirement: Migrations are tracked and applied in order
The system SHALL use goose to apply pending SQL migrations in ascending numeric order on startup, recording each applied migration in the `goose_db_version` table.

#### Scenario: First boot with no prior migrations
- **WHEN** the backend starts and `goose_db_version` does not exist
- **THEN** goose SHALL create the tracking table and apply all migrations in order

#### Scenario: Boot with migrations already applied
- **WHEN** the backend starts and all migrations are already recorded in `goose_db_version`
- **THEN** goose SHALL apply no migrations and startup SHALL proceed immediately

#### Scenario: Boot with new pending migrations
- **WHEN** the backend starts and one or more migrations are not yet recorded in `goose_db_version`
- **THEN** goose SHALL apply only the pending migrations in order before the server begins accepting requests

---

### Requirement: Migration files use goose Up/Down format
Each migration file SHALL be a `.sql` file in `backend/db/migrations/` with a numeric prefix, containing a `-- +goose Up` block and a `-- +goose Down` block.

#### Scenario: Valid migration file format
- **WHEN** a migration file named `001_create_users.sql` exists with `-- +goose Up` and `-- +goose Down` sections
- **THEN** goose SHALL parse and apply the Up block on `goose.Up` and the Down block on `goose.Down`

#### Scenario: Migration files are embedded in the binary
- **WHEN** the backend binary is built
- **THEN** all files matching `migrations/*.sql` SHALL be embedded via `//go:embed` and goose SHALL read from the embedded filesystem at runtime

---

### Requirement: Migration files contain DDL only
Migration files SHALL contain only DDL statements (CREATE, ALTER, DROP, CREATE INDEX). DML statements that operate on existing rows (UPDATE, INSERT ... SELECT, DELETE) SHALL NOT appear in migration files.

#### Scenario: DDL migration completes quickly
- **WHEN** a migration file contains only DDL statements
- **THEN** the migration SHALL complete in seconds regardless of table size and SHALL NOT hold long-running locks

---

### Requirement: Startup fails if migrations fail
If any pending migration fails to apply, the backend SHALL exit with a non-zero status rather than starting in a partially-migrated state.

#### Scenario: Migration error at startup
- **WHEN** a migration file contains invalid SQL or a constraint violation
- **THEN** goose SHALL return an error, `db.New()` SHALL propagate it, and `main()` SHALL call `log.Fatalf` to halt startup
