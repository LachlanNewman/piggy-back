package db

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_create_users.sql
var migration001 string

//go:embed migrations/002_create_ride_requests.sql
var migration002 string

func New(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	for _, m := range []string{migration001, migration002} {
		if _, err := pool.Exec(ctx, m); err != nil {
			return nil, fmt.Errorf("run migration: %w", err)
		}
	}

	return pool, nil
}
