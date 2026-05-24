package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func rideRequestRow(r RideRequest) *mockRow {
	return &mockRow{
		scanFn: func(dest ...any) error {
			*dest[0].(*string) = r.ID
			*dest[1].(*string) = r.Status
			*dest[2].(*string) = r.PickupAddress
			*dest[3].(*string) = r.DropoffAddress
			*dest[4].(*time.Time) = r.RequestedAt
			return nil
		},
	}
}

func insertedIDRow(id string) *mockRow {
	return &mockRow{
		scanFn: func(dest ...any) error {
			*dest[0].(*string) = id
			return nil
		},
	}
}

func TestDB_CreateRideRequest_Success(t *testing.T) {
	const wantID = "550e8400-e29b-41d4-a716-446655440000"
	store := newTestDB(insertedIDRow(wantID))
	id, err := store.CreateRideRequest(context.Background(), CreateRideRequestParams{
		RiderID:        "auth0|abc",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Oak Ave",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != wantID {
		t.Errorf("expected id %q, got %q", wantID, id)
	}
}

func TestDB_CreateRideRequest_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection refused")})
	_, err := store.CreateRideRequest(context.Background(), CreateRideRequestParams{
		RiderID:        "auth0|abc",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Oak Ave",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDB_GetRideRequestByID_Found(t *testing.T) {
	want := RideRequest{
		ID:             "550e8400-e29b-41d4-a716-446655440000",
		Status:         "pending",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Oak Ave",
		RequestedAt:    time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
	}
	store := newTestDB(rideRequestRow(want))
	got, err := store.GetRideRequestByID(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestDB_GetRideRequestByID_NotFound(t *testing.T) {
	store := newTestDB(&mockRow{err: pgx.ErrNoRows})
	_, err := store.GetRideRequestByID(context.Background(), "unknown-id")
	if !errors.Is(err, ErrRideRequestNotFound) {
		t.Errorf("expected ErrRideRequestNotFound, got %v", err)
	}
}

func TestDB_GetRideRequestByID_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection refused")})
	_, err := store.GetRideRequestByID(context.Background(), "some-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, ErrRideRequestNotFound) {
		t.Errorf("expected generic error, got ErrRideRequestNotFound")
	}
}
