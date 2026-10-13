# Private Clear sessions — engineering defaults

**DEFAULT (Oct 13 2026), owner can change.** Internal behavior only; not final legal wording or a public privacy promise.

- `CLEAR_SESSIONS_ENABLED` defaults **false**. Existing beta verify behavior does not change unless enabled.
- On successful one-time invite + phone challenge activation, create a random 256-bit bearer cookie scoped to `/api/me/`, with an HMAC digest at rest, HttpOnly and SameSite=Strict. Secure applies under HTTPS-origin configuration.
- Rolling idle expiry: **30 days**; token revocation on logout and revocation-by-user supported. This is not a durable re-authentication mechanism for sensitive linking: a fresh Clear challenge and a separate verified AT Protocol OAuth session remain necessary.
- GET `/api/me/clear-session` validates the active user and session. POST `/api/me/clear-session/logout` revokes it. Neither endpoint confers alpha membership or admin privilege.
- Remaining boundaries: full reauthentication for already activated users, absolute maximum lifetime, account deletion orchestration, real hosted-PDS and staged validation are not delivered here. Do not enable linking or deployment on this basis.
