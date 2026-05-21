## ADDED Requirements

### Requirement: Login redirect
The system SHALL initiate an OIDC authorization-code flow when the user triggers login, redirecting to the configured identity provider.

#### Scenario: User clicks login
- **WHEN** an unauthenticated user triggers the login action
- **THEN** the browser is redirected to the identity provider's authorization endpoint with `response_type=code`, `scope=openid profile email`, and the configured `redirect_uri`

### Requirement: Callback handling
The system SHALL complete the OIDC exchange when the identity provider redirects back to the app with an authorization code.

#### Scenario: Successful callback
- **WHEN** the browser lands on `/callback` with `code` and `state` query params
- **THEN** the app exchanges the code for tokens and navigates to `/`

#### Scenario: Failed callback
- **WHEN** the callback exchange fails (invalid code, state mismatch, network error)
- **THEN** the app displays an error message and does not navigate away

### Requirement: Split token storage
The system SHALL store the refresh token in a browser cookie and all other token data in memory only, to limit XSS exposure of the refresh token.

#### Scenario: Tokens written after login
- **WHEN** authentication completes
- **THEN** the `oidc_rt` cookie is set with the refresh token; the access token is not written to any persistent storage

#### Scenario: Refresh token persists across page reload
- **WHEN** the user reloads the page
- **THEN** the `oidc_rt` cookie is present and the app exchanges it for fresh tokens via silent renew

### Requirement: Session restore on reload
The system SHALL restore an authenticated session on page reload when a valid refresh token cookie exists.

#### Scenario: Valid refresh token cookie on reload
- **WHEN** the page loads and a valid `oidc_rt` cookie is present
- **THEN** the app performs a silent token refresh and enters the authenticated state

#### Scenario: No refresh token cookie on reload
- **WHEN** the page loads and no `oidc_rt` cookie is present
- **THEN** the app enters the unauthenticated state without making a token request

### Requirement: Logout redirect
The system SHALL end the session by clearing local token state and redirecting to the identity provider's end-session endpoint.

#### Scenario: User triggers logout
- **WHEN** an authenticated user triggers the logout action
- **THEN** local token state and the `oidc_rt` cookie are cleared, and the browser is redirected to the identity provider's end-session endpoint

### Requirement: Auth state exposed to components
The system SHALL expose authentication state to all React components via a hook without requiring components to interact with the OIDC library directly.

#### Scenario: Component reads auth state
- **WHEN** a component calls `useAuth()`
- **THEN** it receives `isAuthenticated`, `isLoading`, `user`, and actions to trigger login and logout

#### Scenario: Hook used outside provider
- **WHEN** `useAuth()` is called outside an `AuthProvider`
- **THEN** an error is thrown indicating incorrect usage
