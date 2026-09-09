package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/db"
	"backend/events"
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
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo, &recordingPublisher{}))
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
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo, &recordingPublisher{}))
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
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo, &recordingPublisher{}))
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
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo, &recordingPublisher{}))
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
	mux.HandleFunc("/api/v1/ride-requests/{id}/decline", DeclineRideRequest(repo, &recordingPublisher{}))
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
	mux.HandleFunc("/api/v1/ride-requests/{id}/decline", DeclineRideRequest(repo, &recordingPublisher{}))
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
	mux.HandleFunc("/api/v1/ride-requests/{id}/decline", DeclineRideRequest(repo, &recordingPublisher{}))
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/ride-requests/unknown/decline?sub=auth0|driver", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	assertError(t, w, "not found")
}

// Live event publishing

func TestAcceptRideRequest_NotifiesTheRider(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return pendingRequest("auth0|driver"), nil
		},
	}
	pub := &recordingPublisher{}
	w := patchAction("/api/v1/ride-requests/{id}/accept", "auth0|driver", AcceptRideRequest(repo, pub))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	got := pub.events()
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 event, got %d", len(got))
	}
	if got[0].Sub != "auth0|rider" {
		t.Errorf("notified %q, want the rider auth0|rider", got[0].Sub)
	}
	if got[0].Event.Type != events.TypeRideRequestUpdated {
		t.Errorf("event type = %q", got[0].Event.Type)
	}
	if got[0].Event.Status != "accepted" {
		t.Errorf("event status = %q, want accepted", got[0].Event.Status)
	}
}

func TestDeclineRideRequest_NotifiesTheRider(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return pendingRequest("auth0|driver"), nil
		},
	}
	pub := &recordingPublisher{}
	w := patchAction("/api/v1/ride-requests/{id}/decline", "auth0|driver", DeclineRideRequest(repo, pub))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	got := pub.events()
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 event, got %d", len(got))
	}
	if got[0].Event.Status != "declined" {
		t.Errorf("event status = %q, want declined", got[0].Event.Status)
	}
}

func TestRideRequestAction_NoEventWhenRejected(t *testing.T) {
	cases := []struct {
		name string
		req  db.RideRequest
		sub  string
	}{
		{"wrong driver", pendingRequest("auth0|driver"), "auth0|someone-else"},
		{"already accepted", acceptedRequest("auth0|driver"), "auth0|driver"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockActionRepo{
				getFn: func(_ context.Context, _ string) (db.RideRequest, error) { return tc.req, nil },
			}
			pub := &recordingPublisher{}
			patchAction("/api/v1/ride-requests/{id}/accept", tc.sub, AcceptRideRequest(repo, pub))

			if n := len(pub.events()); n != 0 {
				t.Fatalf("expected no events, got %d", n)
			}
		})
	}
}

func TestAcceptRideRequest_NoEventWhenUpdateFails(t *testing.T) {
	repo := &mockActionRepo{
		getFn: func(_ context.Context, _ string) (db.RideRequest, error) {
			return pendingRequest("auth0|driver"), nil
		},
		setStatusFn: func(_ context.Context, _, _ string) error { return errors.New("db error") },
	}
	pub := &recordingPublisher{}
	w := patchAction("/api/v1/ride-requests/{id}/accept", "auth0|driver", AcceptRideRequest(repo, pub))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if n := len(pub.events()); n != 0 {
		t.Fatalf("expected no events when the write failed, got %d", n)
	}
}

func acceptedRequest(driverID string) db.RideRequest {
	rr := pendingRequest(driverID)
	rr.Status = "accepted"
	return rr
}
