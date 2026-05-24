package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")
var ErrLocationNotFound = errors.New("location not found")

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

type NearbyUser struct {
	ID          int
	AuthSubject string
	FirstName   string
	LastName    string
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

// UpsertUserLocation stores the user's current GPS coordinates, overwriting any prior location.
func (d *DB) UpsertUserLocation(ctx context.Context, sub string, lat, lng float64) error {
	var discard string
	return d.pool.QueryRow(ctx, `
		INSERT INTO user_locations (user_id, lat, lng, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id) DO UPDATE
			SET lat = EXCLUDED.lat, lng = EXCLUDED.lng, updated_at = now()
		RETURNING user_id
	`, sub, lat, lng).Scan(&discard)
}

// GetUserLocation returns the last known lat/lng for the given sub, or ErrLocationNotFound.
func (d *DB) GetUserLocation(ctx context.Context, sub string) (lat, lng float64, err error) {
	err = d.pool.QueryRow(ctx, `
		SELECT lat, lng FROM user_locations WHERE user_id = $1
	`, sub).Scan(&lat, &lng)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, ErrLocationNotFound
		}
		return 0, 0, err
	}
	return lat, lng, nil
}

type nearbyUserJSON struct {
	ID          int    `json:"id"`
	AuthSubject string `json:"auth_subject"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

// GetNearbyUsers returns users with a fresh location within radiusKm of (lat, lng), excluding sub.
// A location is considered stale if updated_at is older than staleThreshold.
func (d *DB) GetNearbyUsers(ctx context.Context, sub string, lat, lng, radiusKm float64, staleThreshold time.Duration) ([]NearbyUser, error) {
	var raw json.RawMessage
	err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(
			json_agg(json_build_object(
				'id',           u.id,
				'auth_subject', u.auth_subject,
				'first_name',   u.first_name,
				'last_name',    u.last_name
			)),
			'[]'::json
		)
		FROM users u
		JOIN user_locations l ON l.user_id = u.auth_subject
		WHERE u.auth_subject != $1
		  AND u.profile_complete = true
		  AND l.updated_at > now() - make_interval(secs => $5)
		  AND (
		      6371.0 * acos(
		          LEAST(1.0,
		              cos(radians($2)) * cos(radians(l.lat))
		              * cos(radians(l.lng) - radians($3))
		              + sin(radians($2)) * sin(radians(l.lat))
		          )
		      )
		  ) < $4
	`, sub, lat, lng, radiusKm, staleThreshold.Seconds()).Scan(&raw)
	if err != nil {
		return nil, err
	}

	var rows []nearbyUserJSON
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	users := make([]NearbyUser, len(rows))
	for i, r := range rows {
		users[i] = NearbyUser{ID: r.ID, AuthSubject: r.AuthSubject, FirstName: r.FirstName, LastName: r.LastName}
	}
	return users, nil
}
