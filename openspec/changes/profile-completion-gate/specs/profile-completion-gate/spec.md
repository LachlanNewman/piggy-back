## ADDED Requirements

### Requirement: Profile completion gate in the frontend
After authentication resolves, the frontend SHALL check whether the authenticated user has a complete profile before rendering the application. If the profile is absent or incomplete, the user SHALL be shown a profile completion form instead of the application.

#### Scenario: Authenticated user with complete profile
- **WHEN** auth resolves and `GET /api/v1/users/me?sub=<sub>` returns 200 with `profile_complete: true`
- **THEN** the app SHALL render the main application view

#### Scenario: Authenticated user with no profile row
- **WHEN** auth resolves and `GET /api/v1/users/me?sub=<sub>` returns 404
- **THEN** the app SHALL render the `ProfileCompletionForm` component and SHALL NOT render the main application view

#### Scenario: Authenticated user with incomplete profile
- **WHEN** auth resolves and `GET /api/v1/users/me?sub=<sub>` returns 200 with `profile_complete: false`
- **THEN** the app SHALL render the `ProfileCompletionForm` component and SHALL NOT render the main application view

#### Scenario: Profile check in progress
- **WHEN** auth has resolved but the profile status fetch has not yet completed
- **THEN** the app SHALL render a loading state

#### Scenario: Session restored silently (refresh token path)
- **WHEN** `signinSilent()` completes and sets `isAuthenticated` to true without a full OIDC redirect
- **THEN** the profile gate check SHALL fire identically to the full login flow

### Requirement: ProfileCompletionForm collects biometric fields only
The `ProfileCompletionForm` component SHALL collect only `date_of_birth`, `weight`, and `gender` from the user. `first_name`, `last_name`, and `email` SHALL be sourced from the IdP token claims (`user.profile`) and SHALL NOT be presented as editable inputs.

#### Scenario: Form renders with correct fields
- **WHEN** `ProfileCompletionForm` is rendered
- **THEN** it SHALL display inputs for `date_of_birth`, `weight`, and `gender` only

#### Scenario: Successful profile completion
- **WHEN** the user fills in all three fields and submits the form
- **THEN** the form SHALL POST to `/api/v1/users` with `auth_subject`, `first_name`, `last_name`, and `email` sourced from `user.profile`, and `date_of_birth`, `weight`, `gender` from the form inputs
- **THEN** on a 201 response the gate SHALL recheck profile status and transition to the main application view

#### Scenario: Validation error from backend
- **WHEN** the backend returns 400
- **THEN** the form SHALL display the `error` field from the response body without clearing the form

#### Scenario: Submit while request is in flight
- **WHEN** the user submits and the request has not yet resolved
- **THEN** the submit button SHALL be disabled until the request completes
