## Why

The current auth layer hand-rolls React state management on top of `oidc-client-ts` (`AuthContext.jsx`, `userManager.js`, `CallbackPage.jsx`). `react-oidc-context` is the community-standard wrapper that provides the same `AuthProvider` + `useAuth` surface with less code to own and test.

## What Changes

- Add `react-oidc-context` dependency; `oidc-client-ts` becomes a peer dep (still required by react-oidc-context internally)
- Delete `src/auth/AuthContext.jsx` — replaced by react-oidc-context's built-in `AuthProvider` and `useAuth`
- Delete `src/auth/userManager.js` — OIDC config moves to `AuthProvider` props in `main.jsx`
- Simplify `src/pages/CallbackPage.jsx` — react-oidc-context exposes `hasCodeInUrl` / `signinRedirectCallback` via hook, removing the need to import `userManager` directly
- Update `src/main.jsx` to wrap the app in react-oidc-context's `AuthProvider` with inline config
- Update `src/App.jsx` to use `signinRedirect` / `signoutRedirect` (react-oidc-context hook names)
- Retain `src/auth/splitTokenStore.js` unchanged — passed as `userStore` prop to `AuthProvider`

## Capabilities

### New Capabilities

- `oidc-auth`: OIDC authorization-code-flow authentication — login redirect, callback handling, token storage (refresh token in cookie, rest in memory), silent renew on reload, and logout redirect.

### Modified Capabilities

## Impact

- `frontend/package.json`: add `react-oidc-context`, keep `oidc-client-ts` (peer dep)
- `frontend/src/auth/`: `AuthContext.jsx` and `userManager.js` deleted; `splitTokenStore.js` kept
- `frontend/src/main.jsx`: new `AuthProvider` import and config
- `frontend/src/App.jsx`: `login` → `signinRedirect`, `logout` → `signoutRedirect`
- `frontend/src/pages/CallbackPage.jsx`: rewritten to use hook instead of direct userManager import
- No backend changes; no API contract changes; no env var changes
