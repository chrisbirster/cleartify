# M2 — Bluesky real-OAuth verification plan (not executed)

**STATUS: DOCUMENTATION AND CONFIG CHECKLIST ONLY — NO LIVE TEST, NO DEPLOY AUTHORIZATION.** Internal engineering verification plan, October 13, 2026. This is not public policy or legal wording.

## Explicit blockers and boundaries

A **public HTTPS origin** (isolated staging hostname or carefully secured HTTPS tunnel) and a **dedicated throwaway Bluesky test account** are required. **Neither has been provided yet. Deployment has not been granted.** No one should interpret this document as authorization to create a tunnel, change DNS, deploy, provision secrets, connect to the real Bluesky PDS, or turn on production OAuth. Every step below is for a later owner-approved test window.

The ordinary repo's local tests cannot replace a hosted-PDS verification: the real test requires a browser, reachable HTTPS metadata and callback, Bluesky authorization, and an operator-controlled isolated service. Keep real accounts and data entirely out of local fixtures and CI.

## Prerequisites: operator checklist

1. Owner explicitly approves the test environment, responsible tester, timing, selected test account and its cleanup. Use a separate disposable Bluesky account; no personal/main account.
2. Obtain a publicly reachable HTTPS API origin **without a path, query or fragment** for `ATPROTO_PUBLIC_ORIGIN` (e.g. `https://oauth-test.example.net`). Current `ConfigureOAuthURLs` rejects non-HTTPS remote URLs and remote origins with explicit ports. The origin must serve HTTPS with a trusted certificate. Localhost/127.0.0.1 HTTP is **development-only**, not a valid substitute for this Bluesky test.
3. Configure `APP_ORIGIN` as the approved public HTTPS **frontend origin** (also scheme+host, no path/query/fragment); it may differ from the API origin. Verify browser access and the allowed cross-origin/Origin relationship in the isolated environment.
4. Set `ATPROTO_OAUTH_ENABLED=true` **only in the approved test environment**, and supply `ATPROTO_SESSION_KEY_B64` via its secret manager (base64-encoded 32-byte key). Reference `APP_SECRET`, `TURSO_REMOTE_URL`, and `TURSO_AUTH_TOKEN` by name only **if the selected isolated environment requires them**. Use test-isolated datastore/config; never reuse a production secret or shared production user store. No literal tokens or keys in files, docs or logs.
5. Confirm no unrelated feature flags are inadvertently enabled: `IDENTITY_LINKING_ENABLED`, `IDENTITY_RECOVERY_ENABLED`, `PRIVATE_ALPHA_ROUTES_ENABLED`, and payment/runner/publishing flags remain OFF unless a distinct owner-approved linked-account test explicitly requires each one.
6. Review the pinned Indigo OAuth dependency and network egress expectations in the approved environment; do not change `go.mod`, runtime or deploy configuration as part of this documentation PR.

## OAuth client metadata and endpoint acceptance

The current handler derives the exact URLs from `ATPROTO_PUBLIC_ORIGIN`:

- **Client ID and public metadata URL:** `https://<approved-api-origin>/oauth/client-metadata.json` (the full origin plus path; replace only after environment approval).
- **Redirect URI:** `https://<approved-api-origin>/oauth/callback`.
- **Login entry:** `GET /oauth/login` (renders identifier form); `POST /oauth/login` (requires matching configured Origin header).
- **Logout:** `POST /oauth/logout` (requires matching Origin).

The metadata handler uses Indigo's `oauth.NewPublicConfig(clientID, callback, ["atproto"])` / `ClientMetadata()`, sets client name `Cleartify` and `client_uri` to the configured public origin, and validates it before JSON response. When the test environment exists, inspect the **actual published JSON** rather than guessing provider-specific details: client_id must be the exact externally fetchable HTTPS metadata URL; redirect_uris must include the exact callback; request/response authorization and PKCE/DPoP/PAR capabilities must align with the pinned Indigo SDK and issuer requirements; metadata must have an appropriate JSON Content-Type, be anonymously reachable and not redirect to login, an internal hostname or HTTP. Do not advertise wildcard redirect URLs or client secrets.

