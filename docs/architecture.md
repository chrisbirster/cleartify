# Cleartify Architecture

## Stack
- Frontend: React + Vite + TypeScript + Tailwind CSS, with Effect at the async/data boundary in `/web`
- Backend: Go 1.25 using net/http
- Static serving: Vite build embedded into Go binary with embed
- Database: SQLite for local development
- API style: JSON over REST-like HTTP routes
- Auth: WorkOS AuthKit with backend-enforced protected routes and MVP-scoped session handling and tenant-based urls like foo.cleartify.com bar.cleartify.com
- Deployment target: single Go binary app

## App shape
Monorepo structure:
- /web -> Vite frontend
- /cmd/cleartify -> Go entrypoint
- /internal -> Go backend packages
- /data/query -> embedded production SQL queries
- /data/migration -> embedded production SQL migrations
- /docs -> product and engineering documentation

## Frontend responsibilities
- routing
- dashboard pages
- inbox views
- work object views
- prototype upload/review screens
- plan/review/release screens
- keep async/data effects at the boundary rather than inside presentational JSX components

## Backend responsibilities
- serve API
- serve embedded frontend assets
- handle persistence
- manage domain objects
- expose mockable agent endpoints for MVP
- load production queries and migrations from embedded SQL files in `data/query` and `data/migration`

## MVP constraint
Do not implement real LLM orchestration first.
Use mocked or stubbed agent outputs behind stable interfaces.
Design the agent layer so real providers can be added later.

## Quality goals
- clear folder structure
- simple domain models
- minimal dependencies
- copy-pasteable local setup
- strong typing
- tests for core backend logic
