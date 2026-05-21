## 1. Database Migration

- [x] 1.1 Write migration SQL: add `auth_subject TEXT NOT NULL UNIQUE` to `users`
- [x] 1.2 Write migration SQL: add `profile_complete BOOLEAN NOT NULL DEFAULT false` to `users`
- [x] 1.3 Write migration SQL: alter `date_of_birth`, `weight`, `gender` to nullable

## 2. Backend — DB Layer

- [x] 2.1 Update `CreateUserParams` in `db/users.go` to include `AuthSubject string` and make `DateOfBirth`, `Weight`, `Gender` pointers (nullable)
- [x] 2.2 Update `CreateUser` INSERT query to include `auth_subject` and `profile_complete = true`; handle upsert on duplicate `auth_subject` (update biometrics + set `profile_complete = true`)
- [x] 2.3 Add `GetUserBySubject(ctx, sub string) (User, error)` to `db/users.go` — returns user row or `ErrNotFound`
- [x] 2.4 Update DB layer unit tests to cover new params and `GetUserBySubject`

## 3. Backend — Handlers

- [x] 3.1 Update `createUserRequest` in `handlers/users.go`: add `AuthSubject` field, remove `Email` required-from-client constraint (still required in body), make `DateOfBirth`/`Weight`/`Gender` required (they come from the form)
- [x] 3.2 Update `CreateUser` handler to pass `auth_subject` to DB layer and return `{ "id": <int> }` on 201; handle upsert 200 response on duplicate sub
- [x] 3.3 Create `handlers/users_me.go` with `GetUserMe` handler: reads `sub` query param, calls `GetUserBySubject`, returns 200+profile or 404
- [x] 3.4 Register `GET /api/v1/users/me` route in `cmd/api/main.go`
- [x] 3.5 Add handler unit tests for `GetUserMe` (found, not found, missing sub param)

## 4. Frontend — ProfileCompletionForm

- [x] 4.1 Create `frontend/src/ProfileCompletionForm.jsx` with fields: `date_of_birth`, `weight`, `gender` only
- [x] 4.2 On submit, build POST body from `user.profile` claims (`sub` → `auth_subject`, `given_name` → `first_name`, `family_name` → `last_name`, `email`) + form fields
- [x] 4.3 On 201 response, call a callback to signal completion (parent re-checks profile status)
- [x] 4.4 Handle 400 error display; disable submit while in flight
- [x] 4.5 Delete `frontend/src/SignupForm.jsx`

## 5. Frontend — Profile Gate in App.jsx

- [x] 5.1 Add `profileStatus` state (`'loading' | 'complete' | 'incomplete'`) to `App.jsx`
- [x] 5.2 After `isAuthenticated` resolves (and not `isLoading`/`restoringSession`), fetch `GET /api/v1/users/me?sub=<user.profile.sub>`
- [x] 5.3 Map response to `profileStatus`: 404 or `profile_complete: false` → `'incomplete'`; `profile_complete: true` → `'complete'`
- [x] 5.4 Render `<ProfileCompletionForm>` when `profileStatus === 'incomplete'`; render app when `'complete'`; render loading state when `'loading'`
- [x] 5.5 On profile completion callback, re-fetch profile status to transition gate to `'complete'`
