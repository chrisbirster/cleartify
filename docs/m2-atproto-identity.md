# M2 AT Protocol identity — implementation boundary

This is **M2 Slice 1: identity value object**, not OAuth login. The Go package validates untrusted identifier syntax and establishes the invariant that a DID, never a handle, is the durable public-account key. It performs **no** PDS network access, identity ownership verification, token issuance, login, account linking, or user creation.

## OAuth implementation required next

1. Choose and verify a maintained, security-reviewed OAuth client compatible with the Go server and AT Protocol's current profile (including metadata publishing, PKCE, PAR/DPoP as required). Do not invent a generic OAuth flow.
2. Implement discover/authorize/callback with exact redirect allowlist, CSRF/state, nonce and replay protections, session binding and token/key lifecycle.
3. Resolve DID documents/handle correspondence through the actual AT Protocol identity mechanisms; refuse DID mismatch after login.
4. Persist internal-user-ID to DID binding privately in Turso, explicitly approved by the authenticated user. Decide whether ATProto is mandatory for social-only signup.
5. Support revocation, expiry, handle changes, PDS migration, logout, and audit logs without leaking credentials.
6. Exercise integration tests with a supported test PDS/OAuth provider, plus failure tests for callback substitution, stale token, handle change and mismatched DID.
7. Never publish private Clear run data, personal memories or approvals to an ATProto PDS.

`ParseIdentity` is **not** an account-verification step. Do not expose an endpoint that accepts arbitrary DID+handle as a login.
