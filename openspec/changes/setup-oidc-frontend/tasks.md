## 1. Dependencies & Configuration

- [x] 1.1 Install `oidc-client-ts` via npm in `frontend/`
- [x] 1.2 Add `VITE_OIDC_AUTHORITY`, `VITE_OIDC_CLIENT_ID`, and `VITE_OIDC_REDIRECT_URI` to `frontend/.env.example`
- [x] 1.3 Validate that all three env vars are defined at app startup; throw a descriptive error if any are missing

## 2. OIDC Auth Module

- [x] 2.1 Create `frontend/src/auth/userManager.js` — instantiate and export a singleton `UserManager` configured from env vars with `response_type: "code"` and sessionStorage
- [x] 2.2 Verify `UserManager` config: authority, client_id, redirect_uri, scope, response_type

## 3. Auth Context

- [x] 3.1 Create `frontend/src/auth/AuthContext.jsx` — React context with `{ user, isAuthenticated, isLoading, login, logout }`
- [x] 3.2 Implement `AuthProvider` component: load existing user from `UserManager.getUser()` on mount, set `isLoading` to false after check completes, subscribe to `UserManager` events for user loaded/unloaded
- [x] 3.3 Export `useAuth()` hook that reads from `AuthContext`; throw if used outside `AuthProvider`

## 4. Callback Page

- [x] 4.1 Create `frontend/src/pages/CallbackPage.jsx` — call `UserManager.signinRedirectCallback()` on mount, redirect to `/` on success, display error message on failure

## 5. Routing & App Integration

- [x] 5.1 Add a client-side router (react-router-dom) if not already present, or use the existing router
- [x] 5.2 Register `/callback` route pointing to `CallbackPage`
- [x] 5.3 Wrap the app in `AuthProvider` in `main.jsx`
- [x] 5.4 Update `App.jsx` to consume `useAuth()`: render loading indicator when `isLoading`, render login button (calls `login()`) when unauthenticated, render main app content when authenticated

## 6. Login & Logout UI

- [x] 6.1 Add a login button component (or inline in `App.jsx`) that calls `login()` from `useAuth()`
- [x] 6.2 Add a logout button visible when authenticated that calls `logout()` from `useAuth()`
