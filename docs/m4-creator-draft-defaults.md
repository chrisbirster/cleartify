# M4 authenticated creator draft defaults

**DEFAULT (Oct 13 2026), owner can change.** `CREATOR_DRAFTS_ENABLED=false` and also requires flag-gated social OAuth.

- Draft ownership is bound to the verified OAuth session's DID, never a request-supplied creator DID or handle.
- Immutable `(creator_did,slug,version)` unique versions; valid constrained slug, title, license metadata and JSON-object spec are required.
- This endpoint creates status=`draft` only; it never publishes, endorses, installs or executes a protocol, and existing discovery only reads status=`published`.
- A draft is not proof of copyright or a final approved content license. This implementation does not enable public creator publishing, remix, official endorsement or takedown review.
- Future owner approvals: license/attribution categories, verified handle ownership, publication review, permissions policy, consent and staged real-PDS verification. No deployment or provider call.
