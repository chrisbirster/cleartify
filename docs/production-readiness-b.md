# Production Readiness Tranche B

## Scope

Tranche B covers real auth and first-run onboarding only.

It includes:
- real hosted WorkOS AuthKit login flow
- callback handling
- logout handling
- real session payload from the backend
- `needsOnboarding` support
- onboarding walkthrough
- onboarding completion persistence

## Non-goals

- no `/me` or `/team` yet
- no CRUD surfaces
- no custom OTP or local auth system
- no background jobs
- no provider orchestration or broader AI work

## Routes / files likely to change

Backend:
- auth service and WorkOS integration files
- config
- auth handlers and route registration
- signed state / cookie helpers

Frontend:
- app bootstrap and session types
- unauthenticated entry surface
- onboarding presentation components or routes

Likely files:
- `internal/app/app.go`
- `internal/auth/service.go`
- `internal/auth/workos.go`
- `internal/auth/workosclient/client.go`
- `internal/auth/state.go`
- `internal/server/routes.go`
- `internal/server/handlers_auth.go`
- `internal/domain/models.go`
- `internal/config/config.go`
- `web/src/App.tsx`
- `web/src/lib/api.ts`
- `web/src/lib/types.ts`
- new auth/onboarding UI files as needed

## Backend work

- Add hosted WorkOS auth entry, callback, and logout endpoints.
- Resolve real session cookies from the callback flow.
- Extend the session payload with:
  - authenticated user
  - organization or current team/workspace context
  - `needsOnboarding`
- Persist onboarding completion, preferably on `users`, with a minimal completion marker.

## Frontend work

- Replace the current unauthenticated placeholder with real hosted-auth entry handling.
- Show onboarding only when `needsOnboarding` is true.
- Provide a compact onboarding walkthrough covering the product flow:
  - Context Inbox
  - Work Objects
  - Prototype-to-Plan
  - Agent Workspace
  - Review and Releases
- Persist completion and return the user into the main app flow.

## Tests

- auth login / callback / logout success and error paths
- session payload includes the expected auth and onboarding fields
- unauthenticated bootstrap behavior
- onboarding shown only when required
- onboarding completion persists and survives reload

## Validation commands

- `go test ./...`
- `go build ./...`
- `cd /Users/gm/dev/personal/cleartify/web && bun run test`
- `cd /Users/gm/dev/personal/cleartify/web && bun run build`

## Hard definition of done

- marketing `Sign in` no longer drops directly into an always-authenticated app shell
- the app uses a real hosted WorkOS AuthKit login flow
- first-run users see onboarding and can complete it
- completion persists and onboarding no longer reappears after reload
