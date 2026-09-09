// Package auth verifies OIDC access tokens and carries the authenticated
// subject through the request context.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type contextKey struct{}

var subjectKey contextKey

// Authenticator validates bearer tokens against the provider's JWKS.
type Authenticator struct {
	verifier *oidc.IDTokenVerifier
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

		token, err := a.verifier.Verify(r.Context(), raw)
		if err != nil {
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
