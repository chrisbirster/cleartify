# M2 OAuth security foundation — implementation constraints

Status: **foundation slice only**. This PR is not an operational AT Protocol OAuth login.

## Implemented
- Explicit, operator-configured public origin maps to exactly `/oauth/client-metadata.json` and `/oauth/callback` with HTTPS required except loopback development; arbitrary caller redirect URIs are never accepted.
- AES-256-GCM codec for encrypted OAuth request/session payloads using a 32-byte key supplied through `ATPROTO_SESSION_KEY_B64` by the deployment secret manager. Different nonce each seal; associated-data binding prevents swapping payloads between session IDs. No keys or credentials are committed.
- Durable Turso tables for OAuth requests (including state and browser nonce binding), OAuth client sessions, and browser session token digests. Tables remain empty until OAuth is wired; migration is additive and idempotent.
- Unit tests for URL restrictions, encryption tampering and migration.

## Still required before real login
1. Wire the official Indigo `oauth.ClientApp` with a `ClientAuthStore` implementation that encrypts OAuth request data (including PKCE verifier and DPoP key), persists it in Turso, binds browser nonce to OAuth state, consumes/revokes requests atomically and expires stale rows.
2. Expose GET client metadata, POST login and GET callback under exact configured endpoints. Use SDK's discovery/PAR, S256 PKCE, DPoP and issuer/state checks; DO NOT roll our own OAuth wire protocol.
3. Handle authenticated same-site browser session cookie with `Secure`, `HttpOnly`, `SameSite=Lax`, high-entropy token and short session expiry. Protect login initiation against CSRF and login-CSRF through state + browser nonce binding. Do not log tokens or callback query parameters.
4. Link stable DID only after successful SDK token exchange and separately authorized account linking. Never treat user-provided DID/handle or an unclaimed agent listing as ownership proof.
5. Test state mismatch, issuer substitution, duplicate callbacks, invalid redirect, token expiry, process restart, migrated identity/handle, session revoke, and failed Turso writes. Staging must test real hosted-PDS OAuth before promotion.
6. Confirm toolchain: current Indigo module `go.mod` requires Go 1.26, while Cleartify currently uses Go 1.24; decide whether to upgrade Go/Turso CI/Docker coherently or pin a reviewed compatible Indigo commit. This PR does not silently change runtime versions.

## Operator config for later integration
- `ATPROTO_PUBLIC_ORIGIN`: e.g. `https://api.cleartify.com`, exact allowlisted public origin
- `ATPROTO_SESSION_KEY_B64`: base64 of 32 secure random bytes, provided as a runtime secret
- `APP_ORIGIN`: existing Cleartify frontend origin. Post-login redirect must be a fixed path beneath this origin.

All private tokens/state remain in Turso, not a public PDS. Do not enable login endpoints until this checklist is implemented and tested.
