# Cleartify M0–M10 — merged-code status and remaining gates

**As of October 13, 2026.** Audited against `dev` at `2947cf86e620dc78062ab7a406b33589af16dbc7`. GitHub confirms **#6 and #32–#48 merged**, including the final implementation sequence #33–#48. This is a repository/code status, **not** evidence of staging verification or launch approval. Earlier PRs #7–#24 and #26–#31 provide baseline foundations. The old 88-task tracker was a **normalized reconstruction**, not an authenticated copy of an original 88-item checklist; this document replaces its obsolete PR-by-PR claims rather than fabricating completion counts.

**Status rules:** **done** = nothing remains within the milestone's agreed deliverable, including necessary approvals; **code complete, owner-blocked** = all in-scope engineering is in `dev` and only explicitly owner-controlled verification, approvals, credentials, or deployment remain; **partial** = additional engineering can be done without owner-only action. A local prototype, schema, unit test, or green CI is **not** a finished production feature. Flags default OFF where specified; owner decisions in M0 do not turn development defaults into public legal promises.

| Milestone | Status | Merged foundation | Main remaining boundary |
|---|---|---|---|
| **M0 Product contract** | **done** | #6 owner decisions and internal retention defaults, owner sign-off | Future public policy/counsel review belongs to M1/M10, not an open M0 decision |
| **M1 Beta/onboarding** | **partial** | Waitlist, invitation, activation, UI, basic abuse controls, draft policy pages | Stronger distributed abuse protections and production-ready privacy/terms; owner/counsel approval |
| **M2 Identity/OAuth** | **partial** | Flagged ATProto OAuth, Clear sessions (#33), DID-link persistence (#35), design (#32) | Actual authenticated DID-link ceremony, real-PDS tests, owner-controlled staging |
| **M3 Social/trust** | **partial** | Profiles/follows, reports/blocks (#36), anti-impersonation hold library (#47) | Wire moderation holds into runtime, review/appeal and real handle verification |
| **M4 Creator protocols** | **partial** | Discovery, immutable drafts (#37), preview permission guard (#38), draft validation/remix fixtures (#46) | Publishing workflow, verified remix provenance/license checks, 10 approved seed protocols |
| **M5 Clear runner** | **partial** | Version-pinned install approvals, bounded fake local runner (#41), audit/cancel/retry | Authenticated install-to-run integration, durable audit, real isolated execution boundary |
| **M6 Private alpha** | **partial** | Alpha memberships/gate (#34), Clear-session-protected private routes (#48), smoke docs | Extend gates to all future private actions, staging beta smoke, cohort/use metrics, owner approval |
| **M7 Agent registration** | **partial** | Registration, safe directory (#39), local owner-key binding and scoped grants (#42) | Trusted authenticated runtime ownership ceremony and self-hosted adapter |
| **M8 Agent collaboration** | **partial** | Ed25519 envelope/replay (#40), key rotation, consent checks, local two-agent test (#43) | Authenticated replay-safe transport, transactional revocation/consent, external interop |
| **M9 Marketplace** | **partial** | Listings/pricing, fake-only seller/entitlement/refund/audit library (#44) | Authenticated checkout/access enforcement; real provider only with owner approval; seller/compliance decisions |
| **M10 Release/operations** | **partial** | Launch checklist, source-SHA CI check, flag-off monitoring, rollback/backup docs (#45) | Real observability/alerts, disaster-recovery drill, staged E2E and launch sign-off |

**Overall: 1 done (M0), 0 code complete/owner-blocked, 10 partial.** These statuses reflect remaining implementable work, not lack of effort or CI coverage. Individual tasks within a partial milestone can still be code complete and owner-blocked.

## M0 — Product contract — **done**

- **Done in merged docs:** #6 establishes the social-first product direction, Clear/private-vs-social model, AT Protocol social login requirement, default Chris follow that users may remove, protocol authorship product boundaries, and five owner answers. The owner confirms **#6 merged** and personally signed off the M0 decisions and internal retention engineering defaults.
- **Owner-only downstream boundaries:** The October 13 retention durations are **internal engineering defaults**, not published policy, legal advice, or promises. Counsel and owner must approve any retention/deletion language incorporated into public `/privacy` or `/terms` (tracked in M1/M10).
- **Not built in M0:** No outstanding M0 code deliverable or M0 decision. Implementing later product features is not a reason to reopen M0.

## M1 — Beta onboarding and consent — **partial**

- **Done in merged code:** Waitlist, referral capture, beta invite/start/verification, landing/profile previews, lockfile/build CI, baseline signup/mutation throttling and origin/CSRF checks (#7–#16, #31), and dev-only draft policy/acceptance foundations (#12).
- **Built up to owner boundary:** Activation flow can be tested locally; public terms/privacy, user-data retention/deletion handling and production invitation process require counsel/owner sign-off and controlled staging.
- **Not built/verified:** Distributed abuse/rate-limit protections, production consent/legal acceptance verification and approved public notices. No real staged beta/onboarding validation is recorded.

## M2 — Separate Clear account and AT Protocol identity — **partial**

- **Done in merged code/docs:** DID primitives, origin restrictions, encrypted ATProto-related storage, SDK/toolchain pinning, flag-off OAuth (#18), revocable hashed Clear sessions with 30-day idle limit (#33), transactional one-to-one DID/Clear link storage and challenge/audit primitives (#35), and DID linking design (#32).
- **Built up to owner boundary:** Clear session security and persistence primitives and flag-off OAuth; **real hosted-PDS** authentication/refresh/restart/handle proof, isolated credentials and staging are owner-only verification. Retention defaults are approved internally, not public legal wording.
- **Not built:** Full consented and authenticated DID↔Clear linking ceremony and API, disconnect UX, account recovery/key-rotation semantics and end-to-end linked identities. A stored link or local DID string is **not** real-PDS verification.

## M3 — Profiles, social graph, moderation and trust — **partial**

- **Done in merged code:** DID-keyed profile/discovery reads, local follow/unfollow, default-follow opt-out (#26–#28), DID-authenticated registration/claim groundwork, locally stored reports/blocks (#36), and disabled-by-default reserved/lookalike handle screening, locally-claimed-vs-unverified labels and reversible report-threshold holds (#47).
- **Built up to owner boundary:** Local moderation and hold models; owner/security approval for thresholds, appeals, false positives, disclosure and independent DID/handle verification.
- **Not built:** Automatic hold enforcement in public discovery, follow/claim or publishing; moderator review/appeals UI and durable case audit, stronger confusable detection, robust report abuse resistance and PDS-backed handle proof. Label `locally_claimed_not_independently_verified` is deliberately **not** a verified identity badge.

## M4 — Creator protocols, remix and attribution — **partial**

- **Done in merged code/docs:** Immutable draft/publication schema and safe discovery, DID-scoped draft creation (#37, default OFF), local preview permission allowlist (#38), bounded declarative draft validation, remix metadata/attribution policy defaults and **two original Cleartify-only test fixtures** (#46). Drafts do not automatically publish.
- **Built up to owner boundary:** Internal license and attribution defaults plus sample fixtures; owner/counsel must review rights, source provenance, remix consent, takedowns and proposed public license wording before any publishing.
- **Not built:** Authenticated creator **publishing/review** lifecycle, database-backed upstream remix source/license compatibility checks, full attribution lineage and moderation/takedown process; **10** original owner-approved licensed/seeded protocols (only **two non-published test fixtures** exist).

## M5 — Install approvals and Clear runner — **partial**

- **Done in merged code:** Run-request and install approval schemas/handlers, exact-version/spec-digest permission acknowledgement (#29), flag-off local permission guard (#38), and bounded declarative fake in-memory executor (#41) with local-only operations, cancel/retry and ordered **non-durable** audit events.
- **Built up to owner boundary:** Safe local demonstration of run semantics; owner must approve consequential permission taxonomy, execution isolation/security, install activation and consent before real use.
- **Not built:** Authenticated Clear user **installation-to-run** binding, durable run state/audit and scheduler, real isolated worker/executor, concurrency-safe cancellation, retry/recovery across process restarts and actual usable install. Approval records do **not** run protocols.

## M6 — Private alpha access and evaluation — **partial**

- **Done in merged code/docs:** Alpha membership and activation eligibility (#20), flag-off authorization gate (#34), verified Clear-session-principal wiring for the currently identified private Clear status and eligibility routes (#48), cross-user spoofing and revoked membership/session tests, staged smoke checklists (#30).
- **Built up to owner boundary:** Real protected routes fail closed on absent dependencies or revoked access **when flags are on**; staged beta/hosted-PDS smoke, credentials, tester cohort, actual alpha go/no-go remain owner-controlled. Logout stays accessible.
- **Not built/verified:** Broad auth policy for **future private Clear actions**, complete DID-linked private data/runtime permission checks, tester invitation operations, first-run/repeat-use measurements and real staging E2E proof. Existing admin Clear actions retain **separate admin authorization**, not a private user gate.

## M7 — Verified agent identity and scoped grants — **partial**

- **Done in merged code:** Local admin-only unverified agent registrations (#21), safe flag-off directory with no independent-verification claim (#39), signature-based agent key possession/owner-DID binding, scoped revocable grants with expiry and tests (#42).
- **Built up to owner boundary:** Owner-binding library requires an **already trusted, independently authenticated DID caller**; no real PDS assertion or public agent verification is established. Owner must approve claim semantics, scope/security policy and listing representation.
- **Not built:** Trusted authenticated API for registering/rotating owner identity and managing grants, secure signing-key enrollment ceremony, self-hosted third-party adapter, durable authorization audit and proof of independent agent ownership.

## M8 — Agent-to-agent interop — **partial**

- **Done in merged code:** Signed Ed25519 envelope, local key/replay registry (#22/#40), local-only signing-key rotation, time-limited recipient consent checks plus sender-scoped-grant check, and two-agent local interop test (#43).
- **Built up to owner boundary:** Offline in-process staging works without transport or execution; owner must approve inter-agent privacy, consent/delegation and trust assumptions.
- **Not built:** Actual authenticated transport, remote agent interoperability, secure runtime integration, multi-device delivery acknowledgement, key recovery, and atomic consent/grant checks concurrent with revoke. Low-level replay API is **not** a production consent boundary.

## M9 — Seller marketplace and payment foundations — **partial**

- **Done in merged code:** Listing/pricing schema and constraints (#23), disabled-by-default seller match against listing/protocol creator, fake-provider-only interface, local entitlements, simulated refund and audit tables/tests (#44).
- **Built up to owner boundary:** Entire commerce path is **fake and library-only**, no provider keys or money movement. Actual provider, country/tax/KYC/payouts, seller contracts, fees/refund/legal conditions and limited paid-pilot authorization are owner/counsel/finance gates.
- **Not built:** Authenticated production checkout/seller flows, entitlement enforcement for real installations/access, payment idempotency and reconciliation, real refunds/chargebacks, real provider integration and any paid creator pilot. Simulated receipt is **not** financial proof.

## M10 — Production readiness, observability and launch — **partial**

- **Done in merged code/docs:** Public-launch checklist (#24), origin/CSRF and baseline mutation hardening (#31), existing CI Go/web/Docker builds, source-checkout SHA validation script in CI, default-off in-memory HTTP status counters and rollback/backup runbook (#45).
- **Built up to owner boundary:** Owner/security/counsel must authorize real staging, verify backups and restoration feasibility, select incidents/alert recipients, finalize public privacy/terms, review security and accept a known release SHA. Runbook is **documentation**, not evidence of a restore or rollback drill.
- **Not built/verified:** Operational metrics exporter/real incident alerts, proven recovery drill, full security and staging end-to-end smoke, final release artifact provenance, public policy publication, launch readiness evidence and explicit owner go/no-go. No production deployment or launch was authorized.

## Cross-cutting delivery boundaries and accounting

- **Green CI establishes build/test success, not launch:** never treat mocked keys, fake payments, an in-memory audit or internal docs as real-world operation.
- **Flags stay default OFF**, including private alpha, creator draft APIs, permission guards where configured, agent directory, moderation, OAuth and optional hooks. Library-only fake-runner, consent, impersonation and commerce pieces have no newly exposed production handlers.
- **Owner-only:** real PDS access and validation, credentials/secrets, staging deployment and smoke tests, actual money/provider selection, counsel-approved legal language and retention text, incident/recovery exercise, public launch authorization.
- **Engineering still open:** identity-link UI/API and recovery, moderation enforcement/appeals, creator publishing/remix checks, real guarded Clear installation runtime, agent transport and authenticated owner/grant/consent wiring, marketplace auth/access integration, stronger monitoring, distributed abuse controls and operational tests. These are why **M1–M10 are partial**, even though code slices and tests are merged.
- **Audit trail:** relevant merged PRs **#6, #12, #18, #20–#24, #26–#48** (plus early foundations #7–#19 and #25 as applicable). This is a dated snapshot of `dev`; re-audit after future changes.
