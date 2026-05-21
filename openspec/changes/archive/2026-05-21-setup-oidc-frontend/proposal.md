## Why

The frontend has no authentication layer. Adding OIDC support lets users log in via any standards-compliant identity provider without coupling the app to a specific vendor.

## What Changes

- Add an OIDC client library to the frontend (`oidc-client-ts` or similar)
- Implement the Authorization Code + PKCE flow in the browser
- Add a login page / login button that redirects to the provider's authorization endpoint
- Handle the callback redirect and extract tokens
- Store the session (access token, id token, refresh token) in memory or sessionStorage
- Expose auth state (user info, login/logout) via a React context
- Gate the main app behind authentication (redirect to login if unauthenticated)
- No backend changes — token validation is deferred to a future change

## Capabilities

### New Capabilities
- `oidc-auth`: Provider-agnostic OIDC Authorization Code + PKCE flow; login, callback handling, session storage, and logout
- `auth-context`: React context exposing current user, auth state, and login/logout actions to the component tree

### Modified Capabilities
<!-- None — no existing specs have requirement changes -->

## Impact

- **New dependency**: `oidc-client-ts` (browser OIDC/OAuth2 client, actively maintained W3C-aligned library)
- **Frontend files affected**: `main.jsx`, `App.jsx`, new `src/auth/` directory
- **Configuration**: OIDC provider URL, client ID, and redirect URI must be supplied via environment variables (`.env` / Vite `import.meta.env`)
- **No backend changes** in this phase; access tokens are held client-side only
