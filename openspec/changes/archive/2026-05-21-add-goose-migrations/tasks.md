## 1. Dependencies

- [x] 1.1 Add `github.com/pressly/goose/v3` to `backend/go.mod` via `go get`

## 2. Convert Existing Migration

- [x] 2.1 Rewrite `backend/db/migrations/001_create_users.sql` with `-- +goose Up` and `-- +goose Down` annotations — Up block contains the existing DDL, Down block drops the table and type

## 3. Update db.New()

- [x] 3.1 Replace `//go:embed migrations/001_create_users.sql` with `//go:embed migrations/*.sql` pointing at an `embed.FS`
- [x] 3.2 Open a `*sql.DB` using pgx's stdlib adapter (`github.com/jackc/pgx/v5/stdlib`) for the goose migration run
- [x] 3.3 Call `goose.SetBaseFS(embedFS)`, `goose.SetDialect("postgres")`, and `goose.Up(sqlDB, "migrations")` before returning the pool
- [x] 3.4 Close the `*sql.DB` after migrations complete (defer)
- [x] 3.5 Propagate any goose error via `log.Fatalf` in `main.go` (already handled — `db.New` returns error)

## 4. Verification

- [x] 4.1 Run `go build ./...` — confirm the backend compiles cleanly
- [x] 4.2 Start the stack (`docker compose up --build`) against a fresh database — confirm `goose_db_version` table is created and migration `001` is recorded
- [x] 4.3 Restart the backend without rebuilding — confirm no migrations re-run (goose skips already-applied)
- [x] 4.4 Run `go test ./db/... ./handlers/...` — confirm all existing tests still pass
