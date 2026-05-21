## Context

The frontend currently has no authentication. Users access the app anonymously. We need to add OIDC login using the Authorization Code + PKCE flow — the correct browser-native pattern that avoids exposing client secrets. This phase is frontend-only; the backend does not validate tokens yet.

## Goals / Non-Goals

**Goals:**
- Implement Authorization Code + PKCE flow using `oidc-client-ts`
- Handle the post-login redirect callback and store tokens
- Expose auth state (user, isAuthenticated, login, logout) via React context
- Gate the app: unauthenticated users see a login screen, authenticated users see the app
- Make OIDC provider configurable via Vite env vars (authority, client_id, redirect_uri)

**Non-Goals:**
- Backend token validation (deferred)
- Refresh token rotation (handled by `oidc-client-ts` silently)
- Role-based access control
- Supporting implicit flow or client credentials
- Specific provider setup instructions (Google, Okta, Auth0, etc.)

## Decisions

### Use `oidc-client-ts` over a raw fetch-based implementation

`oidc-client-ts` implements PKCE, state/nonce verification, silent renew, and UserManager lifecycle correctly. Rolling this from scratch introduces security surface area. The library is actively maintained, has TypeScript types, and is the community successor to `oidc-client`.

Alternatives considered:
- **`@auth0/auth0-react`** — provider-locked, ruled out
- **`react-oidc-context`** — thin wrapper over `oidc-client-ts`, adds a React context layer we'd need to customise anyway; adding our own context over `oidc-client-ts` directly gives us more control
- **Manual fetch** — too much security complexity to get right

### Store tokens in sessionStorage (via `oidc-client-ts` default)

`oidc-client-ts` defaults to sessionStorage for token storage. This means tokens are cleared on tab close, reducing XSS persistence risk. We accept the UX trade-off (re-login after browser restart) for this phase.

Alternatives considered:
- **In-memory only** — survives no navigation; incompatible with PKCE redirect flow
- **localStorage** — persists tokens across sessions, increases XSS blast radius

### Configuration via Vite env vars

Provider authority, client ID, and redirect URI are injected at build time via `import.meta.env.VITE_OIDC_*`. This keeps provider details out of source and supports different environments (dev/prod) via `.env` files.

### Auth context wraps the app at root level

A single `AuthProvider` in `main.jsx` owns the `UserManager` instance and exposes `{ user, isAuthenticated, isLoading, login, logout }`. Components never import `oidc-client-ts` directly — they consume context only.

### Callback handling via a dedicated `/callback` route

The OIDC redirect lands at `/callback`. A `CallbackPage` component calls `UserManager.signinRedirectCallback()` then redirects to `/`. This keeps callback logic isolated and avoids polluting `App.jsx`.

## Risks / Trade-offs

- **No backend validation** → Tokens are trusted client-side only. A malicious actor who can XSS the app can read the session. Mitigation: sessionStorage (not localStorage) limits persistence; backend validation is the next change.
- **Provider configuration errors at runtime** → Wrong `VITE_OIDC_AUTHORITY` or `VITE_OIDC_CLIENT_ID` will cause silent failures on redirect. Mitigation: validate required env vars at startup and throw a clear error.
- **Redirect URI mismatch** → OIDC providers reject unregistered redirect URIs. Mitigation: document that `VITE_OIDC_REDIRECT_URI` must match exactly what is registered in the provider.

## Open Questions

- Should silent token renewal be enabled in this phase? (`automaticSilentRenew: true` requires a `silent_redirect.html` iframe endpoint.) Default: disable for now, enable in a follow-up.
