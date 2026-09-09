package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
)

const (
	testAudience = "https://api.piggyback"
	testKeyID    = "test-key"
)

// newProviderServer stands in for the identity provider: it serves the OIDC
// discovery document and the JWKS the verifier fetches to check signatures.
func newProviderServer(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                srv.URL,
			"jwks_uri":                              srv.URL + "/.well-known/jwks.json",
			"authorization_endpoint":                srv.URL + "/authorize",
			"token_endpoint":                        srv.URL + "/oauth/token",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})

	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{
				Key:       key.Public(),
				KeyID:     testKeyID,
				Algorithm: string(jose.RS256),
				Use:       "sig",
			}},
		})
	})

	return srv
}

func signToken(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: testKeyID}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	sig, err := signer.Sign(payload)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	raw, err := sig.CompactSerialize()
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	return raw
}

func setup(t *testing.T) (*Authenticator, *rsa.PrivateKey, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	srv := newProviderServer(t, key)

	a, err := New(context.Background(), srv.URL, testAudience)
	if err != nil {
		t.Fatalf("new authenticator: %v", err)
	}
	return a, key, srv.URL
}

func validClaims(issuer, sub string) map[string]any {
	return map[string]any{
		"iss": issuer,
		"sub": sub,
		"aud": testAudience,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
}

// serve runs the middleware and reports the status plus the subject the wrapped
// handler saw.
func serve(a *Authenticator, authzHeader string) (int, string) {
	var gotSub string
	h := a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSub, _ = SubjectFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	if authzHeader != "" {
		r.Header.Set("Authorization", authzHeader)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code, gotSub
}

func TestMiddleware_ValidToken(t *testing.T) {
	a, key, issuer := setup(t)
	token := signToken(t, key, validClaims(issuer, "auth0|abc123"))

	code, sub := serve(a, "Bearer "+token)

	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if sub != "auth0|abc123" {
		t.Errorf("expected subject from token, got %q", sub)
	}
}

func TestMiddleware_BearerSchemeIsCaseInsensitive(t *testing.T) {
	a, key, issuer := setup(t)
	token := signToken(t, key, validClaims(issuer, "auth0|abc123"))

	if code, _ := serve(a, "bearer "+token); code != http.StatusOK {
		t.Errorf("expected 200 for lowercase scheme, got %d", code)
	}
}

func TestMiddleware_Rejects(t *testing.T) {
	a, key, issuer := setup(t)

	expired := validClaims(issuer, "auth0|abc123")
	expired["exp"] = time.Now().Add(-time.Hour).Unix()

	wrongAudience := validClaims(issuer, "auth0|abc123")
	wrongAudience["aud"] = "https://api.somewhere-else"

	wrongIssuer := validClaims("https://evil.example.com", "auth0|abc123")

	noSubject := validClaims(issuer, "")
	delete(noSubject, "sub")

	// A token signed by a key the provider never published: the shape is right,
	// only the signature is wrong.
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	tests := []struct {
		name   string
		header string
	}{
		{"no header", ""},
		{"empty bearer", "Bearer "},
		{"wrong scheme", "Basic abc123"},
		{"not a jwt", "Bearer not-a-jwt"},
		{"expired", "Bearer " + signToken(t, key, expired)},
		{"wrong audience", "Bearer " + signToken(t, key, wrongAudience)},
		{"wrong issuer", "Bearer " + signToken(t, key, wrongIssuer)},
		{"no subject", "Bearer " + signToken(t, key, noSubject)},
		{"unknown signing key", "Bearer " + signToken(t, otherKey, validClaims(issuer, "auth0|abc123"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, sub := serve(a, tt.header)
			if code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", code)
			}
			if sub != "" {
				t.Errorf("handler must not run, but saw subject %q", sub)
			}
		})
	}
}

func TestNew_RequiresIssuerAndAudience(t *testing.T) {
	if _, err := New(context.Background(), "", testAudience); err == nil {
		t.Error("expected an error when the issuer is empty")
	}
	if _, err := New(context.Background(), "https://example.com", ""); err == nil {
		t.Error("expected an error when the audience is empty")
	}
}

func TestSubjectFromContext(t *testing.T) {
	if _, ok := SubjectFromContext(context.Background()); ok {
		t.Error("expected no subject on a bare context")
	}
	if _, ok := SubjectFromContext(ContextWithSubject(context.Background(), "")); ok {
		t.Error("expected an empty subject to report false")
	}
	sub, ok := SubjectFromContext(ContextWithSubject(context.Background(), "auth0|abc"))
	if !ok || sub != "auth0|abc" {
		t.Errorf("expected round trip, got %q ok=%v", sub, ok)
	}
}

func TestUnverifiedClaims(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	t.Run("audience as a single string", func(t *testing.T) {
		raw := signToken(t, key, map[string]any{"iss": "https://issuer/", "aud": "https://api.piggyback"})
		iss, aud, ok := unverifiedClaims(raw)
		if !ok {
			t.Fatal("expected the payload to decode")
		}
		if iss != "https://issuer/" {
			t.Errorf("iss = %q", iss)
		}
		if len(aud) != 1 || aud[0] != "https://api.piggyback" {
			t.Errorf("aud = %v", aud)
		}
	})

	t.Run("audience as an array", func(t *testing.T) {
		raw := signToken(t, key, map[string]any{
			"iss": "https://issuer/",
			"aud": []string{"https://api.piggyback", "https://issuer/userinfo"},
		})
		_, aud, ok := unverifiedClaims(raw)
		if !ok {
			t.Fatal("expected the payload to decode")
		}
		if len(aud) != 2 || aud[0] != "https://api.piggyback" {
			t.Errorf("aud = %v", aud)
		}
	})

	// An Auth0 opaque access token: not a JWT, so there is nothing to decode.
	// This is the case the debug log calls out by name.
	t.Run("opaque token", func(t *testing.T) {
		for _, raw := range []string{"", "not-a-jwt", "aB3xY9zQ7wE1rT5yU8iO0pL2kJ6hG4fD"} {
			if _, _, ok := unverifiedClaims(raw); ok {
				t.Errorf("expected %q not to decode as a JWT", raw)
			}
		}
	})

	t.Run("does not verify the signature", func(t *testing.T) {
		other, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}
		raw := signToken(t, other, map[string]any{"iss": "https://evil/", "aud": "https://spoofed"})
		_, aud, ok := unverifiedClaims(raw)
		if !ok || len(aud) != 1 || aud[0] != "https://spoofed" {
			t.Errorf("expected claims to decode regardless of signer, got %v ok=%v", aud, ok)
		}
	})
}
