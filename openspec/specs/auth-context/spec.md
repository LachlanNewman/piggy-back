# auth-context Specification

## Purpose
TBD - created by archiving change setup-oidc-frontend. Update Purpose after archive.
## Requirements
### Requirement: AuthProvider wraps the app
The system SHALL provide an `AuthProvider` React component that owns the `UserManager` instance and makes auth state available to the entire component tree. `AuthProvider` SHALL be mounted in `main.jsx` as the outermost wrapper.

#### Scenario: App renders inside AuthProvider
- **WHEN** the React app mounts
- **THEN** all child components SHALL be able to consume auth state via `useAuth()` without directly importing `oidc-client-ts`

### Requirement: useAuth hook exposes auth state
The system SHALL export a `useAuth()` hook that returns `{ user, isAuthenticated, isLoading, login, logout }`.

- `user`: the OIDC `User` object (or `null` if unauthenticated)
- `isAuthenticated`: boolean derived from `user !== null && !user.expired`
- `isLoading`: boolean true while the initial session check is in progress
- `login()`: calls `UserManager.signinRedirect()`
- `logout()`: calls `UserManager.signoutRedirect()`

#### Scenario: Authenticated user
- **WHEN** a valid, non-expired token is present in sessionStorage
- **THEN** `useAuth()` SHALL return `isAuthenticated: true` and a populated `user` object

#### Scenario: Unauthenticated user
- **WHEN** no token is present or the token is expired
- **THEN** `useAuth()` SHALL return `isAuthenticated: false` and `user: null`

#### Scenario: Loading state on mount
- **WHEN** the app first mounts and is checking sessionStorage for an existing session
- **THEN** `isLoading` SHALL be `true` until the check completes, preventing a flash of the login screen for already-authenticated users

### Requirement: Unauthenticated users see login screen
The system SHALL render a login screen (with a login button) for any unauthenticated user attempting to access the app. Authenticated users SHALL see the main app content.

#### Scenario: Unauthenticated access
- **WHEN** `isAuthenticated` is `false` and `isLoading` is `false`
- **THEN** the app SHALL render a login prompt and SHALL NOT render the main app content

#### Scenario: Authenticated access
- **WHEN** `isAuthenticated` is `true`
- **THEN** the app SHALL render the main app content

#### Scenario: Auth loading
- **WHEN** `isLoading` is `true`
- **THEN** the app SHALL render a loading indicator and SHALL NOT render either the login screen or the main app content

