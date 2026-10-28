# Social-first product and implementation audit

Prepared October 27, 2026. Starting revision: `12390da` on `dev`.

This is a proposed work plan based on the current conversation. It does not mark features verified, amend accepted ADRs, or authorize deployment. The audit starts with read-only review; fixes follow as discrete, reviewable changes.

## Product baseline

Cleartify is a people-first social platform for discovering, creating, configuring, and sharing reusable AI workflows called protocols. It should feel like a community, with expressive profiles and a playful Myspace/Newgrounds influence. Chris is the automatic, removable first follow. AT Protocol supports social identity. Clear may help users understand, configure, and try protocols; the exact scope of its personal-assistant role remains a product decision.

The attached MemeWave image is a visual reference for personality, profile prominence, navigation, and social layout. Its photos, music player, messaging, groups, online presence, and general posting features are not automatically launch requirements. The reference should be adapted around people and protocols, including a mobile layout and accessible colors, focus states, and motion controls.

Proposed first proof: a person signs up, follows Chris automatically, edits their profile, publishes a protocol, and another person discovers, configures, and successfully uses that protocol in a supported chatbot. A readable export is a candidate first delivery mechanism; execution and compatibility claims require testing.

## 1. Reconcile the product contract together

Read the newer social product contract and recorded owner decisions against the older README, MVP plan, architecture, and ADRs. Preserve the history while making one current direction easy to find.

Decisions to settle:

- Is Clear a protocol helper, a full personal assistant, or a later capability?
- Are relationships follows, mutual friends, or both? Which belong at launch?
- What social activity is needed initially: protocol publications, updates, remixes, collections, or general posts?
- What is the minimum protocol format: purpose, inputs, defaults, instructions/steps, expected result, constraints, version, author, and attribution?
- Which chatbot and delivery method will be tested first? Separate a readable runbook from a verified automated integration.
- How should a new user create an AT Protocol identity, versus sign in with an existing one?
- What do reserved, community-curated, claimed, verified, and official mean? A community adaptation must not imply creator authorship or endorsement.

Output: one current product contract and a small launch scope. Until these decisions are settled, do not remove existing features solely because they appear secondary.

## 2. Build a complete feature inventory

Trace every registered route, frontend screen, configuration flag, service, database table/migration, background/library capability, external adapter, and test to its feature. Include code that exists but is not wired into the executable.

For each feature record:

| Field | Meaning |
| --- | --- |
| Intended behavior | A user-visible action and its expected result |
| Product disposition | Core now, useful later, unresolved, or retire candidate |
| Implementation state | Planned, library-only, API-only, UI-connected, or externally integrated |
| Evidence | Exact revision, files/lines, checks performed, and observed result |
| Dependencies | Identity, flags, permissions, data, providers, and startup wiring |
| Failure behavior | Invalid inputs, ownership mismatch, expiry, restart, and dependency failure |
| Defects and next action | Severity, reproduction, proposed fix, and verification |

Do not reduce these independent dimensions to one completion percentage. Existing code, passing unit tests, browser usability, and live provider validation are different evidence.

Initial observations to validate, not final acceptance:

| Area | Starting point |
| --- | --- |
| Web | Five routes: landing, privacy, terms, beta activation, profile preview |
| Social identity | Flagged OAuth/account registration; hosted-PDS verification documented as outstanding |
| Profiles/follows | Local backend foundations; first-follow opt-out; limited claim flow |
| Protocols | Draft/publish/discovery/install foundations; full authoring and portability experience needs review |
| Reserved creator profiles | Documented design; actual seed data and identity-transfer workflow must be checked |
| Clear/onboarding | Waitlist, beta, private sessions; notification senders currently disabled/debug |
| Execution/interop | Bounded local simulations and libraries; external delivery not established |
| Commerce | Fake provider and accounting; no real-money processing established |
| Operations | CI, templates, readiness and local drills; deployed behavior not established |

## 3. Review code line by line, feature by feature

Review in dependency order, including tests and configuration rather than only application code:

1. Startup, config, security primitives, database schema/migrations, persistence/sync.
2. AT Protocol OAuth, account creation, sessions, identity linking/recovery.
3. Profiles, reserved handles, claims, first follow, follows, blocks, reports, moderation.
4. Protocol schema, configurable inputs, drafts, publication, discovery, versions, provenance, remixes, takedowns.
5. Installation, permissions, approvals, guide delivery/export, access boundaries.
6. Frontend journeys, data requests, accessibility, mobile behavior, errors and empty states.
7. Clear, waitlist, beta and notifications, reviewed against the agreed product role.
8. Agent ownership, execution, scheduler, interop, commerce and accounting, with explicit simulation boundaries.
9. Operations, Docker, CI, backups, retention/deletion, logging, feature flags and documentation.

For each section explain to the owner: what it does, why it exists, what the code actually guarantees, defects, missing product behavior, and the recommended disposition. Ask for product choices at meaningful boundaries; routine technical review does not require line-by-line owner approval.

Maintain a coverage ledger of files and reviewed line spans at a specific revision. Record unreviewed areas explicitly; when code changes, re-review affected spans and their callers. Do not describe sampling as a complete line-by-line audit.

## 4. Verify actual behavior

- Reproduce meaningful defects before fixing them where feasible. Test user outcomes and trust boundaries rather than mirroring implementation.
- Run existing Go tests and frontend build at the recorded revision. Previous checks passed during this chat, but that is not feature acceptance.
- Exercise browser journeys: signup/login, profile editing, first-follow removal, protocol authoring, discovery, customization, and use/export where implemented.
- Check unauthorized and cross-user access, disabled flags, missing dependencies, replay, version changes, permission escalation, duplicate submission, expiry, cancellation, and recovery where relevant.
- Recheck the earlier audit's suspected cross-origin session issue, proxy-based signup throttling, and unwired fake-accounting configuration; record reproduction or dismiss with evidence.
- Verify database changes survive restart and that important writes, approvals, and audit events remain consistent when sync or delivery fails.
- Separate local fake tests from real provider tests. Prepare an isolated real-PDS/browser test when local checks pass; deployment or external writes remain separate actions requiring appropriate authorization.
- Define supported chatbot compatibility through concrete sample protocols and observed results. An agent directory entry or signature test alone does not prove an assistant can consume a protocol.

## 5. Fix and prioritize

Priority order: unauthorized access/data loss, broken core journeys, persistence/identity errors, confusing or misleading behavior, accessibility/usability, then secondary capabilities.

Create small fixes with a reproduction, user impact, focused validation, and documentation updates. Keep unsupported capabilities off. Do not delete later-stage infrastructure during the audit without a separate decision.

Recommend a launch backlog centered on real social signup, profiles/first follow, protocol authoring/discovery, configuration, and one verified use path. Decide on feeds, friends, comments, collections, and profile customization explicitly rather than copying all navigation from the image.

## 6. Completion criteria

- Every tracked source/config/test file is reviewed or explicitly excluded with a reason.
- Every inventoried feature has a disposition and evidence-based implementation/verification status.
- Core user journeys have reproducible acceptance checks; unresolved failures are visible.
- Documentation agrees on the primary product and distinguishes local, simulated, and real behavior.
- No critical unresolved defects remain in the intended release scope; other limitations are recorded.
- The owner can review a concrete launch scope, remaining backlog, and demo before any release decision.

Next collaborative session: resolve section 1, then review the route/flag/data inventory and the startup/database/authentication path. Continue through the ledger in manageable sections.
