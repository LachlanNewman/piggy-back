package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// mockRow implements pgx.Row for testing via a pluggable scan function.
type mockRow struct {
	scanFn func(dest ...any) error
	err    error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if r.scanFn != nil {
		return r.scanFn(dest...)
	}
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

func createRow(id int, inserted bool) *mockRow {
	return &mockRow{
		scanFn: func(dest ...any) error {
			*dest[0].(*int) = id
			*dest[1].(*bool) = inserted
			return nil
		},
	}
}

func userRow(u User) *mockRow {
	return &mockRow{
		scanFn: func(dest ...any) error {
			*dest[0].(*int) = u.ID
			*dest[1].(*string) = u.AuthSubject
			*dest[2].(*string) = u.FirstName
			*dest[3].(*string) = u.LastName
			*dest[4].(*string) = u.Email
			*dest[5].(*bool) = u.ProfileComplete
			return nil
		},
	}
}

func validParams() CreateUserParams {
	dob, _ := time.Parse("2006-01-02", "1995-06-15")
	return CreateUserParams{
		AuthSubject: "auth0|abc123",
		FirstName:   "Jane",
		LastName:    "Doe",
		Email:       "jane@example.com",
		DateOfBirth: dob,
		Weight:      68.5,
		Gender:      "female",
	}
}

func TestDB_CreateUser_Success_Insert(t *testing.T) {
	store := newTestDB(createRow(7, true))
	id, created, err := store.CreateUser(context.Background(), validParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Errorf("expected id 7, got %d", id)
	}
	if !created {
		t.Errorf("expected created=true for new insert")
	}
}

func TestDB_CreateUser_Success_Upsert(t *testing.T) {
	store := newTestDB(createRow(7, false))
	id, created, err := store.CreateUser(context.Background(), validParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Errorf("expected id 7, got %d", id)
	}
	if created {
		t.Errorf("expected created=false for upsert update")
	}
}

func TestDB_CreateUser_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection reset by peer")})
	_, _, err := store.CreateUser(context.Background(), validParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDB_GetUserBySubject_Found(t *testing.T) {
	want := User{
		ID:              3,
		AuthSubject:     "auth0|abc123",
		FirstName:       "Jane",
		LastName:        "Doe",
		Email:           "jane@example.com",
		ProfileComplete: true,
	}
	store := newTestDB(userRow(want))
	got, err := store.GetUserBySubject(context.Background(), "auth0|abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestDB_GetUserBySubject_NotFound(t *testing.T) {
	store := newTestDB(&mockRow{err: pgx.ErrNoRows})
	_, err := store.GetUserBySubject(context.Background(), "auth0|unknown")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDB_GetUserBySubject_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection refused")})
	_, err := store.GetUserBySubject(context.Background(), "auth0|abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("expected generic error, got ErrNotFound")
	}
}
