# M7 agent ownership ceremony defaults

**DEFAULT (Oct 13 2026), owner can change.** `AGENT_OWNERSHIP_ENABLED=false` by default; routes are absent unless enabled with ATProto OAuth. `POST /api/me/agents/{id}/ownership/start` and `/confirm` require the session-authenticated ATProto DID to match the registered agent owner. No request body can assert an alternate owner.

A cryptographically random challenge expires after five minutes. The hash, owner, agent and candidate public key are persisted. Confirm rechecks the current owner and consumes the challenge in a transaction alongside the binding and audit row. Replays, wrong owners, disabled agents, expired/invalid challenges, and incorrect proofs fail closed.

**Critical limitation:** `LocalFakeVerifier` is a deterministic SHA-256 exercise, **not** cryptographic proof of private-key possession. Flag must remain OFF outside controlled development. No real agent runtime or hosted adapter is activated. `HostAdapter` exposes Challenge and Health; `LocalTestAdapter` is a pure no-I/O implementation. No external hosting, key distribution or networking.

Owner/security signoff remains required for real signed proof (Ed25519), authenticated hosting endpoints, dedicated agent credentials, revocation/session freshness, account recovery, persisted remote synchronization failure handling, deployment and live PDS tests. Never interpret fake_verified audit as real-world ownership attestation.
