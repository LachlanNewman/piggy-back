## ADDED Requirements

### Requirement: PKCE login flow
The system SHALL initiate an OIDC Authorization Code + PKCE login by redirecting the user to the provider's authorization endpoint when `login()` is called. The `oidc-client-ts` `UserManager` SHALL be configured with `response_type: "code"` and PKCE enabled.

#### Scenario: User initiates login
- **WHEN** the user clicks the login button
- **THEN** the browser SHALL redirect to the OIDC provider's authorization endpoint with a PKCE code challenge and a `state` parameter

#### Scenario: Provider configuration is missing
- **WHEN** any of `VITE_OIDC_AUTHORITY`, `VITE_OIDC_CLIENT_ID`, or `VITE_OIDC_REDIRECT_URI` is undefined at startup
- **THEN** the app SHALL throw an error before rendering, with a message identifying the missing variable

### Requirement: Callback handling
The system SHALL process the OIDC redirect callback at the `/callback` route by calling `UserManager.signinRedirectCallback()`, which validates the `state`, verifies the PKCE code verifier, and exchanges the authorization code for tokens.

#### Scenario: Successful callback
- **WHEN** the provider redirects to `/callback` with a valid `code` and `state`
- **THEN** the app SHALL exchange the code for tokens, store them in sessionStorage, and redirect the user to `/`

#### Scenario: Invalid or tampered callback
- **WHEN** the callback contains an invalid `state` or missing `code`
- **THEN** the app SHALL display an error message and SHALL NOT store any tokens

### Requirement: Logout
The system SHALL end the user's session by calling `UserManager.signoutRedirect()`, which clears local token storage and redirects to the provider's end_session endpoint.

#### Scenario: User logs out
- **WHEN** the user clicks the logout button
- **THEN** local tokens SHALL be cleared and the browser SHALL redirect to the OIDC provider's logout endpoint

### Requirement: Token storage in sessionStorage
The system SHALL store tokens in sessionStorage (the `oidc-client-ts` default). Tokens SHALL NOT be written to localStorage.

#### Scenario: Token persists within session
- **WHEN** the user navigates between pages within the same browser tab
- **THEN** the user SHALL remain authenticated without re-initiating the login flow

#### Scenario: Tokens cleared on tab close
- **WHEN** the user closes the browser tab or window
- **THEN** tokens SHALL be cleared and a new login SHALL be required on next visit
