## Context

The frontend currently uses `oidc-client-ts` directly with three hand-rolled files:
- `userManager.js` — creates a `UserManager` singleton with env-var config
- `AuthContext.jsx` — React context that wires UserManager events to state, exposes `useAuth`
- `CallbackPage.jsx` — manually calls `userManager.signinRedirectCallback()` after redirect

`react-oidc-context` is a thin React wrapper over `oidc-client-ts` that ships exactly this plumbing as a library. Migrating removes ~80 lines of owned code with no behavioral change.

The `SplitTokenStore` (refresh token in `HttpOnly`-like cookie, access token in memory) is a project-specific security constraint that must survive the migration unchanged.

## Goals / Non-Goals

**Goals:**
- Replace `AuthContext.jsx` and `userManager.js` with react-oidc-context's `AuthProvider` + `useAuth`
- Simplify `CallbackPage.jsx` using the library's callback handling
- Preserve `SplitTokenStore` behavior exactly
- Keep the public API seen by `App.jsx` functionally identical (authenticated state, login, logout, user profile)

**Non-Goals:**
- Changing the OIDC flow (authorization code, same scopes, same redirect URI)
- Modifying `SplitTokenStore` internals
- Adding silent-renew background polling (not in scope; `automaticSilentRenew: false` stays)
- Any backend changes

## Decisions

### 1. Pass `userStore` prop directly to `AuthProvider`

react-oidc-context's `AuthProvider` accepts all `UserManagerSettings` as props, including `userStore`. We pass `new SplitTokenStore()` here rather than constructing a separate `UserManager`.

**Alternative**: Pre-construct a `UserManager` and pass it via the `userManager` prop. Rejected — the prop-based approach is simpler and avoids keeping `userManager.js` alive.

### 2. Use `withAuthenticationRequired` / `hasCodeInUrl` for callback

react-oidc-context provides `hasCodeInUrl` utility and handles the `signinRedirectCallback` internally when the provider mounts on the callback route. `CallbackPage.jsx` can use `useAuth` to read `activeNavigator` / `isLoading` and redirect when done, rather than calling `userManager` directly.

Concretely: wrap `CallbackPage` in a `useEffect` that watches `isLoading` and `isAuthenticated` to navigate away when auth completes.

**Alternative**: Keep explicit `signinRedirectCallback()` call using the `userManager` exposed by `useAuth().user?._userManager` (private API). Rejected — fragile, couples to internals.

### 3. Map `login`/`logout` aliases in `App.jsx` locally

react-oidc-context's hook returns `signinRedirect` and `signoutRedirect`. Rather than aliasing in a wrapper, update `App.jsx` callsites directly to use the library names.

### 4. Env-var validation stays in `main.jsx`

The validation guard currently in `userManager.js` moves to `main.jsx` before the `AuthProvider` renders, so startup still fails fast on missing config.

## Risks / Trade-offs

- **react-oidc-context re-render behavior** — The library re-renders consumers on every auth event. Current hand-rolled context does the same, so no regression expected. → Mitigation: verify no extra renders break `useEffect` dependencies in `App.jsx`.
- **`onSigninCallback` default behavior** — By default react-oidc-context calls `history.replaceState` after callback to strip the code/state params from the URL. This replaces the manual `navigate('/', { replace: true })` in `CallbackPage`. Confirm this is sufficient or provide a custom `onSigninCallback` prop. → Mitigation: provide explicit `onSigninCallback={() => navigate('/')}` for clarity.
- **`automaticSilentRenew`** — Must stay `false` to match current behavior; the library defaults to `true`. → Mitigation: explicitly set in `AuthProvider` props.

## Migration Plan

1. `npm install react-oidc-context` in `frontend/`
2. Rewrite `main.jsx` — replace `AuthProvider` import and wrap with react-oidc-context's provider
3. Delete `src/auth/AuthContext.jsx` and `src/auth/userManager.js`
4. Rewrite `CallbackPage.jsx` — use `useAuth` hook
5. Update `App.jsx` — `login` → `signinRedirect`, `logout` → `signoutRedirect`
6. Smoke-test login flow end-to-end in Docker Compose

Rollback: revert the five file changes; `react-oidc-context` can be left in `package.json` harmlessly or removed.
