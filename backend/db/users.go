package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEmail = errors.New("email already registered")

type CreateUserParams struct {
	FirstName   string
	LastName    string
	Email       string
	DateOfBirth time.Time
	Weight      float64
	Gender      string
}

// querier is the subset of pgxpool.Pool used by DB, allowing it to be mocked in tests.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type DB struct {
	pool querier
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

func (d *DB) CreateUser(ctx context.Context, p CreateUserParams) (int, error) {
	var id int
	err := d.pool.QueryRow(ctx, `
		INSERT INTO users (first_name, last_name, email, date_of_birth, weight, gender)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, p.FirstName, p.LastName, p.Email, p.DateOfBirth, p.Weight, p.Gender).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "users_email_key") {
			return 0, ErrDuplicateEmail
		}
		return 0, err
	}
	return id, nil
}
