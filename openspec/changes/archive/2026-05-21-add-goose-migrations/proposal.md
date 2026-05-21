## Why

The current approach embeds a single SQL file and runs it on every boot using `CREATE TABLE IF NOT EXISTS` — idempotent for table creation but unable to safely track or apply incremental schema changes. As the schema grows, there is no record of what has been applied, and ALTER TABLE statements cannot be made idempotent without a proper migration tracking system.

## What Changes

- Add `github.com/pressly/goose/v3` as a dependency.
- Convert `backend/db/migrations/001_create_users.sql` to goose format with `-- +goose Up` and `-- +goose Down` annotations.
- Replace the raw `pool.Exec(migrationSQL)` call in `db.New()` with a goose migration run using the embedded FS.
- Remove the single-file `//go:embed` of `migrationSQL`; replace with `//go:embed migrations/*.sql` pointed at the migrations directory.
- Goose will create and manage a `goose_db_version` table to track applied migrations.

## Capabilities

### New Capabilities

- `db-migrations`: The backend tracks and applies database schema migrations in order using goose, with a persistent record of what has been applied.

### Modified Capabilities

## Impact

- **`backend/db/db.go`**: Migration logic replaced; goose runs all pending migrations at startup.
- **`backend/db/migrations/001_create_users.sql`**: Reformatted with goose Up/Down annotations.
- **`backend/go.mod`**: New dependency `github.com/pressly/goose/v3`.
- **No API or handler changes** — purely internal infrastructure.
- **New table** `goose_db_version` created in the database on first run.
