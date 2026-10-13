# M3 local moderation engineering defaults

**DEFAULT (Oct 13 2026), owner can change.** No production behavior or external AT Protocol action is authorized.

- `SOCIAL_MODERATION_ENABLED=false`, additionally requires verified social OAuth enabled and a live server-side social DID session.
- Authenticated local social users can block/unblock a registered subject DID and submit a report with one of `spam`, `harassment`, `impersonation`, or `other`.
- Reports remain local, private and `pending`; no automated takedown, public accusation, handle ownership verification, AT Protocol label emission, PDS write, appeal decision or moderation staff UI.
- Blocks are local user-scoped data, not PDS graph updates. Existing read/feed behavior is unchanged until separate policy-approved block filtering is implemented.
- Route origin, CSRF and bounded mutation throttling reuse existing middleware. The authenticated actor DID is never request-supplied.
- Owner decisions still needed: actionable report workflow, appeal/takedown notices, report retention, authorized moderators and impersonation evidentiary standard. Public handle verified badges must not be inferred from local DID matching.
