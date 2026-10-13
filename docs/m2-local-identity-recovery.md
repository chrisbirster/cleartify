# M2 local identity disconnect and recovery (engineering defaults)

**Default OFF:** `IDENTITY_RECOVERY_ENABLED=false`. The routes require the existing `IDENTITY_LINKING_ENABLED`, Clear sessions, social OAuth setup and private-alpha protection when the alpha flag is on. These are *local workflows exercised with fake social identity*; this document is not final privacy/legal wording.

New routes:
- `POST /api/me/identity-link/disconnect` with `{"confirm":true}`: requires a freshly issued authenticated Clear session (10-minute limit), an authenticated social session whose DID matches the active link, and the configured Origin.
- `POST /api/me/identity-link/recover` with `{"confirm":true}`: requires a freshly issued authenticated Clear session and Origin, **not** a social session. This models account recovery after a lost/disconnected social login, **not** recovery from loss of Clear authentication. No identity assertion is taken from request JSON or query parameters.

Each route atomically revokes the active DID link, cancels pending link challenges, revokes **every** active Clear session for that user and appends an unlink audit record. It never revokes another user's sessions. Local-session cookies are cleared and clients must freshly authenticate before relinking. Repeating the operation without an active link fails closed (409). If the local DB commit succeeds but the remote database push fails, the endpoint returns unavailable; do not treat that as rollback or guarantee synchronization.

Local E2E tests exercise challenge/consent/confirm, replay prevention from existing tests, cross-account disconnect denial, dual-session revocation, another user unaffected, a new authenticated relink, recovery without social reauthentication and flag-off 404. There are no real-PDS calls, production credential recovery, email/phone identity replacement, deployed cross-device tests, or new external endpoints or keys. Live deployment and legal policy remain owner-only gates.
