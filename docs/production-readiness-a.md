# Production Readiness Tranche A

## Scope

Tranche A is a stabilization and correctness pass.

It includes only:
- fixing the `Work Object` detail `null.map` crash by guaranteeing `Decision[]` and `Artifact[]` serialize as `[]` and are handled defensively in the UI
- fixing invisible or low-contrast `planning` and `low` chip styling
- replacing placeholder `example.com` demo links with coherent real-looking hosted links
- tightening `Review` and `Releases` copy so it is explicit how a `Work Object` enters `/releases`
- adding one focused hero-path smoke test that fails on JS runtime errors

## Non-goals

- no auth or onboarding work
- no `/me` or `/team`
- no CRUD routes or management surfaces
- no schema changes unless strictly necessary
- no new workflow logic or product scope expansion

## Routes / files likely to change

Frontend:
- `/work/:workObjectId`
- `/review/:workObjectId`
- `/releases`
- shared chip styling on `/work`

Likely files:
- `web/src/pages/WorkObjectDetailPage.tsx`
- `web/src/pages/WorkObjectsPage.tsx`
- `web/src/pages/ReviewPage.tsx`
- `web/src/pages/ReleasesPage.tsx`
- `web/test/fixtures.ts`
- `web/test/smoke/app-shell.spec.ts` or a sibling smoke file
- relevant frontend test files next to the touched pages

Backend:
- `internal/store/work_objects.go`
- `internal/store/seed.go`
- relevant store tests

## Backend work

- Normalize loaded `Work Object` payloads so `Decisions` and `Artifacts` are never `nil`.
- Keep the API wire shape unchanged other than returning `[]` instead of `null` for those collections.
- Update seeded demo URLs to coherent real-looking links:
  - GitHub links for code/PR context
  - hosted docs links for feedback/doc context
  - real-looking prototype links for prototype context

## Frontend work

- Make `Work Object` detail merge helpers robust against `null` or `undefined` `Decision` / `Artifact` arrays.
- Strengthen chip tone visibility for `planning` and `low` anywhere they render on the focus routes.
- Tighten release queue wording so a user can tell that:
  - `Mark release-ready` from review adds the `Work Object` to `/releases`
  - `/releases` is the queue of `Work Object`s that cleared that checkpoint
- Keep all changes presentational and correctness-focused.

## Tests

Backend:
- store test that loaded `Work Object`s always expose non-nil `Decisions` and `Artifacts`

Frontend:
- `Work Object` detail test covering `decisions` / `artifacts` missing or `null` input without crashing
- chip styling assertions for visible `planning` and `low` states
- updated rendering expectations for clearer release queue semantics

Smoke:
- a focused hero-path smoke test that:
  - opens `/work/:workObjectId`
  - opens `/plans/:workObjectId`
  - opens `/agents/:workObjectId`
  - opens `/review/:workObjectId`
  - marks the `Work Object` release-ready
  - opens `/releases`
  - fails on `pageerror`
  - fails on severe runtime `console.error`

## Validation commands

- `go test ./...`
- `go build ./...`
- `cd /Users/gm/dev/personal/cleartify/web && bun run test`
- `cd /Users/gm/dev/personal/cleartify/web && bun run build`
- `cd /Users/gm/dev/personal/cleartify/web && bun run test:smoke`

## Hard definition of done

- the hero path runs without JS runtime errors
- `Work Object` detail does not crash with missing `Decision` / `Artifact` arrays
- `planning` and `low` chips are visibly styled
- release queue copy clearly explains how items enter `/releases`
- seeded demo links no longer look like placeholders
