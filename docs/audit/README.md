# Cleartify social-first implementation audit

Completed source-review pass: October 27, 2026, against `dev` at `12390da9504ae91821851d027556968268fa0290`.

All 29 numbered findings now have local remediation (October 28, 2026). See [Remediation progress](remediation.md) and the [remaining 21 repairs](remaining-findings-repairs.md) for evidence and upgrade requirements. The report below preserves the original audited revision; feature/product and live deployment acceptance remain separate.

## Assessment

Cleartify has substantial local backend foundations, but it does not yet provide a usable social protocol-sharing platform. The first launch should prove social signup → profile → first follow → publish a configurable protocol → discover/configure/use it. Broad autonomous assistance, paid marketplace operations, and remote agent interoperability should be deferred until that loop works.

At the audited revision, the project was not ready for real onboarding: activation could bind an unchallenged phone and accept revoked invitations. Publication accessibility, installation identity, and browser routing also need correction. Passing tests establish the scenarios they cover, not acceptance of these missing transitions.

## Scope and evidence

- 293 tracked files inventoried with hashes. 203 source/config/test/fixture/context files read completely, including all 178 internal Go files and the executable/frontend. See coverage.csv for exact disposition and line spans.
- 87 supporting design/history documents excluded from exhaustive runtime line audit; product and architecture context/conflicts were reviewed. No exhaustive editorial, legal or third-party-code audit is claimed.
- Two generated dependency metadata files were checked through installation/module verification; one binary design artifact excluded from source review.
- 29 numbered findings include reproduced defects, static integration findings, and explicit product/policy gaps. These are not 29 equivalent security vulnerabilities. Detailed reports distinguish administrator-only, library-only, simulated, and ordinary user paths.
- The original audit changed no runtime source. Subsequent local repairs are tracked in [Remediation progress](remediation.md); the numbered findings and coverage hashes describe the original revision.
- No production deployment, live PDS login, messaging delivery, real commerce, or external chatbot integration was performed.

## Verification completed

| Check | Result and scope |
| --- | --- |
| Uncached `go test -p 1 -count=1 ./...` | Passed with local networking permitted for existing loopback tests |
| `go vet ./...` | Passed |
| `go mod verify` | All modules verified |
| `npm ci --ignore-scripts` / npm audit | Installed locked dependencies; reported zero known vulnerabilities in that run |
| `npm run build` | Passed |
| `docker build -t cleartify-audit:12390da .` | Passed; pinned Go 1.26.1 builder, local Docker build only |
| Local browser | Landing/policy rendered; configured waitlist submission succeeded; `/@chris` blank-page failure reproduced |
| Targeted synthetic probes | Error redaction, failed-start exit code, foreign keys, profile lifecycle, takedowns, validators, install/run identity, marketplace failure states, activation revocation/phone binding, login cooldown and agent lifecycle/grants |

Host Go was 1.27.1; container build used the repository's pinned toolchain. The initial restricted uncached run failed only at local TCP binds and was rerun successfully with permission. No live provider claims follow from these checks. Browser responsive/accessibility acceptance and load/race/fuzz testing remain separate verification work.

## Findings index

P1 means fix before enabling the affected path for intended real use. P2 means correctness/integration work required within the feature's eventual scope. Priority does not imply the feature must be included at launch.

| ID | Finding | Evidence/type | Launch relevance |
| --- | --- | --- | --- |
| F01 | Error objects bypass hardened log redaction | Reproduced, P1 | Foundation |
| F02 | Listener failure exits 0 | Reproduced, P2 | Foundation |
| F03 | Fake accounting flag/service unwired | Static integration, P2 | Deferred commerce |
| F04 | Local transport constructed/discarded | Static integration, P2 | Deferred interop |
| F05 | Foreign keys disabled | Reproduced in memory/file DBs, P1 | Foundation |
| F06 | New social placeholder cannot become public via current onboarding | Reproduced boundary/product gap | Core |
| F07 | Blocks stored without interaction enforcement | Reproduced/incomplete policy, P2 | Social trust |
| F08 | Hidden target cannot be unfollowed | Reproduced, P2 | Core |
| F09 | Paid-discovery query drops takedown filter | Reproduced, P1 | If flag enabled |
| F10 | General install ignores hidden takedowns | Reproduced service flow, P1 | Core authorization |
| F11 | Accepted published guides lack required install permissions array | Reproduced, P2 | Core protocol schema |
| F12 | Manifest-analysis flag rejects supported guides | Reproduced, P2 | Core protocol schema |
| F13 | Creator mutations omit explicit sync push | Static integration, P2 | Remote durability |
| F14 | Legacy installs authorize other creators' same slug/version | Reproduced, P1; admin request only | Canonical identity |
| F15 | Legacy approval survives install disablement | Reproduced, P2; approval only | Before runtime use |
| F16 | Guide ID installs conflict with slug-based request lookup | Reproduced, P2 | Protocol integration |
| F17 | Hidden publication completes fake execution | Reproduced handler flow, P1; no real side effects | Before real runtime |
| F18 | Purchase accepts draft/unclaimed seller | Reproduced against seeded state, P2 | Deferred commerce |
| F19 | Unlisted paid offer restores nonbuyer/refunded access | Reproduced/policy gap, P2 | Deferred commerce |
| F20 | Idempotency result-write failure permanently blocks retry | Fault-injected, P2; library-only | Deferred commerce |
| F21 | Accounting accepts reversal without refund | Corruption-injected, P2; library-only | Deferred commerce |
| F22 | Activation verifies a phone that never received its challenge | Reproduced, P1 | Before real onboarding |
| F23 | Revoked activation still creates account | Reproduced, P1 | Before real onboarding |
| F24 | Login cannot recover after failed-attempt cooldown | Reproduced, P2 | Private-login correctness |
| F25 | Ownership proof does not complete agent activation | Reproduced/integration gap, P2 | Deferred agents |
| F26 | Duplicate valid grants deny transport | Reproduced, P2; local-only | Deferred interop |
| F27 | Split-origin sessions/CORS omit credential/method support | Static + preflight probe, P1 | Browser identity |
| F28 | `/@chris` route does not match; blank page | Browser + matcher reproduction, P1 | Core |
| F29 | Activation refresh/delivery retry loses recoverable state | Static lifecycle trace, P2 | Before real onboarding |

