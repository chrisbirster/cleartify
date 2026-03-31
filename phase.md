# Cleartify Phase Tracker

Last updated: 2026-03-31

Current target: Phase 4 - Agent Workspace

## Status Summary

| Phase | Name | Status | Notes |
| --- | --- | --- | --- |
| 0 | Foundation | Done | Core repo structure, Go app scaffold, embedded `web/`, SQLite bootstrap, seed data, app shell, and test/build pipeline are in place. |
| 1 | Context Inbox | Done | Polished `/app/inbox` and `/app/inbox/:clusterId` views are in place, `work_objects.cluster_id` is persisted, and Cluster -> Work Object creation now lands on the new Work Object detail route. |
| 2 | Work Objects | Done | The MVP Work Object loop is complete: seeded examples, richer detail page, evidence panel, linked Cluster context, status and priority transitions, core editing, and lightweight Decision and Artifact capture all work end to end. Further Work Object slices should be treated as future enhancements unless explicitly required. |
| 3 | Prototype-to-Plan | Done | Prototype attach, deterministic draft plan generation, and lightweight persisted draft plan editing are now implemented end to end for MVP review. |
| 4 | Agent Workspace | In progress | Agent Workspace now supports one deterministic persisted mock action that turns a Work Object plus its persisted draft plan into an `implementation_brief` Artifact. Additional mock actions and timeline behavior are still pending. |
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

Status: Done

Completed:
- `work_objects`, `decisions`, and `artifacts` persistence
- seeded Work Object examples
- `GET /api/work-objects`
- `POST /api/work-objects/:id/status`
- `POST /api/work-objects/:id/priority`
- `POST /api/work-objects/:id/core`
- `POST /api/work-objects/:id/decisions`
- `POST /api/work-objects/:id/artifacts`
- `/app/work`
- `/app/work/:workObjectId`
- linked Cluster context on the Work Object detail page
- evidence panel sourced from linked Cluster Signals
- polished Decisions and Artifacts panels
- compact status and priority controls
- lightweight core-field editing
- lightweight Decision capture
- lightweight Artifact capture
- persisted updates reflected in both detail and list views where relevant

Exit criteria met:
- a Work Object can be reviewed, updated, prioritized, progressed, and enriched with Decisions and Artifacts
- the Work Object detail loop is solid enough to support Prototype-to-Plan next

Deferred:
- standalone Work Object creation outside the Cluster path
- broader generic editing systems
- review/release workflow features

### Phase 3 - Prototype-to-Plan

Status: Done

Completed:
- `/app/prototypes`
- `/app/prototypes/new`
- `/app/plans/:workObjectId`
- lightweight prototype attach flow on `/app/work/:workObjectId`
- persisted prototype link and note on `work_objects`
- `/app/prototypes` rendering attached prototype context from `Work Object`s
- deterministic persisted draft plan stub backed by SQLite
- `GET /api/work-objects/:id/plan`
- `POST /api/work-objects/:id/plan`
- `PUT /api/work-objects/:id/plan`
- `/app/plans/:workObjectId` rendering persisted overview, implementation steps, edge cases, and open questions
- compact Generate Draft Plan and Refresh Draft Plan actions with inline pending and error states
- lightweight manual editing of persisted overview, implementation steps, edge cases, and open questions

Exit criteria met:
- a Work Object can carry prototype context into a deterministic draft plan
- the draft plan can be generated, refined, persisted, and reloaded without agent logic

Deferred:
- richer plan acceptance workflow
- promotion of draft plan output into downstream review/release flows

### Phase 4 - Agent Workspace

Status: In progress

Completed:
- `/app/agents/:workObjectId`
- persisted mock Agent Workspace action for generating an `implementation_brief` Artifact
- deterministic Artifact derivation from persisted Work Object, linked Cluster context, prototype context, and draft plan sections
- `POST /api/work-objects/:id/implementation-artifact`
- Agent Workspace rendering of current Work Object context and draft plan summary
- immediate and persisted Artifact visibility on the Agent Workspace route

Remaining:
- second mock action such as task breakdown or test plan generation
- lightweight activity/timeline treatment for generated outputs
- clearer relationship between multiple mock actions and generated Artifacts

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

1. Add one more narrow deterministic Phase 4 action, preferably a task-breakdown or test-plan Artifact generated from the same persisted draft plan context.
2. Keep Agent Workspace narrow and product-facing instead of introducing orchestration, background jobs, or real agent logic.
3. Treat Phase 4 as incomplete until there is at least one additional mock action or a lightweight activity history for generated outputs.
