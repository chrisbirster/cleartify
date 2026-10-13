# M0 product contract — Cleartify Social + Clear
Date: 2026-10-12
Status: **Proposed product-direction reconciliation for review** (not an amendment to accepted ADRs)
Base: `dev`

## 1. Product thesis
Cleartify is a people-first social discovery network and marketplace for **humans, AI assistants, and reusable executable protocols**. Users discover creators and assistants, follow public updates, install a versioned protocol, and run it privately through Clear or a compatible connected agent. The critical loop is **discover person → inspect protocol → install → run → optionally share outcome**.

Clear remains a text-first personal assistant; **no app download is required to use Clear**. The website exists for social discovery, publishing, identity linking, consent, permissions, billing, and account management. “No required app” does not mean “no web authorization step”.

## 2. Intended users / participants
- **Learner/operator:** wants to put a trusted person's method into practice, not just consume content.
- **Creator:** publishes original or properly licensed, explainable, maintainable workflows.
- **Agent owner/developer:** lists an assistant and, only with a supported API and verified authorization, connects it to run protocols.
- **Clear user:** can use a private phone-first assistant without being forced into public social identity.

Person, assistant, and protocol are separate entities. A directory listing **does not imply technical integration** or creator endorsement. Agent-to-agent following means only access to opted-in public updates, not access to memory, user accounts, or execution privileges.

## 3. Launch promise and non-promise
**Promise to prove:** A new user can find a real creator's public protocol, understand what it does and what permissions it requests, install a specific version, complete one useful run with Clear, and share a link voluntarily.

**Not promises:** all proprietary assistants are connectable; all creators officially participate; protocols guarantee life outcomes; human or agent follows permit arbitrary delegation; user memory is publicly portable; ATProto hosts private blobs free.

## 4. Data and identity boundaries
| Datum | Canonical home | Visibility and rule |
| --- | --- | --- |
| Public social DID, handle, opt-in profile, public follow, public protocol metadata | AT Protocol/PDS when supported | Assume public and replicable; indexed in disposable Cleartify view |
| Protocol definition/version | Public ATProto record for freely distributable public material; private/paid payload stored under controlled access | Immutable versions; provenance and license required; never expose paid payload in a public record |
| Private install configuration, run state, user inputs, private outcomes | Turso | Private by default; never infer public sharing |
| Messages, memories, phone/email, invitations, sessions, approvals, grants | Turso | Minimal collection, explicit retention/deletion rules required before launch |
| Assistant listing/capabilities and connection grants | Public description may be indexed; verified grants and credentials in Turso | “Listed”, “claimed” and “connected” are distinct statuses |
| Attachments | Not part of M0/MVP | Future TTL and encrypted object storage require separate reviewed policy |

Store stable DID for ATProto-linked identities; handles can change. Keep an internal Cleartify user ID for private state and explicitly link/unlink identities. A follow does not authenticate an agent endpoint. Do not write sensitive profile progress or private protocol runs to public PDS records. A PDS or social index is not a substitute for private authorization.

## 5. Explicit proposed decision on identity
**Recommended hybrid, pending Chris's approval:** Keep ADR-0007's private Clear identity path (verified mobile / internal ID) while introducing **ATProto OAuth for Cleartify Social onboarding at the earliest social milestone**. For visitors who only want to run Clear by text, do not require an ATProto account. For public social features, use supported hosted PDS account creation and OAuth, without operating our own PDS. WorkOS is *not* an MVP dependency until an additional account/auth requirement is demonstrated.

This clarifies—not silently supersedes—ADR-0007. A future ADR must resolve whether ATProto identity is required for all social accounts and whether a phone number is required before protocol execution.

## 6. Narrow first end-to-end alpha
1. Open public human profile `/@chris` and view authored sample protocol.
2. Create/link social identity through supported ATProto OAuth or use a clearly labeled limited browse-only path.
3. Inspect protocol purpose, publisher, provenance, version, steps and permission manifest.
4. Install an immutable free protocol version into a **private** Clear account.
5. Run one bounded `Weekly Reset` protocol, with user prompts and durable state, no external high-risk actions.
6. End with a visible, accurately recorded completion result; offer **optional** public share with explicit preview and consent.
7. Revisit public profile and share a working protocol link; another user can repeat steps 2–6.

This is a **proposed alpha acceptance contract**. The approved October 8 roadmap currently prioritizes text-first beta activation; changing its milestone numbering requires a separate decision.

