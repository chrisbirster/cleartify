# Audit checkpoint 8: web experience, packaging, and operations

Revision: `12390da9504ae91821851d027556968268fa0290`. October 27, 2026. All first-party frontend, runtime, test, script, and deployment/configuration files read; supporting docs reviewed for product consistency but excluded from exhaustive line-by-line editorial/legal review. See coverage.csv for hashes and exclusions.

## F28 — P1/core journey: public profile route renders a blank page

`src/router.tsx` registers `/@:handle`. The installed router only recognizes a dynamic parameter when the path segment starts with `:`; it treats `@:handle` as a literal segment.

Reproduction: followed the landing page's Profiles link to `/@chris` in the browser. The root was empty and the page blank, without an application error view. `/privacy` rendered correctly. A Node-level check against the installed createMatcher returned null for createMatcher('/@:handle')('/@chris'), and a positive match for /privacy.

Fix direction: use a supported dynamic route plus an explicit @-prefix match/extraction strategy (or change the URL scheme deliberately). Verify actual profile rendering, missing-profile states and handle changes in browser tests. TypeScript/build success does not test route matching. Add a catch-all not-found experience so unmatched routes do not silently blank.

## Browser findings and product gaps

- Landing page and policy pages render. The landing story remains a texting assistant, not the owner's social-first platform.
- An isolated local browser waitlist submission initially failed with the default empty API base because Vite has no API proxy. Restarting with VITE_API_BASE_URL=http://localhost:8088 made the same synthetic submission visibly succeed. README mentions configuring that URL; make the complete local setup explicit rather than calling this an unexplained API defect.
- Profiles show only identity/bio when reachable. No social signup/profile editor, follow controls, protocol authoring/catalog/detail/configuration/export, friend/feed, or reserved-account-claim UI exists.
- Landing protocol cards advertise Weekly Reset, Trip Planning, Ship It as included, while beta installs remember-this/follow-through/weekly-reset keys. Those keys are not working protocol definitions or external chatbot integrations.
- Waitlist inputs have placeholders but no persistent labels. Beta code input also lacks an explicit label. Input styles remove focus outlines without a replacement; keyboard/screen-reader usability needs focused remediation.
- Marketing updates are prechecked. Decide the desired consent UX and final wording rather than presenting this implementation as an approved policy.
- Browser smoke did not establish a complete activation/login/social journey: missing routes/integrations and F22-F29 block that acceptance. No live PDS, user credentials, or notification providers were used.

## Packaging and operations

- Go API and SPA are separate deployable surfaces. Docker packages only the API, which is consistent with the documented architecture.
- Local Docker build succeeded using the repository's Go 1.26.1 builder; host checks used Go 1.27.1. No production container or service was deployed.
- Final uncached go test -p 1 -count=1 ./... passed after allowing the local TCP listeners required by existing HTTP integration tests. The restricted first attempt failed because localhost binds were blocked; that was an environment failure, not an application regression.
- go vet ./... and go mod verify passed. npm ci/npm audit reported zero vulnerabilities in the installed lockfile set; npm run build passed. This is not an independent dependency vulnerability/license/security certification.
- CI builds web, tests Go, and builds the container. It has no browser test job, so F28 is invisible to current CI. Feature branch naming in README (feature/*) differs from the push workflow (feat/**); PRs targeting dev/main still trigger checks.
- Fly configuration is a placeholder template, not deployment evidence, and uses liveness rather than SQL readiness. Optional hardening/readiness defaults off. Confirmed logging/startup defects are F01/F02.
- The closed temporary-database backup drill verifies a disposable copy; it is not a running production/Turso backup/restore system. Metrics/alerts are local libraries with no external operations integration.
- No account export/delete workflow or enforced retention jobs were found. Internal policy target constants/documents do not implement them.
- CREATE IF NOT EXISTS bootstrap is not a versioned upgrade system; deployment safety needs a migration ledger, explicit upgrades and restore tests on realistic historical data. Foreign-key enforcement needs F05's remediation first.
- OAuth's local versioned migration utility is not wired into the production SecretBox reader. Its tests cover atomic conversion/rollback; do not run it on live credentials before a compatible reader cutover and recovery plan.

## Completion boundary

The first-party source/config/test line review is complete at the recorded revision. Generated dependency metadata was verified by tooling rather than manually audited hash-by-hash; binary design artwork was excluded. Supporting historical/design docs received context review rather than exhaustive editorial/legal certification. Source review does not mean every feature is working: the consolidated backlog separates reproduced defects, static integration gaps, missing product capabilities, and real-world verification still outstanding.
