# Cleartify Phase Tracker

Last updated: 2026-04-01

Current target: MVP complete

## Status Summary

| Phase | Name | Status | Notes |
| --- | --- | --- | --- |
| 0 | Foundation | Done | Core repo structure, Go app scaffold, embedded `web/`, SQLite bootstrap, seed data, app shell, and test/build pipeline are in place. |
| 1 | Context Inbox | Done | Polished `/app/inbox` and `/app/inbox/:clusterId` views are in place, `work_objects.cluster_id` is persisted, and Cluster -> Work Object creation now lands on the new Work Object detail route. |
| 2 | Work Objects | Done | The MVP Work Object loop is complete: seeded examples, richer detail page, evidence panel, linked Cluster context, status and priority transitions, core editing, and lightweight Decision and Artifact capture all work end to end. Further Work Object slices should be treated as future enhancements unless explicitly required. |
| 3 | Prototype-to-Plan | Done | Prototype attach, deterministic draft plan generation, and lightweight persisted draft plan editing are now implemented end to end for MVP review. |
| 4 | Agent Workspace | Done | Agent Workspace now supports two deterministic persisted mock actions that turn a Work Object plus its persisted draft plan into `implementation_brief` and `task_breakdown` Artifacts. Treat broader agent scope as future enhancement unless explicitly required. |
| 5 | Review and Release | Done | The MVP review loop is complete: `/app/review/:workObjectId` is a real checkpoint, `release_ready_at` persists, `POST /api/work-objects/:id/release-ready` exists, and `/app/releases` surfaces release-ready Work Objects. Treat release detail and outcome-note scope as optional future enhancement unless explicitly required. |
| 6 | Polish | Done | Shared state polish, cross-page copy/hierarchy polish, and seeded demo-data refinement now make the MVP flow read clearly from Work Object to plan, agent output, review, and releases. |

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

Status: Done

Completed:
- `/app/agents/:workObjectId`
- persisted mock Agent Workspace action for generating an `implementation_brief` Artifact
- persisted mock Agent Workspace action for generating a `task_breakdown` Artifact
- deterministic Artifact derivation from persisted Work Object, linked Cluster context, prototype context, and draft plan sections
- `POST /api/work-objects/:id/implementation-artifact`
- `POST /api/work-objects/:id/task-breakdown-artifact`
- Agent Workspace rendering of current Work Object context and draft plan summary
- immediate and persisted Artifact visibility on the Agent Workspace route
- compact multi-action Agent Workspace UI with pending and inline error states

Exit criteria met:
- Agent Workspace can now read persisted product context and draft plan context to generate more than one concrete deterministic Artifact
- the MVP Agent Workspace feels real without introducing orchestration, background jobs, or real agent logic

Deferred:
- lightweight activity or timeline treatment for generated outputs
- additional mock actions beyond the MVP pair
- any real agent execution, orchestration, or background processing

### Phase 5 - Review and Release

Status: Done

Completed:
- `/app/review/:workObjectId`
- `/app/releases`
- real review checkpoint for a `Work Object`
- persisted `release_ready_at` state on `work_objects`
- `POST /api/work-objects/:id/release-ready`
- review summary with core Work Object details, linked Cluster context, current status and priority, persisted draft plan summary, Decisions, and Artifacts
- compact `Mark ready for release` action with immediate and persisted UI updates
- simple releases list sourced from release-ready `Work Object`s

Exit criteria met:
- a human can review the key execution context before release and mark a Work Object ready
- release-ready state persists, reloads cleanly, and is visible in both review and releases views

Deferred:
- optional release detail surface
- optional outcome-note capture
- any broader approvals, deployment automation, or release workflow engines

### Phase 6 - Polish

Status: Done

MVP polish focus:
- improve loading and empty states on the key execution routes
- tighten copy and visual consistency across Work Object, Plan, Agent Workspace, Review, and Releases
- improve seeded demo data where it makes the product story easier to review
- keep the pass small, product-focused, and demo-oriented

Completed:
- shared loading and empty-state polish across `/app/work/:workObjectId`, `/app/plans/:workObjectId`, `/app/agents/:workObjectId`, `/app/review/:workObjectId`, and `/app/releases`
- reusable state-surface treatment for route-level empty states and section-level missing/loading panels
- tighter empty-state copy and more consistent panel framing across the key execution routes
- cross-page copy and hierarchy polish across `/app/work/:workObjectId`, `/app/plans/:workObjectId`, `/app/agents/:workObjectId`, `/app/review/:workObjectId`, and `/app/releases`
- more consistent page eyebrows, section labels, action copy, badge wording, and card spacing across the key MVP flow
- seeded demo-data polish focused on one clear hero Work Object for `/app/work -> /app/plans -> /app/agents -> /app/review -> /app/releases`
- tighter seeded Cluster, prototype, draft-plan, Decision, Artifact, and release-ready story for the onboarding demo path

Exit criteria met:
- the MVP execution flow reads clearly end to end with seeded demo content
- the app feels intentional and demo-ready without reopening earlier phases or adding new features

Proposed slices:
1. Shared loading and empty-state polish
   - make route loading states feel more intentional and consistent
   - tighten empty-state copy so each page explains what the user should do next
2. Cross-page copy and hierarchy polish
   - normalize headings, section descriptions, button copy, and badge language across `/app/work/:workObjectId`, `/app/plans/:workObjectId`, `/app/agents/:workObjectId`, `/app/review/:workObjectId`, and `/app/releases`
   - align panel spacing, section labels, and summary treatments so the flow feels like one product
3. Seeded demo-data polish
   - tune a few seeded `Decision`, `Artifact`, and plan examples so the core demo path reads more clearly end to end
   - avoid expanding the schema or adding new feature scope

Completed slices:
- shared loading and empty-state polish
- cross-page copy and hierarchy polish
- seeded demo-data polish

## Next Recommended Focus

1. Treat the MVP as complete across Phases 0 through 6.
2. Keep future work narrow and intentional: hosted provider integration, external agent/tool surface, and optional release-detail/outcome-note enhancements.
3. Avoid reopening completed MVP phases unless a concrete demo or product gap appears.
