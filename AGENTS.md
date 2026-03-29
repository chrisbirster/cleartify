# AGENTS.md

## Mission
Build Cleartify, an AI-native product execution system.
Tickets are not the source of truth; context is.
The MVP must feel fast, minimal, and product-focused.

## Working style
- Make small, high-confidence changes.
- Prefer simple, maintainable implementations over clever ones.
- Preserve existing behavior unless the task explicitly changes it.
- When adding a feature, also update any nearby docs.
- Do not introduce major new dependencies unless necessary.
- Keep UI polished but lightweight.
- Use clear naming and simple abstractions.
- Stub complex AI behavior behind interfaces rather than overbuilding real orchestration.

## Architecture rules
- Frontend is React + Vite + TypeScript + Tailwind.
- Use Effect at the `/web` async/data boundary.
- Backend is Go 1.25 with net/http.
- Frontend build output is embedded into the Go server with embed.
- SQLite is the local database.
- Turso is the production database
- Keep the backend modular but not overly abstract.
- Place production SQL queries in `data/query` and production migrations in `data/migration`, then embed them in the app.
- Auth stack is WorkOS AuthKit; do not replace it with a different auth provider or local auth system unless explicitly asked.

## UI rules
- Premium dark theme
- Clean layout
- Strong hierarchy
- Minimal clutter
- Product dashboard aesthetic
- Reusable components
- Accessible contrast and states

## Domain language
Use these exact terms:
- Signal
- Cluster
- Work Object
- Decision
- Artifact
- Prototype-to-Plan
- Context Inbox
- Agent Workspace
- Product Memory

## Delivery rules
For each task:
1. explain the change briefly
2. implement the code
3. run relevant checks
4. summarize files changed
5. note follow-up tasks if any

## Validation
After modifying frontend files:
- run frontend build
- run relevant tests if present

After modifying backend files:
- run backend tests
- ensure app still boots

## MVP priorities
1. working app shell
2. seeded demo data
3. inbox screen
4. work object detail
5. prototype-to-plan flow
6. stubbed agent actions
7. review/release loop

## Avoid
- full auth system
- real billing
- real LLM provider integration in MVP
- enterprise-only features
- large state management frameworks unless needed

## Definition of done
A feature is done when:
- UI is functional
- data flows end to end
- demo data makes the feature easy to review
- docs are updated if structure changed
- build/test passes
