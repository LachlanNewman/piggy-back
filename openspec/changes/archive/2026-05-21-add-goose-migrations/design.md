## Context

The backend currently embeds a single SQL file and executes it on every boot via `pool.Exec(migrationSQL)`. The SQL relies on `CREATE TABLE IF NOT EXISTS` and a `DO $$ BEGIN...EXCEPTION...END $$` block to be idempotent. This works for one migration but has no tracking — there is no record of what has been applied, and future ALTER TABLE statements cannot be written idempotently without significant complexity.

Goose is a Go-native migration library that tracks applied migrations in a `goose_db_version` table and runs only pending migrations, in order, on each boot.

## Goals / Non-Goals

**Goals:**
- Replace the raw `Exec` approach with goose-managed, tracked migrations.
- Keep migrations embedded in the binary via `//go:embed`.
- Convert the existing migration to goose format with Up and Down blocks.
- Enforce the discipline that migration files contain DDL only — no large DML backfills.
- Run migrations at startup (existing behaviour preserved).

**Non-Goals:**
- A separate migration CLI step or init container (startup-run is sufficient for now).
- Go-based (`.go`) migrations — SQL only at this stage.
- Any changes to the `users` table schema.
- Supporting rollback automation — Down migrations are defined but not run automatically.

## Decisions

### goose over golang-migrate

**Decision**: Use `github.com/pressly/goose/v3`.

**Rationale**: goose uses a single file per migration with `-- +goose Up` / `-- +goose Down` annotations, which is simpler to manage than golang-migrate's dual-file format (`.up.sql` + `.down.sql`). goose also has native pgx v5 support and a clean embedded FS API (`goose.SetBaseFS`). golang-migrate has more stars but adds more wiring for pgx v5 and offers no advantage for a single-database Go service.

---

### Run migrations at startup via library (not CLI)

**Decision**: Call `goose.Up` inside `db.New()` using the embedded SQL filesystem.

**Rationale**: Keeps the current operational model — start the container, migrations run, server starts. A separate CLI or init container adds deploy complexity that isn't warranted yet. If the team moves to Kubernetes or needs zero-downtime deploys, the migration step can be extracted then.

---

### DDL-only migration files

**Decision**: Migration files SHALL contain only DDL (CREATE, ALTER, DROP, CREATE INDEX CONCURRENTLY where appropriate). Data backfills are handled in application code, not migration files.

**Rationale**: Large DML inside a migration file blocks startup and holds table locks for the duration. Goose wraps each migration in a transaction by default; a long-running UPDATE inside a transaction holds an exclusive lock. This discipline keeps migrations fast and startup predictable.

---

### `database/sql` adapter for goose + pgx

**Decision**: Use `goose`'s `database/sql`-compatible driver via `goose.SetDialect("postgres")` and open a standard `*sql.DB` alongside the pgxpool for migration purposes only.

**Rationale**: Goose's library API operates on `*sql.DB`. Rather than pulling in the pgx goose driver separately, the cleanest approach is to open a standard `database/sql` connection (using `lib/pq` or `pgx`'s stdlib adapter) solely for the migration run, then close it. The `pgxpool` remains the runtime connection pool. This avoids coupling goose to pgx internals.

## Risks / Trade-offs

- **Two DB connections at startup** → One `*sql.DB` for goose (opened and closed during migration), one `pgxpool` for runtime. Negligible overhead; both connect to the same DSN.
- **Startup still blocks on migrations** → Acceptable now; fast because migrations are DDL-only. Revisit if a migration ever needs to backfill millions of rows.
- **`goose_db_version` table** → Goose creates this automatically. It is a permanent addition to the database schema and should not be dropped.
- **Existing DB has no goose tracking** → On first run against an already-provisioned database, goose will see `001` as pending and attempt to re-run the `CREATE TABLE IF NOT EXISTS` — which is idempotent and safe. The `goose_db_version` row will be inserted and subsequent boots skip it.
