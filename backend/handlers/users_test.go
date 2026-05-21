package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/db"
)

type mockUserRepo struct {
	fn func(ctx context.Context, p db.CreateUserParams) (int, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, p db.CreateUserParams) (int, error) {
	return m.fn(ctx, p)
}

func post(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestCreateUser_Success(t *testing.T) {
	repo := &mockUserRepo{fn: func(_ context.Context, _ db.CreateUserParams) (int, error) {
		return 42, nil
	}}

	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"first_name":"Jane","last_name":"Doe","email":"jane@example.com",
		"date_of_birth":"1995-06-15","weight":68.5,"gender":"female"
	}`))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp createUserResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ID != 42 {
		t.Errorf("expected id 42, got %d", resp.ID)
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	repo := &mockUserRepo{}
	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`not json`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "invalid JSON")
}

func TestCreateUser_MissingField(t *testing.T) {
	repo := &mockUserRepo{}
	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"last_name":"Doe","email":"jane@example.com",
		"date_of_birth":"1995-06-15","weight":68.5,"gender":"female"
	}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "first_name is required") // json tag name via RegisterTagNameFunc
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	repo := &mockUserRepo{}
	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"first_name":"Jane","last_name":"Doe","email":"not-an-email",
		"date_of_birth":"1995-06-15","weight":68.5,"gender":"female"
	}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "email must be a valid email address")
}

func TestCreateUser_InvalidDateFormat(t *testing.T) {
	repo := &mockUserRepo{}
	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"first_name":"Jane","last_name":"Doe","email":"jane@example.com",
		"date_of_birth":"15-06-1995","weight":68.5,"gender":"female"
	}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "date_of_birth must be in YYYY-MM-DD format")
}

func TestCreateUser_InvalidGender(t *testing.T) {
	repo := &mockUserRepo{}
	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"first_name":"Jane","last_name":"Doe","email":"jane@example.com",
		"date_of_birth":"1995-06-15","weight":68.5,"gender":"other"
	}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "gender must be one of: male, female, unknown")
}

func TestCreateUser_WeightNotPositive(t *testing.T) {
	repo := &mockUserRepo{}
	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"first_name":"Jane","last_name":"Doe","email":"jane@example.com",
		"date_of_birth":"1995-06-15","weight":0,"gender":"female"
	}`))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertError(t, w, "weight must be greater than 0")
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{fn: func(_ context.Context, _ db.CreateUserParams) (int, error) {
		return 0, db.ErrDuplicateEmail
	}}

	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"first_name":"Jane","last_name":"Doe","email":"jane@example.com",
		"date_of_birth":"1995-06-15","weight":68.5,"gender":"female"
	}`))

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	assertError(t, w, "email already registered")
}

func TestCreateUser_DBError(t *testing.T) {
	repo := &mockUserRepo{fn: func(_ context.Context, _ db.CreateUserParams) (int, error) {
		return 0, errors.New("connection refused")
	}}

	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, post(`{
		"first_name":"Jane","last_name":"Doe","email":"jane@example.com",
		"date_of_birth":"1995-06-15","weight":68.5,"gender":"female"
	}`))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	assertError(t, w, "could not create user")
}

func TestCreateUser_MethodNotAllowed(t *testing.T) {
	repo := &mockUserRepo{}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	CreateUser(repo).ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func assertError(t *testing.T, w *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != want {
		t.Errorf("expected error %q, got %q", want, body["error"])
	}
}
