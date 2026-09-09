package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/db"
)

type mockNearbyRepo struct {
	locationFn func(ctx context.Context, sub string) (float64, float64, error)
	nearbyFn   func(ctx context.Context, sub string, lat, lng, radiusKm float64, stale time.Duration) ([]db.NearbyUser, error)
}

func (m *mockNearbyRepo) GetUserLocation(ctx context.Context, sub string) (float64, float64, error) {
	if m.locationFn != nil {
		return m.locationFn(ctx, sub)
	}
	return 0, 0, nil
}

func (m *mockNearbyRepo) GetNearbyUsers(ctx context.Context, sub string, lat, lng, radiusKm float64, stale time.Duration) ([]db.NearbyUser, error) {
	if m.nearbyFn != nil {
		return m.nearbyFn(ctx, sub, lat, lng, radiusKm, stale)
	}
	return nil, nil
}

func getNearby(sub string, repo nearbyUserRepository) *httptest.ResponseRecorder {
	r := withSubject(httptest.NewRequest(http.MethodGet, "/api/v1/users/nearby", nil), sub)
	w := httptest.NewRecorder()
	GetNearbyUsers(repo, 5, 60*time.Second).ServeHTTP(w, r)
	return w
}

func TestGetNearbyUsers_Success(t *testing.T) {
	repo := &mockNearbyRepo{
		nearbyFn: func(_ context.Context, _ string, _, _, _ float64, _ time.Duration) ([]db.NearbyUser, error) {
			return []db.NearbyUser{{ID: 2, FirstName: "Alice", LastName: "Smith"}}, nil
		},
	}
	w := getNearby("auth0|abc", repo)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body []map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if len(body) != 1 || body[0]["first_name"] != "Alice" {
		t.Errorf("unexpected body: %+v", body)
	}
}

func TestGetNearbyUsers_Empty(t *testing.T) {
	repo := &mockNearbyRepo{
		nearbyFn: func(_ context.Context, _ string, _, _, _ float64, _ time.Duration) ([]db.NearbyUser, error) {
			return []db.NearbyUser{}, nil
		},
	}
	w := getNearby("auth0|abc", repo)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetNearbyUsers_NoLocation(t *testing.T) {
	repo := &mockNearbyRepo{
		locationFn: func(_ context.Context, _ string) (float64, float64, error) {
			return 0, 0, db.ErrLocationNotFound
		},
	}
	w := getNearby("auth0|abc", repo)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	assertError(t, w, "location not found — push your location first")
}

func TestGetNearbyUsers_Unauthenticated(t *testing.T) {
	repo := &mockNearbyRepo{}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/nearby", nil)
	w := httptest.NewRecorder()
	GetNearbyUsers(repo, 5, 60*time.Second).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	assertError(t, w, "unauthorized")
}

func TestGetNearbyUsers_DBError(t *testing.T) {
	repo := &mockNearbyRepo{
		nearbyFn: func(_ context.Context, _ string, _, _, _ float64, _ time.Duration) ([]db.NearbyUser, error) {
			return nil, errors.New("connection refused")
		},
	}
	w := getNearby("auth0|abc", repo)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	assertError(t, w, "could not fetch nearby users")
}

func TestGetNearbyUsers_RadiusCap(t *testing.T) {
	var capturedRadius float64
	repo := &mockNearbyRepo{
		nearbyFn: func(_ context.Context, _ string, _, _, r float64, _ time.Duration) ([]db.NearbyUser, error) {
			capturedRadius = r
			return nil, nil
		},
	}
	r := withSubject(httptest.NewRequest(http.MethodGet, "/api/v1/users/nearby", nil), "auth0|abc")
	w := httptest.NewRecorder()
	GetNearbyUsers(repo, 50, 60*time.Second).ServeHTTP(w, r) // 50 > max 20

	if capturedRadius != maxNearbyRadiusKm {
		t.Errorf("expected radius capped at %f, got %f", maxNearbyRadiusKm, capturedRadius)
	}
}
