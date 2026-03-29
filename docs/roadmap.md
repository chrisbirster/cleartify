# Cleartify MVP Roadmap

## Phase 0 - Foundation
Goal:
Create repo structure, app shell, build pipeline, embedded frontend serving, database bootstrap, and seeded demo data.

Deliverables:
- monorepo layout
- Vite app booting
- Go server booting
- embedded dist served from Go
- SQLite schema init
- seed script
- shared app layout
- dark theme baseline

## Phase 1 - Context Inbox
Goal:
Build inbox UI and backend APIs for Signals and Clusters.

Deliverables:
- signals table + clusters table
- seeded signals
- context inbox page
- cluster grouping UI
- cluster detail panel
- create Work Object action

## Phase 2 - Work Objects
Goal:
Create the main execution object and its detail page.

Deliverables:
- work_objects table
- work object detail page
- evidence panel
- decisions panel
- artifacts panel
- status transitions
- seeded examples

## Phase 3 - Prototype-to-Plan
Goal:
Allow a prototype or design link to generate a plan.

Deliverables:
- prototype attach form
- stub plan generation endpoint
- generated plan page
- editable plan output
- edge cases / questions section

## Phase 4 - Agent Workspace
Goal:
Show visible agent actions without real provider complexity.

Deliverables:
- agent workspace UI
- mock actions: summarize, generate tasks, draft test plan, draft release notes
- artifact creation from agent actions
- activity timeline

## Phase 5 - Review and Release
Goal:
Close the loop from plan to shipping.

Deliverables:
- review screen
- release summary view
- linked decisions + artifacts
- outcome notes

## Phase 6 - Polish
Goal:
Make the MVP demo-ready.

Deliverables:
- better loading/empty states
- richer seeded data
- keyboard shortcuts if easy
- visual polish
- docs cleanup