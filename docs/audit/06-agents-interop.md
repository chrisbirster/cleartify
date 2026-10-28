# Audit checkpoint 6: agents and local interoperability

Revision: `12390da9504ae91821851d027556968268fa0290`. October 27, 2026. All first-party agent/interop runtime and test files were read. This is not independent verification of proprietary chatbot compatibility or third-party cryptography.

## Actual capability

Administrator-mediated registration requires a claimed public owner and creates an unverified agent. An OAuth-authenticated owner can prove Ed25519 key possession through a single-use, contextual challenge. Separate grant/consent/key registries support signed message envelopes, replay rejection, and a manually invoked SQL outbox/inbox worker. The worker delivers a digest/envelope record within the same authoritative database; it does not deliver protocol bodies to another assistant.

There is no public external-agent delivery endpoint, verified Instinct/Grok/Meta adapter, protocol handoff UI, or production worker loop. The executable discards its local transport object (F04). Agent grants and reconciliation are primarily library/test capabilities. Do not make them launch dependencies for readable protocol sharing.

## Reproduced findings

### F25 — P2/integration gap: ownership proof does not finish agent activation

Ownership.Confirm (`internal/agents/ownership_ceremony.go`) stores a binding and proof audit but never changes agent_registrations.status. Directory and grants require status=verified. The ownership API returns verified=true (and the stale field fakeVerifierOnly=true), despite the executable using the real Ed25519 verifier.

Probe: registered an agent for a claimed owner; generated a key, completed Start/Confirm with a valid signature, then queried its registration/directory. Status remained unverified and the directory was empty. Existing grant/interop tests manually update status to verified through SQL.

Key-possession verification and directory eligibility may deliberately be separate decisions. The missing piece is an explicit reviewed promotion mechanism and accurate response/status language, not automatically promoting every proved key. No production promotion endpoint was found.

### F26 — P2/library-only: duplicate valid grants deny message transport

GrantService.Grant inserts distinct rows for repeated same-scope grants, and Allows accepts count>0. LocalTransport.Send, AuthenticatedWorker.Submit, and Outbox.authorized require exactly one matching grant, while DurableInbox uses EXISTS. Two live grants can make these predicates disagree.

Probe: adapted the existing local transport fixture to issue the same sender/recipient/scope grant twice. GrantService.Allows returned true while LocalTransport.Send returned ErrTransportDenied despite valid recipient consent and signature.

Fix direction: use EXISTS or equivalent semantics for authorization unless a uniqueness invariant is actually enforced; define renewal/replacement semantics and test duplicates, overlapping TTLs, and revocation. Do not weaken independent owner/consent checks.

## Supported controls and limitations

- Ownership proof binds agent, owner, candidate key, and nonce; wrong actor/signature and replay tests pass.
- Grant tests cover narrow scopes, maximum TTL, expiry, sender/recipient unbinding and owner mismatch.
- Transport validates signatures, recipient consent, registered keys, owner bindings, replay, and database transaction consistency.
- Worker atomically records/enqueues and requires an explicitly trusted in-process caller. The caller's Verified boolean is not a network authentication protocol.
- Inbox tests cover crash-after-commit idempotency and authorization checks after queueing. Leases fence acknowledgments, not arbitrary external effects.
- Rotation proves continuity from an old registered message key. This does not establish PDS identity, ownership of a proprietary assistant account, or security of a remote runtime.
- The older consented/replay library separates authorization checks across statements. Prefer the transactional transport/worker path when integrating, and independently review race/revocation boundaries.
- Ownership challenge confirmation can replace an existing binding via upsert, whereas legacy Bind rejects rebinding. Define key rotation/recovery and downstream grant invalidation consistently before exposure.
- Some grant/consent comparisons use variable-width RFC3339Nano strings while newer outbox checks use date-aware comparisons. Boundary consistency remains a targeted regression task.

No network interop or live provider connection was performed. Existing agent/interop tests passed in the complete uncached suite; the duplicate-grant probe was removed after reproduction.
