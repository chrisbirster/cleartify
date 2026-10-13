# M8 design — disabled agent-to-agent interoperability

This slice is **data and local signature verification only**; no API routes, network calls, remote agent discovery, automatic follows, or permissioned delegation are enabled. `AGENT_INTEROP_ENABLED` is reserved for a later implementation and must be treated as off until an audited adapter and authenticated agent ownership exist.

## Envelope and registry
An envelope names sender and recipient agent IDs, unique message ID, signing key ID, SHA-256 payload digest, issuance/expiry timestamps and Ed25519 signature. Canonical serialization signs a SHA-256 digest of the fixed-order fields. A registry stores public Ed25519 keys and revocation timestamps; an audit table records per-message outcomes. No private signing keys are held by these tables.

A future receiver must, **before** recording or executing any request: authenticate sender ownership, load an unrevoked key scoped to the sender, verify the signature and short TTL, atomically reject duplicate message IDs, check recipient acceptance and narrowly scoped capability grants, require explicit human approval for consequential actions, sanitize untrusted message content, limit traffic and record an audit trail. A valid signature is not authorization. Following another agent never creates delegated authority.

## Next gates
- Explicit interoperability protocol and transport selection (A2A compatibility evaluation).
- Public key rotation/revocation, anti-replay transaction semantics and durable identity proof.
- Sandbox and prompt-injection threat model; private memory cannot become public by default.
- Real adapter contract tests against two separately owned test agents.
- No interop network endpoint until all security gates are reviewed.
