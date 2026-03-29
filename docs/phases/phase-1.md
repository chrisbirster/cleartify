# Phase 1 - Context Inbox

This document defines the remaining implementation scope for Phase 1 only.

Phase 0 stays authoritative for foundation work that is already complete:
- repo layout
- embedded frontend serving
- SQLite bootstrap
- seeded demo data
- shared app shell
- WorkOS auth boundary

This phase must preserve the existing WorkOS auth boundary, continue using seeded SQLite/demo data only, and keep the UI premium, dark, and minimal.

## Current repo state summary

- `docs/phases/` did not exist before this file. There is no existing `docs/phases/phase-0.md` in the repo yet.
- The Go app already serves the embedded authenticated app at `/app` and the current read-only JSON endpoints at:
  - `GET /api/auth/session`
  - `GET /api/signals`
  - `GET /api/clusters`
  - `GET /api/work-objects`
- The existing WorkOS auth boundary is already in place through `internal/auth/*` and `GET /api/auth/session`. Phase 1 should not replace or expand that boundary beyond current usage.
- Seeded SQLite/demo data already includes:
  - 6 `Signal` records
  - 3 `Cluster` records
  - 2 `Work Object` records
  - seeded `Decision` and `Artifact` records for the existing `Work Object` examples
- The frontend already loads all app data at boot in `web/src/App.tsx` via `web/src/lib/api.ts`.
- `/app/inbox` and `/app/inbox/:clusterId` exist today, but they are read-only views over seeded data.
- The missing Phase 1 slice is the write path from `Cluster` to `Work Object`.

## What is already done that overlaps Phase 1

### Backend

- `signals`, `clusters`, and `cluster_signals` tables already exist in `internal/store/migrate.go`.
- `Signal` and `Cluster` seeded demo data already exists in `internal/store/seed.go`.
- `internal/store/signals.go` already returns `Signal` rows.
- `internal/store/clusters.go` already returns `Cluster` rows with nested `Signal` data.
- `internal/server/handlers_signals.go` already serves `GET /api/signals`.
- `internal/server/handlers_clusters.go` already serves `GET /api/clusters`.
- Route registration already includes `GET /api/signals` and `GET /api/clusters`.
- Backend tests already cover the current read-only `Signal` and `Cluster` paths.

### Frontend

- `web/src/router.tsx` already includes:
  - `/app/inbox`
  - `/app/inbox/:clusterId`
- `web/src/pages/InboxPage.tsx` already renders seeded `Cluster` cards.
- `web/src/pages/ClusterDetailPage.tsx` already renders the selected `Cluster` and its nested `Signal` entries.
- `web/src/pages/InboxPage.test.tsx` and `web/src/pages/ClusterDetailPage.test.tsx` already cover the current read-only rendering path.

## Exact remaining work for `/app/inbox`

The current inbox works as a basic list, but it is not yet a finished Phase 1 Context Inbox.

Remaining work:

1. Keep `/app/inbox` as a seeded `Cluster` list, but upgrade it from a plain link list to a stronger grouped Context Inbox view.
2. Show clearer `Cluster` status in each card:
   - severity
   - status
   - `Signal` count
   - a more visible “open cluster” affordance
3. Show better grouping context directly on the card:
   - preview the top 1 to 2 `Signal` titles or source types
   - keep the design dark, minimal, and premium rather than adding dense table UI
4. Add a proper empty state when there are zero `Cluster` records.
5. Keep the page seeded/demo-only. No real ingestion UI is required in Phase 1.
6. Do not add new global client state just for the inbox. Stay consistent with the current app bootstrap pattern.

## Exact remaining work for cluster detail

The current cluster detail route exists, but it is still a simple read-only list of nested `Signal` cards.

Remaining work:

1. Keep `/app/inbox/:clusterId` as the single detail route for a `Cluster`.
2. Upgrade the header to show the key `Cluster` metadata:
   - title
   - summary
   - severity
   - status
   - `Signal` count
