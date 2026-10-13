# ADR-0008: Social Marketplace Realignment (proposed)

- **Status:** Proposed — awaiting owner decision
- **Date:** 2026-10-12
- **Supersedes:** None; ADR-0007 remains accepted until an explicit replacement/amendment.

## Context
Cleartify's existing M0/M1 documentation leads with a phone-first private assistant, with social discovery as an optional later plane. Current product exploration makes the humans + agents + portable protocols marketplace a stronger primary differentiator. The implementation and API foundations are reusable, but the onboarding order and launch success metric need deliberate reconciliation.

## Proposed decision
1. Market Cleartify as a social marketplace where people, assistants and protocols are discoverable. Clear remains the default runnable assistant.
2. Retain Go, Turso, Solid and the existing landing page/first-friend implementation as the starting point; no rewrite.
3. Use ATProto for **consented public, portable** identities and protocol publication where feasible. Use the internal user ID and private Turso data for assistant memory, approvals and runs.
4. Start with **human profile -> public protocol -> private run -> voluntary sharing** as the smallest product loop.
5. Model agent profiles as **directory only** until ownership is verified and a real supported adapter is connected. Allow public follows, not implicit authority.
6. Keep creator provenance, version pinning, safe runner permissions and moderation as prerequisites for broad publishing.

## Alternatives
- **Phone-first only (current ADR-0007):** fastest private messaging launch; social network and creator loop delayed.
- **ATProto for every signup:** portable social onboarding; adds provider/PDS and OAuth/recovery risk, and may undermine text-without-account onboarding.
- **Optional ATProto public identity:** recommended bridge while piloting; avoids forcing every text customer into public social infrastructure.

## Unresolved owner decision
Should social alpha require a DID or allow conventional/internal signup with optional later linking? This ADR recommends the latter until hosted PDS onboarding and recovery are validated, but does not override ADR-0007.

## Consequences
A public AppView/search index, separate creator verification, privacy/moderation and protocol-version storage are necessary. ATProto does not make private memory encrypted, proprietary assistants interoperable, or application hosting free. Social profile/follow semantics must be specified and tested rather than assuming Bluesky conventions cover all Cleartify interactions.

## Acceptance
The linked M0 product contract becomes the primary scope reference once approved. M1 delivery remains separately tracked.
