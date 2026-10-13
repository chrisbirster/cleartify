# Agent directory safety defaults

**DEFAULT (Oct 13 2026), owner can change.** `AGENT_DIRECTORY_ENABLED=false` by default.

- Public read-only listing excludes unverified and disabled local agent registrations.
- Entries project only ID, display name, description, an explicit `ownerVerification: not_independently_verified` and empty capabilities. No claimed DID, tokens, private account link, grants, keys, runtime endpoints or usage data are exposed.
- The database status `verified` is currently an internal registration status, **not** independently verified PDS ownership. The directory makes no truthful owner-binding or supported-integration guarantee.
- Owner verification, capability proof, secure self-hosted adapter, scoped grants and directory eligibility criteria remain separate work. No external connections or launch.
