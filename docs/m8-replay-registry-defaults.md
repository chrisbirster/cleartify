# M8 offline agent replay registry

**DEFAULT (Oct 13 2026), owner can change.** No HTTP transport, PDS integration, agent execution or external network calls.

- Each signed envelope is checked against a locally registered non-revoked Ed25519 key whose recorded agent ID matches the sender, then checked for signature, issuance, expiration and clock bounds.
- A unique message ID is inserted with status `pending`. Retries/duplicate pushes cannot be accepted twice. Recorded does **not** mean authorized, accepted, delivered or executed.
- Tampered payload metadata, wrong key owner, revoked keys and expired envelopes are rejected.
- Actual authenticated transport remains **disabled/nonexistent** pending scoped revocable consent grants, recipient validation, key ownership proof and isolated two-agent integration. No agent-to-agent message is sent.
- Owner chooses permitted message classes, clock tolerance, retention, recovery and user-visible consent; no final policy is implied.
