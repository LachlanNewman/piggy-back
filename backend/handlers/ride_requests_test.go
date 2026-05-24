package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/db"
)

type mockRideRequestRepo struct {
	createFn           func(ctx context.Context, p db.CreateRideRequestParams) (string, error)
	getFn              func(ctx context.Context, id string) (db.RideRequest, error)
	hasActiveFn        func(ctx context.Context, riderID string) (bool, error)
	setStatusFn        func(ctx context.Context, id, status string) error
	getIncomingFn      func(ctx context.Context, driverID string) ([]db.RideRequest, error)
}

func (m *mockRideRequestRepo) CreateRideRequest(ctx context.Context, p db.CreateRideRequestParams) (string, error) {
	return m.createFn(ctx, p)
}

func (m *mockRideRequestRepo) GetRideRequestByID(ctx context.Context, id string) (db.RideRequest, error) {
	return m.getFn(ctx, id)
}

func (m *mockRideRequestRepo) HasActivePendingRequest(ctx context.Context, riderID string) (bool, error) {
	if m.hasActiveFn != nil {
		return m.hasActiveFn(ctx, riderID)
	}
	return false, nil
}

func (m *mockRideRequestRepo) SetRideRequestStatus(ctx context.Context, id, status string) error {
	if m.setStatusFn != nil {
		return m.setStatusFn(ctx, id, status)
	}
	return nil
}

func (m *mockRideRequestRepo) GetIncomingRequests(ctx context.Context, driverID string) ([]db.RideRequest, error) {
	if m.getIncomingFn != nil {
		return m.getIncomingFn(ctx, driverID)
	}
	return nil, nil
}

var defaultTTL = 15 * time.Minute

func postRideRequest(sub, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/ride-requests?sub="+sub, bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func getRideRequestWithRepo(id string, repo rideRequestRepository) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}", GetRideRequest(repo))
	r := httptest.NewRequest(http.MethodGet, "/api/v1/ride-requests/"+id, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

const validRideBody = `{"pickup_address":"123 Main St","dropoff_address":"456 Oak Ave","driver_id":"auth0|driver"}`

// CreateRideRequest tests

func TestCreateRideRequest_Success(t *testing.T) {
	repo := &mockRideRequestRepo{
		createFn: func(_ context.Context, _ db.CreateRideRequestParams) (string, error) {
			return "550e8400-e29b-41d4-a716-446655440000", nil
		},
	}
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, postRideRequest("auth0|abc", validRideBody))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	assertJSONField(t, w, "id", "550e8400-e29b-41d4-a716-446655440000")
}

func TestCreateRideRequest_ActiveRequestConflict(t *testing.T) {
	repo := &mockRideRequestRepo{
		hasActiveFn: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, postRideRequest("auth0|abc", validRideBody))

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	assertError(t, w, "you already have an active request")
}

func TestCreateRideRequest_MissingSub(t *testing.T) {
	repo := &mockRideRequestRepo{}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/ride-requests", bytes.NewBufferString(validRideBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "sub is required")
}

func TestCreateRideRequest_MissingPickupAddress(t *testing.T) {
	repo := &mockRideRequestRepo{}
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, postRideRequest("auth0|abc", `{"dropoff_address":"456 Oak Ave","driver_id":"auth0|driver"}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "pickup_address is required")
}

func TestCreateRideRequest_MissingDropoffAddress(t *testing.T) {
	repo := &mockRideRequestRepo{}
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, postRideRequest("auth0|abc", `{"pickup_address":"123 Main St","driver_id":"auth0|driver"}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "dropoff_address is required")
}

func TestCreateRideRequest_MissingDriverID(t *testing.T) {
	repo := &mockRideRequestRepo{}
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, postRideRequest("auth0|abc", `{"pickup_address":"123 Main St","dropoff_address":"456 Oak Ave"}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "driver_id is required")
}

func TestCreateRideRequest_InvalidJSON(t *testing.T) {
	repo := &mockRideRequestRepo{}
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, postRideRequest("auth0|abc", `not json`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "invalid JSON")
}

func TestCreateRideRequest_DBError(t *testing.T) {
	repo := &mockRideRequestRepo{
		createFn: func(_ context.Context, _ db.CreateRideRequestParams) (string, error) {
			return "", errors.New("connection refused")
		},
	}
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, postRideRequest("auth0|abc", validRideBody))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	assertError(t, w, "could not create ride request")
}

func TestCreateRideRequest_MethodNotAllowed(t *testing.T) {
	repo := &mockRideRequestRepo{}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/ride-requests?sub=auth0|abc", nil)
	w := httptest.NewRecorder()
	CreateRideRequest(repo, defaultTTL).ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// GetRideRequest tests

func TestGetRideRequest_Found(t *testing.T) {
	const id = "550e8400-e29b-41d4-a716-446655440000"
	repo := &mockRideRequestRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return db.RideRequest{
				ID:             id,
				Status:         "pending",
				PickupAddress:  "123 Main St",
				DropoffAddress: "456 Oak Ave",
				RequestedAt:    time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
				ExpiresAt:      time.Date(2026, 5, 24, 10, 15, 0, 0, time.UTC),
			}, nil
		},
	}
	w := getRideRequestWithRepo(id, repo)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "pending" {
		t.Errorf("expected status=pending, got %q", body["status"])
	}
	if body["pickup_address"] != "123 Main St" {
		t.Errorf("expected pickup_address=%q, got %q", "123 Main St", body["pickup_address"])
	}
}

func TestGetRideRequest_NotFound(t *testing.T) {
	repo := &mockRideRequestRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return db.RideRequest{}, db.ErrRideRequestNotFound
		},
	}
	w := getRideRequestWithRepo("unknown-id", repo)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	assertError(t, w, "not found")
}

func TestGetRideRequest_DBError(t *testing.T) {
	repo := &mockRideRequestRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return db.RideRequest{}, errors.New("connection refused")
		},
	}
	w := getRideRequestWithRepo("some-id", repo)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	assertError(t, w, "could not get ride request")
}

func TestGetRideRequest_MethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}", GetRideRequest(&mockRideRequestRepo{}))
	r := httptest.NewRequest(http.MethodPost, "/api/v1/ride-requests/some-id", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func assertJSONField(t *testing.T, w *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body[key] != want {
		t.Errorf("expected %s=%q, got %q", key, want, body[key])
	}
}
