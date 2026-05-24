package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func rideRequestRow(r RideRequest) *mockRow {
	return &mockRow{
		scanFn: func(dest ...any) error {
			*dest[0].(*string) = r.ID
			*dest[1].(*string) = r.DriverID
			*dest[2].(*string) = r.RiderID
			*dest[3].(*string) = r.Status
			*dest[4].(*string) = r.PickupAddress
			*dest[5].(*string) = r.DropoffAddress
			*dest[6].(*time.Time) = r.RequestedAt
			*dest[7].(*time.Time) = r.ExpiresAt
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

func boolRow(v bool) *mockRow {
	return &mockRow{
		scanFn: func(dest ...any) error {
			*dest[0].(*bool) = v
			return nil
		},
	}
}

func jsonRow(raw json.RawMessage) *mockRow {
	return &mockRow{
		scanFn: func(dest ...any) error {
			*dest[0].(*json.RawMessage) = raw
			return nil
		},
	}
}

func validCreateParams() CreateRideRequestParams {
	return CreateRideRequestParams{
		RiderID:        "auth0|rider",
		DriverID:       "auth0|driver",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Oak Ave",
		ExpiresAt:      time.Now().Add(15 * time.Minute),
	}
}

func TestDB_CreateRideRequest_Success(t *testing.T) {
	const wantID = "550e8400-e29b-41d4-a716-446655440000"
	store := newTestDB(insertedIDRow(wantID))
	id, err := store.CreateRideRequest(context.Background(), validCreateParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != wantID {
		t.Errorf("expected id %q, got %q", wantID, id)
	}
}

func TestDB_CreateRideRequest_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection refused")})
	_, err := store.CreateRideRequest(context.Background(), validCreateParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDB_GetRideRequestByID_Found(t *testing.T) {
	future := time.Now().Add(15 * time.Minute).UTC().Truncate(time.Second)
	want := RideRequest{
		ID:             "550e8400-e29b-41d4-a716-446655440000",
		DriverID:       "auth0|driver",
		RiderID:        "auth0|rider",
		Status:         "pending",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Oak Ave",
		RequestedAt:    time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
		ExpiresAt:      future,
	}
	store := newTestDB(rideRequestRow(want))
	got, err := store.GetRideRequestByID(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID || got.Status != want.Status || got.DriverID != want.DriverID {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestDB_GetRideRequestByID_ExpiredVirtualStatus(t *testing.T) {
	past := time.Now().Add(-1 * time.Minute)
	rr := RideRequest{
		ID:        "some-id",
		Status:    "pending",
		ExpiresAt: past,
	}
	store := newTestDB(rideRequestRow(rr))
	got, err := store.GetRideRequestByID(context.Background(), rr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "expired" {
		t.Errorf("expected status=expired for past ExpiresAt, got %q", got.Status)
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

func TestDB_HasActivePendingRequest_True(t *testing.T) {
	store := newTestDB(boolRow(true))
	has, err := store.HasActivePendingRequest(context.Background(), "auth0|rider")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected true, got false")
	}
}

func TestDB_HasActivePendingRequest_False(t *testing.T) {
	store := newTestDB(boolRow(false))
	has, err := store.HasActivePendingRequest(context.Background(), "auth0|rider")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected false, got true")
	}
}

func TestDB_HasActivePendingRequest_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection refused")})
	_, err := store.HasActivePendingRequest(context.Background(), "auth0|rider")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDB_GetIncomingRequests_HasResults(t *testing.T) {
	raw, _ := json.Marshal([]incomingRequestJSON{{
		ID:             "some-uuid",
		RiderID:        "auth0|rider",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Oak Ave",
		RequestedAt:    time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
		ExpiresAt:      time.Date(2026, 5, 24, 10, 15, 0, 0, time.UTC),
	}})
	store := newTestDB(jsonRow(raw))
	got, err := store.GetIncomingRequests(context.Background(), "auth0|driver")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "some-uuid" || got[0].RiderID != "auth0|rider" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestDB_GetIncomingRequests_Empty(t *testing.T) {
	store := newTestDB(jsonRow(json.RawMessage("[]")))
	got, err := store.GetIncomingRequests(context.Background(), "auth0|driver")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %+v", got)
	}
}

func TestDB_GetIncomingRequests_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection refused")})
	_, err := store.GetIncomingRequests(context.Background(), "auth0|driver")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDB_SetRideRequestStatus_Success(t *testing.T) {
	store := newTestDB(insertedIDRow("some-id"))
	err := store.SetRideRequestStatus(context.Background(), "some-id", "accepted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDB_SetRideRequestStatus_NotFound(t *testing.T) {
	store := newTestDB(&mockRow{err: pgx.ErrNoRows})
	err := store.SetRideRequestStatus(context.Background(), "unknown-id", "accepted")
	if !errors.Is(err, ErrRideRequestNotFound) {
		t.Errorf("expected ErrRideRequestNotFound, got %v", err)
	}
}

func TestDB_SetRideRequestStatus_DBError(t *testing.T) {
	store := newTestDB(&mockRow{err: errors.New("connection refused")})
	err := store.SetRideRequestStatus(context.Background(), "some-id", "accepted")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
