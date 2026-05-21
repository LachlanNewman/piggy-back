package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// mockRow implements pgx.Row for testing.
type mockRow struct {
	id  int
	err error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int) = r.id
	return nil
}

// mockQuerier implements the querier interface for testing.
type mockQuerier struct {
	row pgx.Row
}

func (m *mockQuerier) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return m.row
}

func newTestDB(row pgx.Row) *DB {
	return &DB{pool: &mockQuerier{row: row}}
}

func validParams() CreateUserParams {
	dob, _ := time.Parse("2006-01-02", "1995-06-15")
	return CreateUserParams{
		FirstName:   "Jane",
		LastName:    "Doe",
		Email:       "jane@example.com",
		DateOfBirth: dob,
		Weight:      68.5,
		Gender:      "female",
	}
}

func TestDB_CreateUser_Success(t *testing.T) {
	store := newTestDB(&mockRow{id: 7})
	id, err := store.CreateUser(context.Background(), validParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Errorf("expected id 7, got %d", id)
	}
}

func TestDB_CreateUser_DuplicateEmail(t *testing.T) {
	dbErr := errors.New(`ERROR: duplicate key value violates unique constraint "users_email_key"`)
	store := newTestDB(&mockRow{err: dbErr})

	_, err := store.CreateUser(context.Background(), validParams())
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestDB_CreateUser_UnexpectedError(t *testing.T) {
	dbErr := errors.New("connection reset by peer")
	store := newTestDB(&mockRow{err: dbErr})

	_, err := store.CreateUser(context.Background(), validParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("expected generic error, got ErrDuplicateEmail")
	}
}
