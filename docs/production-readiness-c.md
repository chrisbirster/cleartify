# Production Readiness Tranche C

## Scope

Tranche C adds the minimum operational surfaces needed after auth and onboarding are real.

It includes:
- `/me` current-user page
- `/team` current-team/workspace page
- `Cluster` resolve or archive flow
- `Signal` CRUD
- `Decision` CRUD
- `Artifact` CRUD
- external link validation and source typing as lightweight integrations-first support

## Non-goals

- no internal Confluence-like docs product
- no new `Article` entity
- no enterprise admin or RBAC console
- no background sync jobs or ingestion pipelines yet

## Routes / files likely to change

Frontend:
- shell/top bar/nav
- `/me`
- `/team`
- inbox routes and actions
- new `Signal`, `Decision`, and `Artifact` views

Backend:
- user/team read handlers
- cluster action handlers
- CRUD handlers and store methods for `Signal`, `Decision`, and `Artifact`
- lightweight URL/source validation helpers

Likely files:
- `web/src/components/TopBar.tsx`
- `web/src/router.tsx`
- `web/src/pages/InboxPage.tsx`
- `web/src/pages/ClusterDetailPage.tsx`
- new pages for `Signal`, `Decision`, and `Artifact`
- `web/src/lib/api.ts`
- `web/src/lib/types.ts`
- `internal/server/routes.go`
- `internal/server/handlers_*`
- `internal/store/signals.go`
- `internal/store/work_objects.go`
- new store files as needed for user/team read paths

## Backend work

- Add current-user and current-team/workspace read endpoints.
- Add a narrow persisted `Cluster` resolve or archive action using existing cluster state.
- Add CRUD endpoints for:
  - `Signal`
  - `Decision`
  - `Artifact`
- Add lightweight external source typing and URL validation for linked docs and code references.

## Frontend work

- Make the session user clickable to `/me`.
- Surface the current team/workspace and link to `/team`.
- Add `Cluster` resolve/archive action in Context Inbox.
- Add dedicated list/detail/create/edit surfaces for `Signal`, `Decision`, and `Artifact`.
- Keep integrations lightweight and link-first.

## Tests

- `/me` and `/team` rendering
- cluster resolve/archive behavior and inbox filtering
- CRUD success, validation failure, and not-found cases for `Signal`, `Decision`, and `Artifact`
- source typing and URL validation behavior

## Validation commands

- `go test ./...`
- `go build ./...`
- `cd /Users/gm/dev/personal/cleartify/web && bun run test`
- `cd /Users/gm/dev/personal/cleartify/web && bun run build`

## Hard definition of done

- current user and team/workspace are visible in the app
- resolved or archived `Cluster`s leave the default inbox flow
- `Signal`, `Decision`, and `Artifact` each have a real CRUD surface
- external links are typed and validated consistently
