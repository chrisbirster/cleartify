# Round 6 — identity/session path for two people (engineering proposal)

Status: **analysis and proposed implementation plan only**. No live login, provider verification, or product acceptance is claimed. This document is not an authorization to deploy.

## Verified wiring in `dev`

- `POST /api/beta/verify` calls `beta.Verify`, which creates an active `users` row and verified `phone_identities` row. When ClearSessions are configured, the API calls `clearsession.Service.Issue` and sets an HttpOnly `clear_session` cookie with `Path=/api/me/`, SameSite=Strict, and Secure when app origin is HTTPS. The response does not return a session token in JSON.
- The private Clear cookie is resolved server-side by `ClearSessionPrincipal`. Its 30-day idle lifetime is extended by authenticated use and can be revoked. `privateAlpha` additionally checks alpha eligibility when enabled.
- Bluesky/Indigo OAuth uses its own social session/DID and enables creator/social install request routes. It **does not issue a Clear session** in this code path.
- Identity linking explicitly requires *both* the recent Clear session (fresh within 10 minutes) and authenticated social DID, then a one-use challenge and explicit consent. A social DID alone never authorizes private Clear data.
- `POST /api/me/protocol-install-requests` is currently social-OAuth-bound and the installation decision uses an admin route. This cannot serve as ordinary user self-service without a new authenticated owner/consumer path.
- The existing `/api/beta/verify` is a one-time activation; repeating it on an activated invite returns an error. **No returning-user Clear sign-in has been demonstrated**. Treat a second browser/restart and a second user as unverified until covered by tests.
- Debug/Disabled senders are the only currently wired notification implementations; debug challenges are not a production sign-in channel.

## Recommended user A / user B flow

1. Admin invites A and B separately through existing beta workflow; each completes their *own* email invite + SMS challenge. Local test harness may use the Debug sender to capture challenges. Never reuse invite tokens, share cookies, or accept a caller-supplied `userId` as authority.
2. Each activation obtains a distinct private Clear session. Test that A's cookie cannot access B's private records and vice versa.
3. For creator/social features, each person independently connects Bluesky with OAuth. If the feature needs the social DID to represent a private Clear account, require the existing explicit identity-link challenge after reauthentication. Do not infer identity from a claimed handle or shared email.
4. A creates and publishes a **reviewed, visible, non-taken-down** text-only recipe under their creator identity.
5. B browses published recipes and opts to install one using B's *private Clear session*. An authenticated B explicitly consents to the immutable recipe version/digest and declared permissions; success creates only B's installation and follow-ups.
6. B talks to Clear about the installed recipe; Clear's follow-up state is always scoped to B. A sees published recipe metadata and aggregate non-private information only, never B's conversation or follow-up details.

## Smallest auth implementation required before this is a working journey

- Add an **OFF-by-default** returning-user reauthentication flow: verified phone challenge for an existing active account, not a second beta activation. Use the current `notifications.Sender` and digest-only expiring one-use challenge persistence; limit attempts, resend and issuance, fail closed if delivery is unavailable. Do not disclose whether a phone belongs to an account in public responses.
- Issue/revoke `clear_session` only after successful verification, with the existing cookie security attributes. Explicitly test logout, replay, expiry, cross-user isolation, and restart persistence. Avoid putting credentials in URL parameters or responses.
- Where the SPA and API are on separate origins, verify same-site/cookie behavior and explicit allowed-origin credentialed requests in local tests; current CORS headers alone do not establish a cross-site cookie login. Prefer same-origin API routing where possible rather than weakening SameSite.
- Document/implement a separate optional Bluesky connect flow; **never auto-link** a social DID to a private user merely because OAuth succeeded. A renewed phone challenge is required for sensitive link operations.
- Do not silently bypass the current alpha eligibility check to ease testing.

## Proposal: self-serve published recipe consent

- Add an OFF-by-default owner-scoped `POST /api/me/protocol-installs` authenticated by the private Clear session, using a server-selected, current, publicly visible publication/version and digest. Server checks availability, takedown/moderation holds, access rights and permission policy atomically.
- Allow B to accept purely **declarative text guide steps** and harmless explicitly whitelisted local operations without an admin. Persist affirmative per-install consent and an audit record; repeated requests are idempotent.
- Keep an approval path for permission increases, untrusted or elevated capabilities, paid operations (fake only), and moderation hold/review. No creator-provided executable content or code runner.
- Existing admin install/run decisions remain available for guarded cases; they are **not** a mandatory gate for ordinary published guide installs.
- Never treat A's ownership of the recipe as permission to act on B's behalf.

## Tests needed for implementation PRs

- Two separate invite/SMS activations yield distinct sessions and scoped private access.
- Returning-user challenge succeeds once; wrong/replayed/expired/locked challenge fails; auth survives DB restart; revoked session remains revoked.
- Social OAuth without Clear session cannot read private data; Clear session without social OAuth cannot claim A's creator identity; linking requires independently authenticated principals and consent.
- B can install only reviewed, visible, permission-allowed content; A's drafts, hidden/taken-down versions and executable fields are rejected.
- Authn/permission state is checked again before follow-up creation, not only when showing a discovery card.

## Current blockers and non-goals

Returning-user authentication, a complete second-user browser flow, self-service install, text guide schema, durable follow-up worker, conversation API/UI, and a two-person local E2E are **not yet built or verified by this document**. No production SMS, public OAuth/PDS verification, DNS, deployments, payment provider, secrets or legal commitments are in scope. No milestone percentages are asserted.
