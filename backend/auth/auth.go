// Package auth verifies OIDC access tokens and carries the authenticated
// subject through the request context.
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type contextKey struct{}

var subjectKey contextKey

// Authenticator validates bearer tokens against the provider's JWKS.
type Authenticator struct {
	verifier *oidc.IDTokenVerifier
	issuer   string
	audience string
}

// New discovers the provider at issuer and returns an Authenticator that
// accepts tokens minted for audience. The discovery request is made here, so a
// misconfigured issuer fails at startup rather than on the first request.
func New(ctx context.Context, issuer, audience string) (*Authenticator, error) {
	if issuer == "" {
		return nil, fmt.Errorf("OIDC_ISSUER not set")
	}
	if audience == "" {
		return nil, fmt.Errorf("OIDC_AUDIENCE not set")
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discover issuer %q: %w", issuer, err)
	}

	return &Authenticator{
		verifier: provider.Verifier(&oidc.Config{ClientID: audience}),
		issuer:   issuer,
		audience: audience,
	}, nil
}

// Middleware rejects requests without a valid bearer token and stores the
// token's subject in the request context for downstream handlers.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, ok := bearerToken(r)
		if !ok {
			unauthorized(w, "missing bearer token")
			return
		}

		a.logClaims(r)

		token, err := a.verifier.Verify(r.Context(), raw)
		if err != nil {
			// The client deliberately gets a generic message, but without this
			// log an opaque token, a wrong audience, and an expired one are
			// indistinguishable from the outside.
			slog.Warn("token verification failed", "path", r.URL.Path, "err", err)
			unauthorized(w, "invalid token")
			return
		}
		if token.Subject == "" {
			unauthorized(w, "token has no subject")
			return
		}

		next.ServeHTTP(w, r.WithContext(ContextWithSubject(r.Context(), token.Subject)))
	})
}

// logClaims reports what the token actually carries versus what this service
// expects, which is the fastest way to spot an audience or issuer mismatch. The
// claims are read WITHOUT verifying the signature, so this is gated behind
// debug and its output must never be trusted for authorization.
func (a *Authenticator) logClaims(r *http.Request) {
	if !slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		return
	}

	raw, ok := bearerToken(r)
	if !ok {
		return
	}

	iss, aud, ok := unverifiedClaims(raw)
	if !ok {
		slog.Debug("bearer token is not a JWT; Auth0 returns an opaque token when the client requests no audience",
			"path", r.URL.Path, "expected_aud", a.audience)
		return
	}

	slog.Debug("token claims before verification (unverified)",
		"path", r.URL.Path,
		"aud", aud, "expected_aud", a.audience, "aud_ok", slices.Contains(aud, a.audience),
		"iss", iss, "expected_iss", a.issuer, "iss_ok", iss == a.issuer,
	)
}

// audience is the "aud" claim, which the spec allows to be either a single
// string or an array of them.
type audience []string

func (a *audience) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*a = audience{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

// unverifiedClaims decodes the JWT payload without checking the signature. It
// exists only to explain failures in debug logs.
func unverifiedClaims(raw string) (iss string, aud []string, ok bool) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return "", nil, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, false
	}
	var claims struct {
		Issuer   string   `json:"iss"`
		Audience audience `json:"aud"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", nil, false
	}
	return claims.Issuer, claims.Audience, true
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "bearer") || token == "" {
		return "", false
	}
	return token, true
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ContextWithSubject returns ctx carrying sub as the authenticated subject.
func ContextWithSubject(ctx context.Context, sub string) context.Context {
	return context.WithValue(ctx, subjectKey, sub)
}

// SubjectFromContext returns the subject stored by Middleware. It reports false
// when the request did not pass through Middleware.
func SubjectFromContext(ctx context.Context) (string, bool) {
	sub, ok := ctx.Value(subjectKey).(string)
	return sub, ok && sub != ""
}