The browser callback must use the same external host and route behind any reverse proxy; the configured origin must reflect what Bluesky can fetch. Ensure HTTPS routing does not change method, query parameters or cookie behavior. Local callback state/nonce handling must persist across the short redirect flow; sessions are local server records, **not** Clear account links.

## Manual operator-run end-to-end procedure (only after approval)

| Step | Action | Expected result / evidence |
|---|---|---|
| 1 | On the isolated HTTPS environment, fetch `/oauth/client-metadata.json` from outside the host using a standard browser | HTTP 200 JSON; HTTPS origin and exact client_id/redirect_uris, no internal URL or secrets |
| 2 | Open `GET /oauth/login` in a clean browser profile; use the throwaway account handle | Identifier form loads; no previously authenticated browser session |
| 3 | Submit the form from the configured public API origin, following browser-generated cookies/Origin semantics | HTTP 303 to an HTTPS authorization URL; a short-lived `cleartify_atproto_flow` cookie is set; malformed identifiers and Origin mismatch are rejected |
| 4 | Authenticate and approve at Bluesky using **only** the throwaway account | Provider displays expected app/client identity and requested scopes, never an unrecognized redirect |
| 5 | Follow the provider redirect to `/oauth/callback` | Indigo verifies state, issuer, PKCE S256, DPoP and token exchange; no error/stack trace or credential leak |
| 6 | Follow application redirect to `APP_ORIGIN/?atproto=connected` | HTTP 303; opaque `cleartify_atproto_session` cookie (Secure, HttpOnly, SameSite=Lax) scoped to the API host; throwaway DID exists in local OAuth session rows; no automatic Clear account linking |
| 7 | From the same browser, exercise an authenticated social-only route available in the isolated configuration | Server resolves verified DID from the session, not a request-supplied DID/handle; unauthenticated second browser cannot impersonate the user |
| 8 | Repeat after a controlled process restart using the same isolated datastore and session key | Browser session persists within its configured TTL; encrypted OAuth/session rows are readable; no token/DPoP key appears in logs |
| 9 | Run negative checks in an isolated browser without modifying real Bluesky/provider state | Wrong/missing flow cookie, expired/replayed state, modified callback query, wrong Origin, arbitrary post-login redirect, and duplicate callback cannot create a valid session; record only statuses and redacted errors |
| 10 | POST `/oauth/logout` from valid Origin; retry authenticated route | Local browser session cleared and rejected; Indigo attempts remote revocation when supported; failed/unsupported remote revocation must be separately recorded, not assumed successful |
| 11 | Repeat optional separate account-link ceremony only after new explicit authorization and required flags | Verified Clear session and independent social reauthentication required; consent/challenge binds accounts; cross-user link and replay denied; unlink/recovery must not publish private Clear data |

The negative callback checks should use test-generated state and carefully controlled browser replay; never paste authorization codes, tokens, or DPoP keys into tickets or logs. Do not assume remote logout always revokes the provider session.

## Evidence and go/no-go

Capture time, tested commit SHA, intended origins, SDK version, HTTP status/cookie *attributes* (not values), redacted test DID, expected-vs-actual per step, reviewer, and clean-up confirmation. **Stop immediately** on metadata mismatch, callback confusion, invalid cookie attributes, origin/redirect mismatch, failed local session revocation, unverified DID assignment, unexplained exposed credentials, or inability to isolate the test store. A green local CI run is not proof of real Bluesky interoperability.

## Rollback and cleanup

1. Stop the isolated test, disable `ATPROTO_OAUTH_ENABLED` and any optional M2 flags **within the isolated environment only**. Do not alter production configuration.
2. Remove test browser sessions/cookies. Invalidate and delete disposable OAuth session/request/browser records in the isolated datastore under approved procedures; retain redacted evidence only. If remote revocation is unsupported, disconnect the throwaway account through available provider controls and document uncertainty.
3. Revoke/delete only isolated environment credentials through its secret manager, and close/remove any owner-approved tunnel, routing or staging resource using that environment's procedure. Do not delete any shared or production infrastructure.
4. Verify `/oauth/*` routes are no longer mounted when the feature is disabled, review logs for accidental token exposure, and record the outcome. Any release or production enablement requires separate owner approval.

**Nothing in this plan provisions public HTTPS, supplies a throwaway account, grants deployment, or authorizes live-network tests.**
