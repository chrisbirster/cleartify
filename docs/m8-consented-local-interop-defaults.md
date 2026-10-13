# M8 local consent and key-rotation defaults

**DEFAULT (Oct 13 2026), owner can change.** Internal engineering defaults, not public commitments.

- All functionality is **library-only**, not routed into the server; no new production behavior is enabled. No network transports, real PDS claims, agent execution, remote side effects, or outbound requests.
- A recipient owner authorizes a specific sending agent for at most 24 hours. Consent is independently revocable and does not substitute for the sender owner's scoped `message.preview` grant. Both are required to stage a local signed message.
- Agent signing-key rotation requires a domain-separated signature by the **currently active** key and atomically revokes it while registering the new Ed25519 public key. No private key storage. Existing pending records are retained for audit; no messages are delivered.
- Local two-agent test covers ownership setup, explicit grant and consent, signed envelope staging, replay rejection, revocation, stale signatures after rotation, and resumed staging under the rotated key.
- Owner/security follow-ups: real verified DID ceremony, authenticated routing, consent UI, full transactional authorization against concurrent revoke, persistent audit and key recovery, multi-device interop, transport and privacy review. Existing direct `ReplayRegistry.VerifyAndRecord` is an unconsented **low-level** method and must never be exposed to untrusted callers; use this gated facade for future integration.