3. Add a clear way back to `/app/inbox`.
4. Improve each `Signal` card so the grouped context is more legible:
   - source type
   - title
   - body
   - optional link when `url` exists
   - keep the layout visually calm and premium
5. Add the primary Phase 1 action on this page:
   - `Create Work Object`
6. Add action states on the page:
   - idle
   - submitting
   - failure
7. After creation succeeds, move the user directly to the new `Work Object` detail page.

## Exact remaining work for Cluster -> Work Object creation

This is the main unfinished deliverable for Phase 1.

### Backend

1. Add a Phase 1 persistence link from `Cluster` to `Work Object`.
2. The simplest schema change is:
   - add nullable `cluster_id` to `work_objects`
3. Use `cluster_id` rather than a new join table for Phase 1 because:
   - one `Cluster` can still spawn zero or more `Work Object` rows
   - a `Work Object` can still exist without a `Cluster`
   - it keeps the MVP schema light
4. Backfill seeded demo `Work Object` rows with `cluster_id` where appropriate:
   - `wo_mobile_perf` -> `clu_mobile_perf`
   - `wo_onboarding_refresh` -> `clu_onboarding_refresh`
5. Add `POST /api/clusters/:id/create-work-object`.
6. The endpoint should:
   - verify that the `Cluster` exists
   - create a new `Work Object` derived from the `Cluster`
   - persist `cluster_id`
   - set a stable seeded/demo-friendly default status such as `planning`
   - map `Cluster` severity into a default `Work Object` priority
   - use the demo user as the owner for now
   - return the created `Work Object` payload
7. Preserve the current WorkOS auth boundary:
   - do not add a new auth stack
   - do not expand Phase 1 into real AuthKit flow work

### Frontend

1. Add a client function in `web/src/lib/api.ts` for `POST /api/clusters/:id/create-work-object`.
2. Add the `Create Work Object` button to `web/src/pages/ClusterDetailPage.tsx`.
3. On submit:
   - disable the button
   - show a pending label
   - call the new endpoint
4. After success, do a hard navigation to `/app/work/:workObjectId`.
5. Use a hard navigation rather than introducing a new app-wide data cache in Phase 1.
6. On error, show a compact inline error state on the detail page and keep the rest of the page usable.

### Scope guardrails

- Do not implement generic `POST /api/work-objects` in Phase 1 unless it becomes strictly necessary to support `POST /api/clusters/:id/create-work-object`.
- Do not add plan generation, artifact generation, or review logic here.
- Do not expand the create flow into full editable forms yet.
- Do not replace seeded demo behavior with real ingestion.

## Backend endpoints involved

| Endpoint | Status | Purpose in Phase 1 |
| --- | --- | --- |
| `GET /api/auth/session` | existing | Keep the current WorkOS auth boundary untouched while loading the app shell. |
| `GET /api/signals` | existing | Provides the seeded `Signal` pool already grouped into `Cluster` records. |
| `GET /api/clusters` | existing | Powers `/app/inbox` and `/app/inbox/:clusterId`. |
| `GET /api/work-objects` | existing | Allows the app to load created and seeded `Work Object` records after reload. |
| `POST /api/clusters/:id/create-work-object` | missing | Creates a `Work Object` from a selected `Cluster`. |

## Frontend routes/components involved

### Routes

- `/app/inbox`
- `/app/inbox/:clusterId`
- `/app/work/:workObjectId`

### Frontend files already involved

- `web/src/App.tsx`
- `web/src/router.tsx`
- `web/src/lib/api.ts`
- `web/src/lib/types.ts`
- `web/src/layouts/AppShell.tsx`
- `web/src/components/PageHeader.tsx`
- `web/src/pages/InboxPage.tsx`
- `web/src/pages/ClusterDetailPage.tsx`
- `web/src/pages/WorkObjectDetailPage.tsx`

### Frontend files likely to change for Phase 1

