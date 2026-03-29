# Phase 0 TDD-First Implementation Plan

## 1. Assumptions
- Phase 0 will be implemented test-first: every new package, handler, store method, component, route, and shell behavior gets a failing test before production code is written.
- The 90% line-coverage requirement applies to all Phase 0 touched code only, not the entire existing repo.
- Frontend testing is primarily unit/component/router coverage plus Go HTTP integration tests; browser E2E is limited to a few shell smoke paths.
- Canonical repo layout for Phase 0 is:
  - `marketing/` separate site
  - `web/` embedded app shell
  - `cmd/cleartify/` Go entrypoint
  - `internal/` for app, config, auth, store, server, and domain code
- `README.md` is now the repo-level source of truth; stale `app/web` and `app/server` references should be corrected during implementation.
- WorkOS in Phase 0 is boundary-only:
  - config loading
  - client wrapper
  - session contract
  - demo auth implementation for local/dev
  - no full AuthKit redirect/callback/tenant auth flow yet
- SQLite is the only persistence target in Phase 0, with deterministic demo seed data.
- Local dev remains localhost-first:
  - app shell `:3000`
  - marketing `:4173`
  - Go server `:8000`

## 2. Risks
- Existing repo state is not green, so TDD must begin by restoring a passing baseline for the packages Phase 0 will touch.
- The current repo has naming drift (`cleartify`, `microdote`, `VUTADEX_*`, mixed env conventions); if not normalized first, tests will lock in the wrong structure.
- A 90% line-coverage target forces clean seams around embed, config, SQLite, and WorkOS wrappers; if those seams are skipped, coverage will become brittle or fake.
- SPA embed behavior is easy to miss in tests; explicit handler tests are required for `/app`, `/app/*`, `/api/*`, and embedded asset paths.
- Seed data can quietly turn into fixture sprawl; it must stay small, deterministic, and directly tied to shell screens.
- CGO-backed SQLite and Docker build flow need explicit validation because build success is part of the Phase 0 contract.

## 3. Proposed File Structure
```text
/README.md
/.env.example
/Taskfile.yml
/Dockerfile
/cmd/
  cleartify/
    main.go
    main_test.go

/internal/
  app/
    app.go
    app_test.go
  config/
    config.go
    config_test.go
  domain/
    models.go
  auth/
    crypto.go
    crypto_test.go
    state.go
    state_test.go
    service.go
    demo.go
    demo_test.go
    workos.go
    workos_test.go
    workosclient/
      client.go
      client_test.go
  server/
    server.go
    server_test.go
    routes.go
    routes_test.go
    embed.go
    embed_test.go
    middleware.go
    middleware_test.go
    handlers_health.go
    handlers_health_test.go
    handlers_auth.go
    handlers_auth_test.go
    handlers_signals.go
    handlers_signals_test.go
    handlers_clusters.go
    handlers_clusters_test.go
    handlers_work_objects.go
    handlers_work_objects_test.go
  store/
    sqlite.go
    sqlite_test.go
    migrate.go
    migrate_test.go
    seed.go
    seed_test.go
    signals.go
    signals_test.go
    clusters.go
    clusters_test.go
    work_objects.go
    work_objects_test.go
  httpx/
    json.go
    json_test.go
    errors.go
    errors_test.go

/web/
  package.json
  vite.config.ts
  vitest.config.ts
  index.html
  src/
    main.tsx
    main.test.tsx
    App.tsx
    App.test.tsx
    router.tsx
    router.test.tsx
    index.css
    lib/
      api.ts
      api.test.ts
      types.ts
    layouts/
      AppShell.tsx
      AppShell.test.tsx
    components/
      SideNav.tsx
      SideNav.test.tsx
      TopBar.tsx
      TopBar.test.tsx
      PageHeader.tsx
      PageHeader.test.tsx
      EmptyState.tsx
      EmptyState.test.tsx
    pages/
      TodayPage.tsx
      TodayPage.test.tsx
      InboxPage.tsx
      InboxPage.test.tsx
      ClusterDetailPage.tsx
      ClusterDetailPage.test.tsx
      WorkObjectsPage.tsx
      WorkObjectsPage.test.tsx
      WorkObjectDetailPage.tsx
      WorkObjectDetailPage.test.tsx
      PrototypesPage.tsx
      PrototypesPage.test.tsx
      PrototypeNewPage.tsx
      PrototypeNewPage.test.tsx
      PlanPage.tsx
      PlanPage.test.tsx
      AgentWorkspacePage.tsx
      AgentWorkspacePage.test.tsx
      ReviewPage.tsx
      ReviewPage.test.tsx
      ReleasesPage.tsx
      ReleasesPage.test.tsx
      ReleaseDetailPage.tsx
      ReleaseDetailPage.test.tsx
  test/
    smoke/
      app-shell.spec.ts

/scripts/
  check_phase0_coverage.sh
```