## Feature inventory and recommended disposition

| Feature | Current state | Recommendation |
| --- | --- | --- |
| Social identity/signup | Existing-identity OAuth backend; real PDS acceptance outstanding; no complete new-user UI | Core; finish before cohort onboarding |
| Profiles/claims | Metadata reads and limited prelinked DID claim; broken frontend route; no editor | Core; implement ordinary profile setup first |
| Chris first-follow | Local hook and persistent opt-out tested; requires provisioned claimed/public Chris | Core; integrate with real signup and unfollow UI |
| Reserved creator accounts | Design exists; no complete verified reservation transfer workflow | Keep explicit unclaimed status; later controlled claim flow |
| Protocol creation/configuration | Bounded immutable JSON drafts; guide flag; no input/default configuration or Creator Studio | Core; define versioned readable runbook schema |
| Publication/discovery | Local visibility/audit/metadata services with access gaps; no profile catalog/search UI | Core; repair checks and expose useful content |
| Protocol use/export | Manifest and approval records; guide installs stored; no readable export/chatbot delivery | Core; implement one tested readable export/use path |
| Follows/blocks/reports | Local mutations/records; no complete user UI; block and hold gaps | Core social controls; define semantics |
| Feed/collections/friends/custom profiles | Product ideas, not implemented | Choose minimum social experience deliberately |
| Private Clear identity/login | Phone-first local services, challenge/session defects; no real sender | Keep separate from social signup; repair if retained |
| Full personal assistant | No messaging/model/memory/tool loop | Defer; Clear can start as a protocol helper |
| Local runtime/scheduler | Simulations/manual libraries; no real executor | Keep for development; no arbitrary-code launch |
| Agent ownership/grants/interop | Local cryptographic and SQL foundations; no external integrations | Defer from first social release |
| Paid marketplace | Fake-only checkout/accounting, UI/wiring/policy gaps | Defer real payments |
| Operations | Build/tests/container work; local drills; no verified live restore/alerts/retention | Add deployment-appropriate controls before release |

## Concrete repair/build sequence

### 1. Repair identity, privacy, and data invariants

Fix F22/F23 first. Correct log redaction, startup error reporting, per-connection foreign keys plus orphan checks, login cooldown/activation recovery if phone-first beta remains, and the supported browser origin/cookie architecture. Add regression tests that reproduce the actual transitions. Establish versioned migrations before changing persistent schema.

Acceptance: challenge destination cannot change unnoticed; revoked invites create no users; failed starts return nonzero; keys/tokens in errors cannot reach logs; relation constraints work after reconnect; intended browser sign-in persists correctly.

### 2. Standardize protocol identity and accessibility

Use one publication/version identifier across installations, approvals and use, with pinned content digest. Centralize visibility, takedown, hold and entitlement predicates. Make guide validation/manifest/configuration rules agree. Add endpoint-level tests for revocation/disable/takedown between each lifecycle stage.

Acceptance: another creator's same slug cannot reuse an install, disabled/hidden versions cannot progress, and a supported guide can traverse the full lifecycle under intended flag combinations.

### 3. Finish one social browser journey

Repair the profile route; add supported AT Protocol signup/login, public profile setup/editing, Chris follow/removal, a protocol authoring page, public profile protocol catalog, detail/configuration page, and readable export/use with one supported chatbot. Default private data stays private. Use the supplied neon/social reference for personality and layout; do not inherit its entire feature list.

Acceptance: a creator and a second ordinary user complete the journey through the browser, without SQL seeding or administrator intervention for routine free sharing.

### 4. Validate with a small real cohort

Run real hosted-PDS login/revocation and browser acceptance in an isolated authorized environment; exercise moderation, consent, support, backup/restore and rollback. Start with original Cleartify/community creators; do not depend on celebrities claiming profiles. Keep payments and remote agents off until each has a proven use case and dedicated acceptance evidence.

## Detailed reports

- 01-foundation.md — startup/config/storage
- 02-social-identity.md — profiles/follows/account setup
- 03-protocol-lifecycle.md — drafts/publishing/install/use
- 04-execution-identity.md — linking/run authorization
- 05-marketplace.md — simulated commerce/accounting
- 06-agents-interop.md — ownership/grants/local messages
- 07-onboarding-moderation.md — activation/login/moderation
- 08-web-operations.md — browser/package/deployment limits

The audit is complete as a source-review pass, not a claim that the listed defects are fixed or the launch acceptance criteria are met. Fixing runtime source is the next implementation phase.
