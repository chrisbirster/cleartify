# M2 — Indigo OAuth login (feature flag OFF by default)

- `ATPROTO_OAUTH_ENABLED=true` is required to mount any of the `/oauth/*` handlers.
- `ATPROTO_PUBLIC_ORIGIN` is the fixed HTTPS API origin (e.g. `https://api.cleartify.com`). Only the SDK's configured `/oauth/callback` redirect is advertised; no caller-specified post-login redirect is accepted.
- `ATPROTO_SESSION_KEY_B64` is a base64-encoded 32-byte random key from the hosting secret manager. Never check it in.
- `APP_ORIGIN` is the fixed HTTPS frontend origin, used only after a verified callback.
- Uses Indigo OAuth `StartAuthFlow` / `ProcessCallback` / `Logout`. SDK performs PKCE S256, DPoP and PAR, DID/issuer checks.
- Browser flow cookie and state are joined using a SHA-256 nonce digest and a one-time, expiring Turso row. Callback consumes the row atomically before token exchange.
- Tokens, PKCE verifiers and DPoP private keys are encrypted with AES-256-GCM in Turso. The browser cookie has only a random opaque value; its SHA-256 digest is in Turso. Cookie is `HttpOnly`, `SameSite=Lax`, host-only and `Secure` on HTTPS.
- Logout uses Indigo remote token revocation when supported, deletes the local OAuth session, invalidates the browser session and clears the cookie.
- **Do not enable in production** until tested against a real PDS, including metadata reachability, handle login, callback issuer mismatch, logout/revoke, session restart, migration and key rotation. No credentials or provider secrets are committed.
- This implements initial login and independent AT Protocol session identity only. Linking to an existing phone-first Cleartify user is a separately approved and authenticated step; this flow does not silently merge accounts.

Temporary build workflow may be needed to pin and resolve Indigo dependencies for Go 1.26 (generated go.sum rather than synthetic entries). The SDK's security properties are inherited from the pinned upstream revision and warrant dependency review before production enablement.