## 7. M0 verification against repository
- October 8 accepted ADR-0007 sets phone-first Clear / Turso private plane / optional ATProto social plane.
- ADR-0006 defines delegated authority semantics via DEEZ/NUTS; it is not a shipping requirement for public social follow.
- Existing `docs/mvp.md`, `docs/architecture.md`, `docs/protocols.md`, `docs/profiles-and-marketplace.md` cover foundational architecture and product behavior.
- `dev` README documents SolidJS 2 RC, Go/Turso waitlist API, `/` landing, and `/@chris` preview.
- PR #2 M0 and PR #3 DEEZ describe merged foundation per prior project status; GitHub search metadata reviewed but current live CI, deployment, actual merges and latest branch checks **were not independently validated by this document**.
- PR #4 is M1 waitlist foundation; its body documents slice-1 scope and tests on its feature branch, **not proof of current production deployment**.
- `main` README remains the minimal reset text; `dev` contains the richer docs. Do not claim the new product direction is deployed to `main`.

## 8. Product and trust constraints
- Default private; never reveal an individual's active protocols or outcomes without explicit publication.
- Distinguish official/claimed/verified/unclaimed profiles. Remove/correct seeded profiles readily. No implied endorsement, fake partnerships, false “currently running” claims or unauthorized premium creator sales.
- Community interpretations of well-known methods need attribution, original expression, licensing checks, and special care for health claims.
- No arbitrary uploaded code execution in the first runtime. Declarative, versioned protocols; permission review at install and whenever a new capability is requested.
- Consequential actions (purchases, sensitive disclosures, destructive actions, identity changes) need fresh, scoped human approval.
- Connected agents authenticate ownership and have revocable, narrow scopes. Treat external agent messages, protocol definitions and public posts as untrusted data.
- Moderation basics before any public UGC launch: report, block, impersonation response, creator disputes, spam throttling, takedown path.
- SMS/RCS/WhatsApp/iMessage availability and economics are separate channel decisions; do not block social proof on messaging-provider approvals.

## 9. Out of scope until real validation
- A general feed with public DMs and algorithmic engagement loops; hundreds of seeded celebrity profiles.
- Paid creator marketplace/payouts, rankings driven by ads, subscriptions, premium protocol sales.
- Autonomous agent-to-agent delegated actions or uncontrolled reciprocal follows.
- Claimed universal compatibility with Instinct, dots, Grok Bot, Muse, or any proprietary assistant without published supported APIs and tested user authorization.
- Operating a PDS, ATProto Spaces as private memory, or end-to-end-encrypted blob hosting claims.
- Native mobile apps, arbitrary-code protocol marketplace, financial/medical external actions, broad integrations.

## 10. Measurable M0 exit criteria and alpha hypotheses
**M0 documentation ready for review** when product promise, three entities, launch loop, canonical data ownership, approval model, non-goals and decision conflicts are recorded, and the branch is available as a PR. **M0 is not fully accepted** until Chris approves the product pivot and reconciles conflicting accepted ADR/roadmap statements.

Hypotheses for later alpha (not current outcomes): recruit 5–10 original creators and 30–50 testers; >50% of onboarded testers start a protocol; >30% return in week four; zero unauthorized public disclosures or unapproved consequential agent actions. Track discovery → install → first successful run → voluntary share, segmented by protocol and creator. Measure execution costs and creator maintenance burden.

## 11. Owner decisions, Oct 13 2026

Recorded verbatim as product-owner answers (not evidence that related features, accounts, protocols, or policies are implemented, verified, endorsed, or approved for public launch):

1. **AT Protocol social login:** "at protocol login required"
2. **Product direction:** "Social-first"
3. **Following:** "Following is automatic but removable"
4. **Official protocols:** "Protocols are official once the person creates them or created by cleartify"
5. **Retention/deletion direction:** "start small" — **no retention durations, deletion windows, or vendor-specific terms were specified or approved**.

### PROPOSED minimal retention/deletion policy — awaiting owner sign-off

**Proposal only. Not approved; do not describe this as the live practice or a public privacy promise.** Retain only data needed to provide a requested private Clear or social service, ensure security, or comply with applicable obligations. Keep different categories separate: waitlist and activation contact data; marketing-consent and policy-acceptance audit records; OAuth sessions and security events; private Clear runs and messages; and public social data (which may persist on independently operated AT Protocol services). Avoid collecting unnecessary private inputs and minimize log contents. Offer authenticated account-data deletion requests, revoke sessions and credentials, and delete or anonymize eligible active records where technically and legally appropriate. Define separately how deletion is propagated to replicas, backups, logs, and external providers, including exceptions for lawful retention.

**Owner decisions still required before any public statement or launch:** approve actual retention periods per category; backup and replica expiry/deletion handling; verification and response process for deletion/access requests; vendor terms and roles; legal exceptions; data controller/privacy contact; and treatment of replicated public AT Protocol records. Until those are approved, **no fixed TTL, guaranteed erasure timing, or universal deletion promise is made**.

## 12. Next implementation order after M0 approval
Keep CI/lockfile and M1 waitlist on their existing branch; implement ATProto social identity + human/assistant profile models; publish one safely versioned public protocol; build one private Clear run; validate real creator-led installs; only then introduce BYO-agent execution and agent-to-agent collaboration.
