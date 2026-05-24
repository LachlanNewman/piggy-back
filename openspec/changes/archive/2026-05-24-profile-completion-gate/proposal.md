## Why

Users authenticated via SSO/OIDC may have an identity provider account but no app profile (weight, date of birth, gender). Without a gate, these users reach the app in an incomplete state. This change adds a post-login check that redirects incomplete profiles to a completion screen before granting access.

## What Changes

- `users` table gains `auth_subject` (IdP sub, unique identifier), `profile_complete` (boolean, stored, set by backend on write), and makes `date_of_birth`, `weight`, `gender` nullable (they are absent until the user completes the form)
- `POST /api/v1/users` accepts `auth_subject`, `first_name`, `last_name`, `email` from IdP claims alongside the 3 user-entered fields; sets `profile_complete = true` on creation
- New `GET /api/v1/users/me?sub=<sub>` endpoint returns profile existence and completion status (no JWT validation yet — secured in a future change)
- Frontend `App.jsx` gains a profile-status check after auth resolves; incomplete or missing profiles are gated behind `ProfileCompletionForm`
- `SignupForm` is replaced by `ProfileCompletionForm` — only collects `date_of_birth`, `weight`, `gender`; `first_name`, `last_name`, `email` are sourced from the IdP token claims

## Capabilities

### New Capabilities
- `profile-completion-gate`: Post-login gate that checks profile completeness and redirects to a completion form if the user's profile is absent or incomplete
- `user-profile-me`: `GET /api/v1/users/me` endpoint that returns a user's profile and completion status by `sub`

### Modified Capabilities
- `user-signup`: Users table schema changes (new columns, nullable columns); POST endpoint accepts `auth_subject` and drops client-supplied name/email in favour of IdP claims; `profile_complete` flag introduced

## Impact

- **DB**: Migration required — new columns on `users`, nullable constraint changes
- **Backend**: `db/users.go`, `handlers/users.go`, `cmd/api/main.go` — schema, handler signature, new GET handler
- **Frontend**: `App.jsx` — new profile-status fetch and gate logic; `SignupForm.jsx` replaced by `ProfileCompletionForm.jsx`
- **No breaking change to external consumers** — the endpoint path is the same; field additions are additive
