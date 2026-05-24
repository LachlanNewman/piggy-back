package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockLocationRepo struct {
	fn func(ctx context.Context, sub string, lat, lng float64) error
}

func (m *mockLocationRepo) UpsertUserLocation(ctx context.Context, sub string, lat, lng float64) error {
	if m.fn != nil {
		return m.fn(ctx, sub, lat, lng)
	}
	return nil
}

func postLocation(sub, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/location?sub="+sub, bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

const validLocationBody = `{"lat":-33.8688,"lng":151.2093}`

func TestPushLocation_Success(t *testing.T) {
	repo := &mockLocationRepo{}
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, postLocation("auth0|abc", validLocationBody))

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestPushLocation_MissingSub(t *testing.T) {
	repo := &mockLocationRepo{}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/location", bytes.NewBufferString(validLocationBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "sub is required")
}

func TestPushLocation_InvalidJSON(t *testing.T) {
	repo := &mockLocationRepo{}
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, postLocation("auth0|abc", `not json`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "invalid JSON")
}

func TestPushLocation_DBError(t *testing.T) {
	repo := &mockLocationRepo{
		fn: func(_ context.Context, _ string, _, _ float64) error {
			return errors.New("connection refused")
		},
	}
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, postLocation("auth0|abc", validLocationBody))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	assertError(t, w, "could not update location")
}

func TestPushLocation_MissingLat(t *testing.T) {
	repo := &mockLocationRepo{}
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, postLocation("auth0|abc", `{"lng":151.2093}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "lat is required")
}

func TestPushLocation_MissingLng(t *testing.T) {
	repo := &mockLocationRepo{}
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, postLocation("auth0|abc", `{"lat":-33.8688}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "lng is required")
}

func TestPushLocation_ZeroCoordinatesAccepted(t *testing.T) {
	repo := &mockLocationRepo{}
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, postLocation("auth0|abc", `{"lat":0,"lng":0}`))

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for valid 0,0 coordinates, got %d", w.Code)
	}
}

func TestPushLocation_MethodNotAllowed(t *testing.T) {
	repo := &mockLocationRepo{}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/location?sub=auth0|abc", nil)
	w := httptest.NewRecorder()
	PushLocation(repo).ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