- `web/src/lib/api.ts`
- `web/src/lib/types.ts`
- `web/src/pages/InboxPage.tsx`
- `web/src/pages/ClusterDetailPage.tsx`
- `web/src/pages/WorkObjectDetailPage.tsx` only if the new `clusterId` relationship needs to be displayed
- relevant test files beside each touched component

## Database tables/entities involved

### Primary Phase 1 entities

- `Signal`
- `Cluster`
- `Work Object`

### Tables already in use

- `signals`
- `clusters`
- `cluster_signals`
- `work_objects`

### Tables indirectly involved

- `users`
  - current demo owner for created `Work Object` records
- `sessions`
  - current demo session/auth boundary

### Tables not in primary Phase 1 scope

- `decisions`
- `artifacts`

These remain part of seeded Phase 0 and Phase 2 overlap, but they are not the main write target for Phase 1.

### Required Phase 1 schema update

- add nullable `cluster_id` to `work_objects`

## Tests to add

### Backend tests

1. `internal/store`
   - create `Work Object` from a valid `Cluster`
   - fail when the `Cluster` does not exist
   - persist `cluster_id`
   - derive default status and priority correctly
2. `internal/server`
   - route registration for `POST /api/clusters/:id/create-work-object`
   - method rejection for non-POST requests
   - success response returns the created `Work Object`
   - error response when the `Cluster` is missing
3. migration/seed coverage
   - Phase 1 migration adds `cluster_id`
   - seeded `Work Object` rows are backfilled with `cluster_id`

### Frontend tests

1. `InboxPage`
   - renders richer `Cluster` summary metadata
   - renders empty state when there are no `Cluster` records
2. `ClusterDetailPage`
   - renders severity, status, and `Signal` count
   - renders `Signal` links when present
   - renders `Create Work Object` action
   - shows pending state while creating
   - shows inline error state on failure
   - triggers redirect after success
3. `api.ts`
   - covers the new cluster-to-`Work Object` POST helper
4. optional app-level test
   - confirm a hard navigation or equivalent reload path occurs after creation so fresh `Work Object` data is loaded from the backend

## Validation commands

Run these after Phase 1 implementation work:

- `go test ./internal/store ./internal/server ./internal/app ./internal/httpx ./internal/auth ./internal/config`
- `go test ./...`
- `go build ./...`
- `cd /Users/gm/dev/personal/cleartify/web && bun run test`
- `cd /Users/gm/dev/personal/cleartify/web && bun run build`
- `cd /Users/gm/dev/personal/cleartify && task build`
- `cd /Users/gm/dev/personal/cleartify && ./scripts/check_phase0_coverage.sh`

Manual seeded/demo checks:

- boot the app with the seeded SQLite database
- open `/app/inbox`
- open a seeded `Cluster`
- create a new `Work Object`
- confirm the app lands on `/app/work/:workObjectId`
- confirm the new `Work Object` persists after reload

SQLite spot checks:

- `sqlite3 .tasktmp/dev.sqlite.db '.schema work_objects'`
- `sqlite3 .tasktmp/dev.sqlite.db 'select id, cluster_id, title, status, priority from work_objects order by created_at;'`

## Definition of done for Phase 1

Phase 1 is done when all of the following are true:

1. `/app/inbox` presents seeded `Cluster` records as a polished Context Inbox, not just a plain list.
2. `/app/inbox/:clusterId` shows a polished `Cluster` detail view with the grouped `Signal` context clearly visible.
3. A user can create a `Work Object` from a `Cluster` using seeded/demo data only.
4. The create action persists a `cluster_id` relationship in SQLite.
5. After creation, the user lands on the new `Work Object` detail page with fresh data loaded from the backend.
6. The current WorkOS auth boundary remains intact and unchanged in principle.
7. The UI remains premium, dark, and minimal.
8. Backend and frontend tests for the new Phase 1 behavior pass.
9. Phase 1 changes do not rewrite or reopen Phase 0 foundation work.
