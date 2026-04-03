# Cleartify MVP Deployment Checklist

## Goal

Ship the MVP in a way that preserves the demo path, keeps risk low, and avoids reopening product scope.

## 1. Product Readiness

- Confirm `phase.md` marks Phases 0 through 6 as done.
- Confirm the hero demo path is still clear:
  - `/work/wo_onboarding_refresh`
  - `/plans/wo_onboarding_refresh`
  - `/agents/wo_onboarding_refresh`
  - `/review/wo_onboarding_refresh`
  - `/releases`
- Confirm no optional post-MVP scope has accidentally been added.

## 2. Build and Test

Run:

```bash
go test ./...
go build ./...
cd web && bun run test
cd web && bun run build
```

Confirm:

- backend tests pass
- frontend tests pass
- frontend build passes
- Go build passes

## 3. Demo Data Verification

Using a fresh seeded SQLite database:

- verify `Guided onboarding refresh` exists
- verify linked `Cluster`: `Onboarding activation clarity`
- verify the prototype note is present
- verify the seeded `Decision` and `Artifact` read cleanly
- verify draft plan generation works
- verify Agent Workspace actions persist new `Artifact`s
- verify `Mark release-ready` persists and `/releases` shows the Work Object

## 4. Environment and Config

Confirm required environment variables are set for the target environment:

- `PORT`
- `DATABASE_PATH` or production database equivalent
- `APP_PUBLIC_URL`
- `APP_MARKETING_URL`
- `WORKOS_API_KEY`
- `WORKOS_CLIENT_ID`
- `WORKOS_COOKIE_PASSWORD`

Confirm:

- WorkOS AuthKit config is valid
- the existing WorkOS auth boundary is preserved
- no local-only auth shortcuts are enabled in production unexpectedly

## 5. Frontend Embed Flow

Confirm the production build still follows the intended model:

1. build `web`
2. embed frontend assets into the Go server
3. build the single Go app binary

Verify:

- app routes load from the embedded frontend at `/`
- `/api/*` routes still resolve correctly
- SPA fallback works on deep links

## 6. Persistence

Confirm the production database setup matches the current app assumptions:

- migrations apply cleanly
- seeded demo data is only used where intended
- write paths persist correctly for:
  - Work Object status
  - Work Object priority
  - Work Object core fields
  - prototype attach
  - draft plan generate/edit
  - deterministic agent `Artifact`s
  - release-ready state

## 7. UI Bug-Bash

Do one short pass across the MVP routes and look for:

- awkward copy
- inconsistent spacing
- broken empty states
- button state glitches
- confusing duplicated labels
- missing context on the hero path

Recommended pass:

- `/work/:workObjectId`
- `/plans/:workObjectId`
- `/agents/:workObjectId`
- `/review/:workObjectId`
- `/releases`

## 8. Release Smoke Checks

Before shipping:

- boot the built app
- log in through the expected auth path if enabled
- deep-link to a seeded app route
- generate a draft plan
- generate an implementation brief
- mark the Work Object release-ready
- confirm `/releases` reflects the change

## 9. Deployment Decision

After the MVP deploys, choose one post-MVP direction:

1. get feedback on the current workflow
2. add hosted provider integration
3. expose Cleartify as a context API/tool surface for external agents

Do not pursue all three at once.

## 10. Suggested Exit Criteria

The MVP is ready to deploy when:

- the demo path is clear and repeatable
- tests and builds are green
- the embedded frontend serves correctly
- persisted state changes survive reload
- the team can explain the product in one sentence:

> Cleartify is the system of record for product execution context.
