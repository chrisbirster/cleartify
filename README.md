# Cleartify

Cleartify is an AI-native product execution system.

It turns feedback, notes, prototypes, decisions, and code into working software.

## Core thesis

Tickets are not the source of truth.  
Context is the source of truth.

Cleartify is built for fast product teams that want to move from messy product signals to clear execution with less ceremony. Instead of treating specs and tickets as the center of the workflow, Cleartify keeps the full story attached to the work: signals, clusters, prototypes, plans, decisions, artifacts, and review.

## Product principles

- Context stays attached to the work
- Prototypes are the starting point of execution
- Agents handle procedural work
- Humans provide judgment, taste, and final review
- Keep the product fast, minimal, and low-ceremony
- Build for real execution, not workflow theater

## MVP

The MVP should let a user:

1. connect or ingest product signals
2. review clustered context in a Context Inbox
3. convert a cluster into a Work Object
4. attach a prototype or rough notes
5. generate an implementation plan
6. generate execution artifacts through stubbed agent actions
7. review and move toward release

## Workspace modes

Live Workspace V1 starts with an explicit workspace mode choice for authenticated users:

- `sample` workspace mode keeps the hero demo flow available for evaluation and review
- `live` workspace mode starts clean so real users do not land inside seeded demo data by default

The app shell clearly labels sample workspace mode, and the first-run chooser persists the selection before the main product shell opens.

## Core domain language

Use these exact terms throughout the app and docs:

- Signal
- Cluster
- Work Object
- Decision
- Artifact
- Context Inbox
- Prototype-to-Plan
- Agent Workspace
- Product Memory

## Tech stack

### Frontend
- React
- Vite
- TypeScript
- Tailwind CSS
- Effect at the `/web` data and side-effect boundary

### Backend
- Go 1.25
- net/http
- embedded frontend assets via `embed`

### Data
- SQLite for local development and MVP persistence
- SQL queries live in `data/query`
- SQL migrations live in `data/migration`
- SQL files are embedded into the Go app with `embed`

## Ongoing implementation conventions

- Future `/web` async and data code should use Effect at the boundary.
- Future production SQL should go in `data/query` or `data/migration`.

### Auth
- WorkOS AuthKit

## Why WorkOS AuthKit

Cleartify is using WorkOS AuthKit as the auth layer so the MVP can ship with a clean hosted auth flow now while keeping the door open for enterprise auth later.

This means:

- the Go backend remains the source of protected app access
- auth/session concerns are not rebuilt from scratch
- the app can evolve toward enterprise SSO later without swapping auth stacks midstream

For MVP, build the app with auth boundaries and route protection in mind, but do not let full auth complexity block core product work. Use a simple integration path first and avoid overbuilding org admin or enterprise onboarding flows in the first pass.

## Monorepo layout

```text
/marketing  # React + Vite frontend for the marketing seo friendly app
/web        # React + Vite frontend for the authenticated user
/infra      # SST-backed infra setup needed for powering the app
/data       # embedded SQL queries and migrations
/docs       # product, architecture, roadmap, data model, routes
/internal   # Go based logic
/cmd
  /cleartify
    main.go # The main entrypoint
```

## Planned app routes

