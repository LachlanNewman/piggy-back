## 1. Dependency

- [x] 1.1 Add `react-oidc-context` to `frontend/package.json` and run `npm install`

## 2. Provider Setup

- [x] 2.1 Rewrite `frontend/src/main.jsx` — import `AuthProvider` from `react-oidc-context`, validate required env vars, construct `new SplitTokenStore()`, and pass OIDC config (`authority`, `client_id`, `redirect_uri`, `response_type`, `scope`, `userStore`, `automaticSilentRenew: false`, `onSigninCallback`) as props to `AuthProvider`
- [x] 2.2 Delete `frontend/src/auth/userManager.js`
- [x] 2.3 Delete `frontend/src/auth/AuthContext.jsx`

## 3. Callback Page

- [x] 3.1 Rewrite `frontend/src/pages/CallbackPage.jsx` — use `useAuth()` from `react-oidc-context`; watch `isLoading` and `activeNavigator` to detect completion; navigate to `/` on success; display error from `error` field on failure

## 4. App Component

- [x] 4.1 Update `frontend/src/App.jsx` — change `useAuth` import to `react-oidc-context`; rename `login` → `signinRedirect`, `logout` → `signoutRedirect` at callsites

## 5. Verification

- [x] 5.1 Start services with `docker-compose up --build` and confirm the app loads without console errors
- [x] 5.2 Complete a full login flow end-to-end: unauthenticated state → login redirect → callback → authenticated state with user profile displayed
- [x] 5.3 Reload the page while authenticated and confirm session is restored via silent renew (refresh token cookie)
- [x] 5.4 Trigger logout and confirm local state clears and browser redirects to identity provider end-session
