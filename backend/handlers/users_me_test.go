package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/db"
)

type mockUserProfileRepo struct {
	fn func(ctx context.Context, sub string) (db.User, error)
}

func (m *mockUserProfileRepo) GetUserBySubject(ctx context.Context, sub string) (db.User, error) {
	return m.fn(ctx, sub)
}

func getMe(sub string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me?sub="+sub, nil)
	return r
}

func TestGetUserMe_Found(t *testing.T) {
	want := db.User{
		ID: 3, AuthSubject: "auth0|abc", FirstName: "Jane", LastName: "Doe",
		Email: "jane@example.com", ProfileComplete: true,
	}
	repo := &mockUserProfileRepo{fn: func(_ context.Context, _ string) (db.User, error) {
		return want, nil
	}}

	w := httptest.NewRecorder()
	GetUserMe(repo).ServeHTTP(w, getMe("auth0|abc"))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp getUserMeResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ID != want.ID || resp.AuthSubject != want.AuthSubject || !resp.ProfileComplete {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetUserMe_NotFound(t *testing.T) {
	repo := &mockUserProfileRepo{fn: func(_ context.Context, _ string) (db.User, error) {
		return db.User{}, db.ErrNotFound
	}}

	w := httptest.NewRecorder()
	GetUserMe(repo).ServeHTTP(w, getMe("auth0|unknown"))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	assertError(t, w, "user not found")
}

func TestGetUserMe_MissingSub(t *testing.T) {
	repo := &mockUserProfileRepo{}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	w := httptest.NewRecorder()
	GetUserMe(repo).ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "sub is required")
}

func TestGetUserMe_DBError(t *testing.T) {
	repo := &mockUserProfileRepo{fn: func(_ context.Context, _ string) (db.User, error) {
		return db.User{}, errors.New("connection refused")
	}}

	w := httptest.NewRecorder()
	GetUserMe(repo).ServeHTTP(w, getMe("auth0|abc"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	assertError(t, w, "could not fetch user")
}

func TestGetUserMe_MethodNotAllowed(t *testing.T) {
	repo := &mockUserProfileRepo{}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/users/me", nil)
	w := httptest.NewRecorder()
	GetUserMe(repo).ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