Marketing
/ -> marketing site
Product app (app origin)
/ -> Today / overview
/inbox -> Context Inbox
/inbox/:clusterId -> Cluster detail
/work -> Work Objects list
/work/:workObjectId -> Work Object detail
/prototypes -> Prototype list
/prototypes/new -> Attach prototype
/plans/:workObjectId -> Generated plan
/agents/:workObjectId -> Agent Workspace
/review/:workObjectId -> Review flow
/releases -> Release list
/releases/:releaseId -> Release detail
Planned API routes
GET /api/signals
GET /api/clusters
POST /api/workspace-mode
POST /api/clusters/:id/create-work-object
GET /api/work-objects
POST /api/work-objects
GET /api/work-objects/:id
POST /api/work-objects/:id/generate-plan
POST /api/work-objects/:id/generate-artifacts
GET /api/work-objects/:id/artifacts
POST /api/prototypes
Architecture rules
The Vite frontend should build to static assets that are embedded into the Go binary
The Go server should serve both API routes and embedded frontend assets
Keep dependencies light
Prefer simple domain models over premature abstraction
Stub AI behavior behind stable interfaces first
Avoid implementing real LLM orchestration until the product flow is proven
Prefer seeded demo data so features are easy to review during development
Auth architecture notes
WorkOS AuthKit is the chosen auth provider
The backend should own protected route checks
Build with session/cookie-based request validation in mind
Keep auth integration modular so enterprise SSO, organizations, and roles can be layered in later
Do not block MVP feature work on advanced enterprise auth flows
Non-goals for MVP
full Jira replacement
full document suite
full code editor / IDE
real billing
full enterprise admin console
complex workflow/state configuration
real multi-provider LLM orchestration
MVP roadmap
Phase 0 - Foundation
repo structure
frontend scaffold
Go backend scaffold
embedded frontend serving
SQLite bootstrap
seeded demo data
shared app shell
Phase 1 - Context Inbox
Signals and Clusters
seeded signal ingestion examples
inbox page
cluster detail
create Work Object flow
Phase 2 - Work Objects
Work Object schema
work object detail page
evidence, decisions, and artifacts panels
Phase 3 - Prototype-to-Plan
attach prototype
stub plan generation
editable generated plan
edge cases and open questions
Phase 4 - Agent Workspace
mock agent actions
task generation
test plan generation
release note drafting
artifact creation
Phase 5 - Review and Release
review screen
release summary
outcome notes
linked decisions and artifacts
Phase 6 - Polish
loading states
richer seeded demo data
keyboard polish if simple
visual refinement
docs cleanup
Local development

Final commands may evolve, but the intended local workflow is:

Recommended local ports:

- app shell: `http://localhost:3000/`
- marketing: `http://localhost:4173/`
- Go API/server: `http://localhost:8000`

Task-based local dev:

```bash
task bootstrap
task build:env
```

Then fill in `WORKOS_API_KEY` and `WORKOS_CLIENT_ID` inside `.env`.

For hosted WorkOS AuthKit local development, run:

```bash
task dev:e2e
```

Why the ports are split:

- `http://localhost:3000/` is the app UI
- `http://localhost:4173/` is marketing
- `http://localhost:8000` is the Go backend

The browser should use the app origin at `3000`. Vite proxies `/login`, `/logout`, `/auth/*`, and `/api/*` to the Go server on `8000`.

If you want local demo mode without WorkOS, use:

```bash
task dev:sqlite
```

`task dev:e2e` respects your current `.env`.

- use `AUTH_MODE=workos` there for real hosted WorkOS AuthKit testing
- use `AUTH_MODE=demo` there only if you intentionally want the isolated local demo auth path

That starts:

- the Go server on `8000`
- the authenticated app shell on `3000`
- the marketing site on `4173`
- explicit local demo auth mode via `AUTH_MODE=demo`

If you only need the app shell plus backend:

```bash
task dev:sqlite:app
```

Production build
build the frontend
embed the generated frontend assets into the Go server
build the Go binary
run the single binary app
Environment variables

Expected variables will include values like:

PORT
DATABASE_URL or local SQLite path
WORKOS_API_KEY
WORKOS_CLIENT_ID
WORKOS_REDIRECT_URI
WORKOS_COOKIE_PASSWORD or session secret equivalent
AUTH_MODE (`demo` for local/demo usage, `workos` for hosted WorkOS AuthKit)
additional auth/session config as needed

Add a .env.example file early and keep it updated.

Guidance for coding agents

Before implementing features, read:

AGENTS.md
docs/product-prd.md
docs/architecture.md
docs/data-model.md
docs/ui-routes.md
docs/roadmap.md

When making changes:

keep changes small and clean
preserve domain language
avoid inventing a different auth stack
prefer seeded data and stubbed integrations early
update docs when structure changes
run relevant checks before finishing
aim for at least 90% line coverage on touched Go and `web` code
keep future `/web` async and data work on the Effect boundary
keep future production SQL in `data/query` or `data/migration`
Definition of done

A feature is done when:

the UI works end-to-end
the data flow is real or convincingly stubbed
seeded demo data makes the feature easy to inspect
naming matches the product language
docs are updated if needed
builds/tests pass for the touched area
