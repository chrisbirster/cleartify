# Cleartify Phase Tracker

Last updated: 2026-03-29

Current target: Phase 2 - Work Objects

## Status Summary

| Phase | Name | Status | Notes |
| --- | --- | --- | --- |
| 0 | Foundation | Done | Core repo structure, Go app scaffold, embedded `web/`, SQLite bootstrap, seed data, app shell, and test/build pipeline are in place. |
| 1 | Context Inbox | Done | Polished `/app/inbox` and `/app/inbox/:clusterId` views are in place, `work_objects.cluster_id` is persisted, and Cluster -> Work Object creation now lands on the new Work Object detail route. |
| 2 | Work Objects | In progress | Work Object schema, read APIs, seeded examples, detail page, decisions, and artifacts panels exist. Status transitions and fuller evidence flows are still missing. |
| 3 | Prototype-to-Plan | Scaffolded | Stable routes and placeholder pages exist, but no real attach form or plan generation flow yet. |
| 4 | Agent Workspace | Scaffolded | Stable route and placeholder page exist, but mock agent actions and artifact generation are not implemented yet. |
| 5 | Review and Release | Scaffolded | Stable review and release routes exist, but release/review workflows are not implemented yet. |
| 6 | Polish | Not started | No dedicated polish pass yet beyond Phase 0 baseline styling and shell quality. |

## Phase Breakdown

### Phase 0 - Foundation

Status: Done

Completed:
- top-level monorepo structure aligned around `marketing/`, `web/`, `internal/`, and `cmd/cleartify`
- Go app entrypoint and HTTP server scaffold
- embedded frontend serving from the Go binary
- SQLite bootstrap and migrations
- deterministic seed data
- shared app shell and route map
- local build/test flow
- at least 90% coverage for Phase 0 touched Go and `web` source

### Phase 1 - Context Inbox

Status: Done

Completed:
- `signals` and `clusters` persistence
- seeded Signal and Cluster demo data
- `GET /api/signals`
- `GET /api/clusters`
- `/app/inbox`
- `/app/inbox/:clusterId`
- polished Context Inbox cards with Signal previews, metadata, and empty state
- polished Cluster detail with grouped Signal context and action states
- `work_objects.cluster_id` schema link
- `POST /api/clusters/:id/create-work-object`
- redirect from Cluster creation flow to `/app/work/:workObjectId`

### Phase 2 - Work Objects

Status: In progress

Completed:
- `work_objects`, `decisions`, and `artifacts` persistence
- seeded Work Object examples
- `GET /api/work-objects`
- `/app/work`
- `/app/work/:workObjectId`
- decisions and artifacts panels

Remaining:
- evidence panel
- status transition actions
- write flows for creating and updating Work Objects

### Phase 3 - Prototype-to-Plan

Status: Scaffolded

Completed:
- `/app/prototypes`
- `/app/prototypes/new`
- `/app/plans/:workObjectId`

Remaining:
- prototype attach form
- prototype persistence
- stub plan generation endpoint
- editable generated plan
- edge cases and open questions section

### Phase 4 - Agent Workspace

Status: Scaffolded

Completed:
- `/app/agents/:workObjectId`

Remaining:
- mock agent actions
- task generation
- test plan generation
- release note drafting
- artifact creation
- activity timeline

### Phase 5 - Review and Release

Status: Scaffolded

Completed:
- `/app/review/:workObjectId`
- `/app/releases`
- `/app/releases/:releaseId`

Remaining:
- review screen workflow
- release summary view
- linked decisions and artifacts flow
- outcome notes

### Phase 6 - Polish

Status: Not started

Planned:
- better loading and empty states
- richer seeded demo data
- keyboard shortcuts if simple
- visual polish
- docs cleanup

## Next Recommended Focus

1. Expand Phase 2 with Work Object write flows and status transitions.
2. Surface the Cluster relationship more explicitly on the Work Object detail view and lists.
3. Keep Phases 3 through 5 as placeholders until the Work Object editing loop is solid.
