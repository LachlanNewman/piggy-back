package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")

type User struct {
	ID              int
	AuthSubject     string
	FirstName       string
	LastName        string
	Email           string
	ProfileComplete bool
}

type CreateUserParams struct {
	AuthSubject string
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

// CreateUser inserts a new user row. Returns (id, created, err) where created is true for an
// INSERT and false when an existing auth_subject row was updated (upsert path).
func (d *DB) CreateUser(ctx context.Context, p CreateUserParams) (int, bool, error) {
	var id int
	var created bool
	err := d.pool.QueryRow(ctx, `
		INSERT INTO users (auth_subject, first_name, last_name, email, date_of_birth, weight, gender, profile_complete)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		ON CONFLICT (auth_subject) DO UPDATE SET
			date_of_birth    = EXCLUDED.date_of_birth,
			weight           = EXCLUDED.weight,
			gender           = EXCLUDED.gender,
			profile_complete = true,
			updated_at       = NOW()
		RETURNING id, (xmax = 0) AS inserted
	`, p.AuthSubject, p.FirstName, p.LastName, p.Email, p.DateOfBirth, p.Weight, p.Gender).Scan(&id, &created)
	return id, created, err
}

// GetUserBySubject returns the user matching the given OIDC sub, or ErrNotFound.
func (d *DB) GetUserBySubject(ctx context.Context, sub string) (User, error) {
	var u User
	err := d.pool.QueryRow(ctx, `
		SELECT id, auth_subject, first_name, last_name, email, profile_complete
		FROM users
		WHERE auth_subject = $1
	`, sub).Scan(&u.ID, &u.AuthSubject, &u.FirstName, &u.LastName, &u.Email, &u.ProfileComplete)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return u, nil
}
