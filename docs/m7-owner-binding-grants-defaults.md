# M7 local agent owner binding and scoped grants

**DEFAULT (Oct 13 2026), owner can change.** Internal engineering policy only; not a privacy promise or proof of actual PDS verification.

- A person DID from an **already authenticated and independently verified server session** must be supplied by a trusted caller; no public routes are wired and this library alone does not check a remote PDS. Existing claimed person-profile lookup is an additional prerequisite, **not** verification.
- Ownership proof is an Ed25519 signature over a domain-separated agent-ID/person-DID message, confirming possession of a locally registered agent signing key. No cross-owner bind or key swap; owner unbind revokes access. Cryptographic possession does not establish real-world identity or remotely attest an agent.
- Only existing, locally marked `verified` agent registrations may grant or receive permissions. Only `message.preview`, `protocol.preview`, and `local.read` are supported; grants expire in no more than 24 hours and can be revoked immediately. Revoked bindings or disabled agent records fail closed. No default grants.
- This PR adds persistence and unit tests but **does not** introduce API routes, actual agent messaging, external agents, real PDS verification, consent-screen UI, or production behavior. Before runtime wiring, authenticate and authorize the owner at every call, formalize independent DID proof ceremony, bind registered signing keys to the verified principal, address concurrent revocation races and persistent audit, and complete owner security review.
