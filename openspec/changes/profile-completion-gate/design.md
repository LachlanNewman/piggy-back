## Context

Users authenticate via Auth0 (OIDC). The IdP holds identity (sub, email, given_name, family_name). The app holds biometric profile data (date of birth, weight, gender). Currently these two worlds are disconnected: the `users` table has no link to the IdP identity, and the signup form is shown unconditionally after login regardless of whether a profile already exists.

The app also performs silent session restore via `signinSilent()` when a refresh token cookie is present, meaning authenticated users may never pass through `CallbackPage`. The profile gate must therefore live at the app render level, not in the callback handler.

## Goals / Non-Goals

**Goals:**
- Link every `users` row to an IdP identity via `auth_subject` (the OIDC `sub` claim)
- Introduce a `profile_complete` boolean that explicitly signals whether a user's profile satisfies the app's current requirements
- Gate app access on profile completeness after auth resolves — regardless of whether the session was restored silently or via full OIDC flow
- Trim the profile completion form to only the fields the user must provide (DOB, weight, gender); source name and email from IdP claims

**Non-Goals:**
- JWT validation on the backend (secured in a future change; `sub` is passed as a query param for now)
- Full `/me` profile retrieval (only `profile_complete` and existence are consumed by this change)
- Editing an existing complete profile

## Decisions

### Identity link: `auth_subject` (sub) not email
Email is mutable and can be shared across federation scenarios. The OIDC `sub` is stable and unique per IdP. All user lookups use `auth_subject`.

*Alternatives considered*: match by email — rejected because emails can change and two IdP accounts could share an email.

### `profile_complete` as a stored boolean, set on write
`profile_complete` is explicitly set to `true` by the backend when the completion form is submitted. It is not derived from field presence on each read.

This design enables a future operational workflow: when a new required field is added, run `UPDATE users SET profile_complete = false` to re-prompt all users on next login. A purely derived flag would not support this — existing rows would auto-compute as complete even without the new field.

### Nullable biometric columns
`date_of_birth`, `weight`, and `gender` become nullable. A user row is created on first login (sourced entirely from IdP claims) with these fields null and `profile_complete = false`. The completion form populates them and sets `profile_complete = true`.

*Alternatives considered*: create the row only on completion form submit — rejected because it complicates the existence check (404 = new user, 200+incomplete = returning incomplete user are two different states that need different handling).

### Frontend-side profile check
After auth resolves, `App.jsx` calls `GET /api/v1/users/me?sub=<sub>`. The response drives the gate:
- 404 → no row yet, show completion form (POST will create)
- 200 + `profile_complete: false` → row exists but incomplete, show completion form (POST will upsert or a PATCH will complete)
- 200 + `profile_complete: true` → show app

The check runs on every render cycle until profile is confirmed complete, covering both full OIDC flow and silent session restore.

### Single POST endpoint handles both create and upsert
`POST /api/v1/users` is extended to accept `auth_subject`. On first submit it inserts; if a row already exists for that `auth_subject` (rare race condition or retry), it updates the biometric fields and sets `profile_complete = true`. This avoids a separate PATCH endpoint for the completion flow.

## Risks / Trade-offs

- **No JWT validation** → `sub` is caller-supplied; any client can claim any identity. Accepted as a known interim state. Mitigated by adding JWT middleware in a follow-up change.
- **Nullable biometric columns** → existing queries that assume non-null DOB/weight/gender will need null handling. No such queries exist yet.
- **profile_complete drift** → if a bug prevents `profile_complete` being set to `true` on write, users are re-prompted indefinitely. Mitigated: the completion form re-submits idempotently.

## Migration Plan

1. Run DB migration: add `auth_subject`, add `profile_complete DEFAULT false`, alter `date_of_birth`/`weight`/`gender` to nullable
2. Existing rows (if any) will have `auth_subject = NULL` — violates the new NOT NULL constraint. Pre-migration: backfill or truncate. For dev/staging: truncate is acceptable. For production: backfill `auth_subject` from known data before adding the NOT NULL constraint.
3. Deploy backend with new handler
4. Deploy frontend with gate

Rollback: revert frontend deploy (gate disappears); backend endpoint changes are additive and do not break existing callers.