## 4. Step-by-Step Implementation Order
1. Re-establish a green test/build baseline before feature work.
- Normalize imports to `cleartify/...`.
- Standardize env/config naming.
- Add `go.sum` completeness and make `go build` deterministic.
- First tests:
  - config loading
  - main bootstrap wiring
  - health route boots

2. Create the coverage harness before writing the app.
- Add Go coverage script scoped to touched Phase 0 packages.
- Add `web/` Vitest + Testing Library + coverage config.
- Add one browser smoke harness for `/app`.
- Coverage gate target: at least 90% line coverage for every Phase 0 touched Go package and web source file.

3. Scaffold Go entrypoint and server via TDD.
- Write failing tests for:
  - app construction
  - route registration
  - JSON helpers
  - `/api/health`
- Then implement `cmd/cleartify/main.go`, config loading, server construction, and route mux.

4. Scaffold SQLite bootstrap via TDD.
- Write failing tests for:
  - SQLite open options
  - migration runner idempotency
  - required table creation
  - empty DB bootstrap
- Then implement DB open, PRAGMA setup, and migrations.

5. Add deterministic seed strategy via TDD.
- Write failing tests for:
  - seed-on-empty behavior
  - no reseed on non-empty DB
  - exact seeded record counts
  - key relationship integrity across signals, clusters, work objects, decisions, and artifacts
- Then implement seed loading.

6. Define the WorkOS boundary via TDD.
- Write failing tests for:
  - auth config validation
  - demo auth session behavior
  - session response shape
  - WorkOS client wrapper behavior with fakes
- Then implement:
  - `AuthService`
  - `DemoAuthService`
  - `WorkOSAuthService` skeleton
  - `GET /api/auth/session`
- Do not implement redirect/callback flows yet.

7. Build read-only Phase 0 API endpoints via TDD.
- Write failing handler + store tests for:
  - `GET /api/signals`
  - `GET /api/clusters`
  - `GET /api/work-objects`
  - expected empty/error behavior
  - seeded data return shape
- Then implement store methods and handlers.

8. Scaffold `web/` Vite app via TDD.
- Write failing tests for:
  - router base path `/app`
  - shell render
  - navigation items
  - auth session bootstrap behavior
- Then implement Vite config, Tailwind, app bootstrap, API client, router, and shared shell.

9. Add the real Phase 0 shell pages via TDD.
- Write page tests first for:
  - `/app` today overview from seeded data
  - `/app/inbox` list rendering
  - `/app/inbox/:clusterId` detail rendering
  - `/app/work` list rendering
  - `/app/work/:workObjectId` detail rendering
- Then implement those pages.
- For later-phase routes, write scaffold tests that assert intentional placeholder content and stable route registration.

10. Add embedded frontend serving via TDD.
- Write failing server tests for:
  - `/app` returns embedded index
  - `/app/*` SPA fallback works
  - `/app/assets/*` returns built assets
  - `/api/*` never falls through to SPA
  - unknown non-app routes behave intentionally
- Then implement embed logic and static serving.

11. Update tasks and Docker via TDD-style validation.
- Write or add smoke tests/scripts that assert:
  - `task dev:sqlite:app` boots all required services
  - `task build:server` depends on a successful `web` build
  - Docker build succeeds with embedded app assets
- Then update `Taskfile.yml` and `Dockerfile`.

12. Close Phase 0 only when all touched code is fully covered.
- No Phase 0 file merges without its matching tests.
- No package is considered done below 90% line coverage for touched files.
- Final acceptance is build green, tests green, coverage gate green, and shell booting against seeded SQLite data.

## 5. Validation Commands
- `go test ./... -coverprofile=/tmp/cleartify.cover`
- `go tool cover -func=/tmp/cleartify.cover`
- `cd web && bun run test --coverage`
- `cd web && bun run build`
- `go build ./...`
- `task build:server`
- `task build`
- `task dev:sqlite:app`
- `curl -i http://localhost:8000/api/health`
- `curl -i http://localhost:8000/api/auth/session`
- `curl -i http://localhost:8000/api/clusters`
- `curl -I http://localhost:8000/app`
- `sqlite3 .tasktmp/dev.sqlite.db '.tables'`
- `scripts/check_phase0_coverage.sh`

## 6. Open Questions
- None block implementation.
- Default decisions now locked:
  - 90% line coverage applies to Phase 0 touched code only
  - frontend testing is unit/component/router heavy plus Go HTTP integration tests
  - browser E2E is kept to a thin shell smoke layer
  - only `web/` is embedded into the Go binary; `marketing/` stays separate
