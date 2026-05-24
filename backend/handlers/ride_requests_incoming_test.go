package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/db"
)

type mockIncomingRepo struct {
	fn func(ctx context.Context, driverID string) ([]db.RideRequest, error)
}

func (m *mockIncomingRepo) GetIncomingRequests(ctx context.Context, driverID string) ([]db.RideRequest, error) {
	if m.fn != nil {
		return m.fn(ctx, driverID)
	}
	return nil, nil
}

type mockActionRepo struct {
	getFn       func(ctx context.Context, id string) (db.RideRequest, error)
	setStatusFn func(ctx context.Context, id, status string) error
}

func (m *mockActionRepo) GetRideRequestByID(ctx context.Context, id string) (db.RideRequest, error) {
	return m.getFn(ctx, id)
}

func (m *mockActionRepo) SetRideRequestStatus(ctx context.Context, id, status string) error {
	if m.setStatusFn != nil {
		return m.setStatusFn(ctx, id, status)
	}
	return nil
}

func getIncoming(sub string, repo incomingRideRequestRepository) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/ride-requests/incoming?sub="+sub, nil)
	w := httptest.NewRecorder()
	GetIncomingRequests(repo).ServeHTTP(w, r)
	return w
}

func patchAction(path, sub string, handler http.Handler) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	r := httptest.NewRequest(http.MethodPatch, path+"?sub="+sub, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

func pendingRequest(driverID string) db.RideRequest {
	return db.RideRequest{
		ID:             "req-uuid",
		DriverID:       driverID,
		RiderID:        "auth0|rider",
		RiderFirstName: "Jane",
		RiderLastName:  "Doe",
		Status:         "pending",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Oak Ave",
		RequestedAt:    time.Now(),
		ExpiresAt:      time.Now().Add(15 * time.Minute),
	}
}

// GetIncomingRequests tests

func TestGetIncomingRequests_HasResults(t *testing.T) {
	repo := &mockIncomingRepo{
		fn: func(_ context.Context, _ string) ([]db.RideRequest, error) {
			return []db.RideRequest{pendingRequest("auth0|driver")}, nil
		},
	}
	w := getIncoming("auth0|driver", repo)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetIncomingRequests_Empty(t *testing.T) {
	repo := &mockIncomingRepo{
		fn: func(_ context.Context, _ string) ([]db.RideRequest, error) {
			return []db.RideRequest{}, nil
		},
	}
	w := getIncoming("auth0|driver", repo)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetIncomingRequests_MissingSub(t *testing.T) {
	repo := &mockIncomingRepo{}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/ride-requests/incoming", nil)
	w := httptest.NewRecorder()
	GetIncomingRequests(repo).ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "sub is required")
}

func TestGetIncomingRequests_DBError(t *testing.T) {
	repo := &mockIncomingRepo{
		fn: func(_ context.Context, _ string) ([]db.RideRequest, error) {
			return nil, errors.New("db error")
		},
	}
	w := getIncoming("auth0|driver", repo)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	assertError(t, w, "could not fetch incoming requests")
}

// AcceptRideRequest tests

func TestAcceptRideRequest_Success(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return pendingRequest("auth0|driver"), nil
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/req-uuid/accept?sub=auth0|driver", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAcceptRideRequest_WrongDriver(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return pendingRequest("auth0|driver"), nil
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/req-uuid/accept?sub=auth0|other", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	assertError(t, w, "forbidden")
}

func TestAcceptRideRequest_Expired(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			rr := pendingRequest("auth0|driver")
			rr.Status = "expired"
			return rr, nil
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/req-uuid/accept?sub=auth0|driver", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", w.Code)
	}
	assertError(t, w, "request has expired")
}

func TestAcceptRideRequest_AlreadyAccepted(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			rr := pendingRequest("auth0|driver")
			rr.Status = "accepted"
			return rr, nil
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/req-uuid/accept?sub=auth0|driver", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	assertError(t, w, "request already accepted")
}

func TestDeclineRideRequest_Success(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return pendingRequest("auth0|driver"), nil
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}/decline", DeclineRideRequest(repo))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/req-uuid/decline?sub=auth0|driver", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeclineRideRequest_AlreadyDeclined(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			rr := pendingRequest("auth0|driver")
			rr.Status = "declined"
			return rr, nil
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}/decline", DeclineRideRequest(repo))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/req-uuid/decline?sub=auth0|driver", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	assertError(t, w, "request already declined")
}

func TestDeclineRideRequest_NotFound(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return db.RideRequest{}, db.ErrRideRequestNotFound
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ride-requests/{id}/decline", DeclineRideRequest(repo))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/unknown/decline?sub=auth0|driver", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	assertError(t, w, "not found")
}
