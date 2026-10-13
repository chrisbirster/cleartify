# M7 agent ownership ceremony defaults

**DEFAULT (Oct 13 2026), owner can change.** `AGENT_OWNERSHIP_ENABLED=false` by default; routes are absent unless enabled with ATProto OAuth. `POST /api/me/agents/{id}/ownership/start` and `/confirm` require the session-authenticated ATProto DID to match the registered agent owner. No request body can assert an alternate owner.

A cryptographically random challenge expires after five minutes. The hash, owner, agent and candidate public key are persisted. Confirm rechecks the current owner and consumes the challenge in a transaction alongside the binding and audit row. Replays, wrong owners, disabled agents, expired/invalid challenges, and incorrect proofs fail closed.

**Cryptographic proof:** The owner-controlled private Ed25519 key signs the canonical domain-separated ownership-v2 message containing agent ID, authenticated owner DID, candidate key bytes and single-use nonce. The server verifies using the stored candidate public key, commits the binding and audited proof, and rejects replay. Flag remains OFF by default. No real agent runtime or hosted adapter is activated. `HostAdapter` exposes Challenge and Health; `LocalTestAdapter` is a pure no-I/O implementation. No external hosting, key distribution or networking.

Owner/security signoff remains required for trusted real agent key enrollment and identity semantics, authenticated hosting endpoints, dedicated agent credentials, revocation/session freshness, account recovery, persisted remote synchronization failure handling, deployment and live PDS tests. Never interpret ed25519_verified audit as real-world ownership attestation.
