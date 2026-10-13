# Cleartify M0 — Social Marketplace Product Contract
Date: 2026-10-12
Status: Proposed product direction; implementation not yet delivered
Owner: Chris
Scope: M0 reconciliation / planning, not code implementation

## One-sentence definition
Cleartify is an independent social discovery and protocol marketplace for humans and AI assistants: follow creators and agents, discover useful, versioned protocols, and run approved workflows with Clear or compatible connected assistants.

## Positioning
- Public-facing promise: **Your AI needs friends.**
- Product explanation: Find people and agents worth following; adopt practical protocols they share; put those protocols into practice with your assistant.
- Experience: playful early-web/MySpace-style personality (Chris is the removable, default first friend), modern accessibility, no mandatory mobile app download.
- Clear remains Cleartify's own text-first assistant, not a requirement for every marketplace interaction.
- Avoid claiming partnerships with named creators or technical integrations with proprietary assistants unless verified.

## Users, jobs and value
1. **Explorer (launch primary)**: discovers trustworthy practical methods; wants to apply them without manually translating content into a routine.
2. **Creator (supply side)**: publishes an original, licensed protocol with attribution, updates and social discovery.
3. **Agent operator (later)**: claims a verified agent profile and connects a compatible endpoint to accept scoped protocol work.
The launch wedge is **people-to-protocol discovery-to-first-run**, not a general agent chat app and not merely following bots.

## Primary alpha journey
1. Visitor sees the public landing page and understands humans + agents + protocols.
2. Visitor joins the existing waitlist; invited user authenticates using a supported method.
3. User browses Chris's public profile and a small set of original protocols.
4. User installs a protocol, with a plain-language permission/retention explanation.
5. User configures and completes one run through Clear's web-text experience.
6. User sees their private result; can optionally share a sanitized public update or protocol link.
7. Creator can update the protocol; existing installs remain pinned or ask for user approval before a breaking change.
**Launch proof:** a stranger discovers a creator's protocol, runs it, obtains value and shares a link.

## Entities and permissions
- **Person:** internal user ID; optional, explicitly linked ATProto DID; profile consent/visibility; privacy settings.
- **Agent:** unique internal ID, owner, provider label, *directory-only vs connected vs verified* status, capability metadata, allowed scopes, revocation and rate limits. A social follow never grants execution authority.
- **Protocol:** author, rights/license, attribution, category, declarative version, inputs, triggers, permissions, safety notice, changelog, distribution visibility. Public record may reference but never contain secrets or users' private run state.
- **Installation/run:** private user-specific config, version pin, consent, state, outcomes and audit trail.
- **Follow:** explicit source and destination; distinguish person follows from agent subscriptions; reversible, blockable and reportable.
- **Claims:** never impersonate a creator; unclaimed biographies are clearly labeled and removable.

## Product scope
### M1 (already underway)
Waitlist, landing page, Go/Turso API, CI and invite activation. Update the landing copy to explain the marketplace vision after validating existing design.
### First social alpha (M2–M6)
Authentication/account linking; human and creator profiles; follows/block/report; public protocol discovery/search; creator publishing/versioning; one constrained Clear protocol runner; private installation and progress; shareable URLs; a 30–50-user pilot.
### Later
BYO connected agents; agent-to-agent subscriptions and A2A task exchange; DEEZ authority envelopes when delegated execution actually needs them; payments and creator payouts; phone messaging channels beyond baseline.
### Explicitly excluded from alpha
Arbitrary uploaded code execution; automatic access to third-party Dot/Instinct/Grok/Muse agents; mandatory PDS self-hosting; raw private memory on public ATProto records; publicizing health data/activity by default; speculative endorsements; an unrestricted social feed or DMs; marketplace payments before demand.

## Data policy (technical boundary, not final legal retention policy)
**Public/portable, opt-in:** creator profile, published protocol metadata/version, public follow where chosen, voluntary sanitized activity. Candidate system of record: ATProto repository on selected PDS; Cleartify search index is rebuildable.
**Private/operational:** waitlist/consent, verified phone/email, memory, conversations, protocol installs/config/run state, approvals, secrets, entitlement, billing, audit and abuse records. System of record: Go/Turso.
**Media:** do not ship retained media in alpha. Future instant interpretation and short original TTL require explicit retention/deletion notice; optional persisted media requires encryption and a separately approved policy.
**User agency:** export, delete, unlink, unfollow, revoke and block must have clear behavior; define backup/log retention and legal holds before claiming complete deletion.
**Identity:** DID identifies optional ATProto social account; internal account ID is stable for private state; phone number is a channel address and alone insufficient for high-risk operations.

## Security/safety invariants
- Default private protocol runs; no implicit publication.
- Untrusted protocol text, public social records and agent messages never become system-level instructions.
- Declarative allowlisted runner; no arbitrary code or unknown tool privileges.
- Explicit human confirmation for purchases, disclosures, messaging third parties, financial or physical-world consequential actions.
- Versioned permissions and revocable grants; maintain execution receipts, bounded retries/idempotency and audit.
- Creator claims, moderation, IP complaints, spam prevention and blocks available before public scale.

## Metrics and alpha gates (proposed, not commitments)
- 30–50 invited humans and >=10 original/licensed protocols from >=5 creators.
- >=50% of activated testers start a protocol; >=30% return in week 4.
- At least 10 unaided complete install-to-run journeys, with 0 critical privacy/approval failures.
- Measure protocol discovery source, install-to-first-run conversion, repeat runs, opt-in shares and manual support.
- Do not count protocol install as a completed outcome or an agent profile as a connected integration.

## Open decisions / constraints
- Choose between optional ATProto linkage (accepted ADR-0007) versus ATProto-required social registration (new proposal). Do not silently replace the accepted ADR.
- Verify hosted PDS signup availability, OAuth/go SDK maturity and account recovery by end-to-end spike before promising seamless onboarding.
- Determine initial original protocol and measurable real-world result, and whether Clear web text or SMS is the first execution UI.
- Reconcile default Chris-follow with informed consent, ability to unfollow and provider-specific follow semantics.
- WorkOS MAU and SMS MFA economics should be rechecked only if conventional auth becomes a requirement.
- Creator payment, provider API permissions and commercial terms are not validated.

## M0 sign-off checklist
- [x] Audit repo/main vs dev, merged foundational PRs and existing ADRs.
- [x] Define product thesis and user types in this contract.
- [x] Define people/agents/protocols domain boundaries and the initial alpha journey.
- [x] Specify private/public data policy, safety, exclusions and measurement.
- [x] Record conflicts with accepted architecture as an explicit follow-up decision.
- [ ] Owner approves any change to ADR-0007 onboarding default.
- [ ] Product docs and README are updated after that approval.
- [ ] CI/production deployment are verified separately under M1.

## Repository baseline from audit (2026-10-12)
- `main` contains only the reset README. `dev` includes frontend, Go/Turso backend, docs and CI.
- PR #2 M0, #3 DEEZ ADR, #4 M1 waitlist slice 1 are merged into `dev`, not necessarily `main`.
- Current frontend routes `/` and `/@chris`; landing visual reference: `docs/assets/landing-page-mvp.jpg`.
- Existing accepted ADR-0007 establishes phone-first private users, optional ATProto social identities.
- M1 slice 2 email invitations, phone activation and welcome SMS are not proven delivered.
- This planning work makes **no claim** that CI is passing, deployment is live or the alpha exists.
