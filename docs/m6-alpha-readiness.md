# M6 private-alpha release checklist

**Status: staging preparation, not release approval.** M6's private alpha is available only to a verified phone-first beta-activated user who has an unrevoked `alpha_memberships` record. The database guard requires an actual `beta_invites.status='activated'` row attached to an active `users` record. No user-supplied email, handle or DID is sufficient proof.

## Required alpha gate integration
- [ ] Keep invitation issuance behind existing `ADMIN_TOKEN` authorization.
- [ ] Complete the actual email invite → phone verification → activated user transaction.
- [ ] Grant alpha membership only after server-side `alpha.Gate.Admit` validates the activated invite.
- [ ] Enforce `alpha.Gate.Eligible` on **every future private-alpha route** after authenticated identity resolution; default deny if no valid session or membership.
- [ ] Verify revoked users and incomplete activation cannot access alpha functionality.
- [ ] Ensure M5's administrative run-request routes stay admin-only and cannot execute tools; do not expose them to unverified beta invitees.
- [ ] Do not permit public OAuth identity alone to grant alpha privileges. Account linking is a separate verified operation.
- [ ] Verify CSP, session expiry, logging, email/SMS limits, PII retention and production legal notices before a real invitation.
- [ ] Confirm live hosted-PDS OAuth separately if enabling optional social login; OAuth is off by default.
- [ ] Recruit 30–50 consenting testers and original protocol creators; track opt-in product metrics without disclosing private runs.

**Important:** This slice provides the database authorization predicate and revocable membership service. The current application has no complete authenticated end-user session for alpha feature routes, so those routes must remain unavailable until explicit identity/session middleware is implemented and verified. The checklist is not a claim that a private-alpha launch gate already covers every hypothetical route.

## Staging smoke plan

Run in a distinct staging environment, never production. Use disposable email and phone numbers under your control, a staging-only Turso database, staging API/SPA origins, and staging secrets injected through the deployment manager. Do not commit phone verification codes or session keys.

1. Deploy the exact PR head to staging with a verified fresh database. Confirm `GET /healthz` 200 and migrations idempotently create the alpha membership table.
2. Submit valid waitlist signup and verify one normalized entry; duplicate signup should not create a second user.
3. Without an invitation, `Gate.Admit` must return `ErrNotActivated`; incomplete invite and wrong phone code must not grant membership.
4. With an authenticated admin, issue one invite through existing admin endpoint, follow invite link, start activation, and finish phone verification using approved staging transport.
5. Verify the exact activated beta invite references the new active user; `Gate.Admit` then succeeds. Verify `Gate.Eligible` true for that user and false for another user.
6. Revoke membership and prove `Gate.Eligible` returns false immediately, including after restart / Turso sync.
7. Confirm public `GET /api/protocols` and public profiles expose no private installation, session, phone or run information.
8. Confirm every unpublished alpha route is absent or refuses unauthenticated requests; M5 administrative endpoints must return 401 without a valid admin bearer.
9. Exercise negative tests for replayed invite, expired token, wrong SMS code, expired activation, deleted session, malformed request and abusive signup rate.
10. Record specific staging URL, deployed image digest, CI run URL, request status codes, operator, date, rollback digest and any unresolved risks. **Do not publish credentials, real personal information, invite links or raw session tokens.**

### Deployment go/no-go
Go only when both `ci/go` and `ci/web` are green on the merge head, staging smoke is recorded, and the identity-bound alpha route middleware plus legal notice/retention review are complete. Otherwise leave the beta private and behind the existing activation controls.


## October 13 flag inventory — staging planning only

All flags listed below are **environment controls**, not instructions to change production configuration here. OAuth and private-alpha social features must remain disabled until owner verification and release sign-off.

| Flag | Default | Readiness and required checks |
| --- | --- | --- |
| `ATPROTO_OAUTH_ENABLED` | false | Indigo OAuth integration; **block enablement** until live hosted PDS auth, callback, refresh, logout/revocation and token-storage tests pass |
| `SOCIAL_ACCOUNT_REGISTRATION_ENABLED` | false | Requires OAuth; confirms unauthenticated requests cannot create social accounts; does not link a phone-first Clear account |
| `SOCIAL_DEFAULT_FOLLOW_ENABLED` | false | Requires OAuth; local /@chris automatic follow, persistent user opt-out and no PDS follow write |
| `SOCIAL_PROFILE_ACTIONS_ENABLED` | false | Requires OAuth; DID-bound local claims/follows; separately validate handle ownership against real PDS before public use |
| `PROTOCOL_INSTALL_FLOW_ENABLED` | false | Proposed by PR #29, **do not assume present in dev until merged**. Requires OAuth; preview/acknowledge/version-pin/admin approval and never execute |
| `AGENT_REGISTRATION_ENABLED` | false | Admin-only local, unverified agent listing; no external agent identity or network integration |

**Prerequisites, not default flags:** exact staging `APP_ORIGIN` and `ATPROTO_PUBLIC_ORIGIN`, verified `ATPROTO_SESSION_KEY_B64`, `APP_SECRET`, admin bearer and staging-only Turso credentials. All secrets belong in the deployment/secret manager, **never Git, logs, screenshots or test artifacts**. `NOTIFICATION_MODE=debug` logs beta secrets and is prohibited for shared or public staging.

## Updated staging smoke acceptance matrix (plan; not execution record)

- [ ] **Flags off baseline:** new social registration, follow mutations, claim and install endpoints return 404/not registered; phone-first waitlist with draft policy acceptance and beta invite/phone verification continue unchanged. No public product claims from draft terms.
- [ ] **Real PDS login (owner-controlled staging only):** test a disposable hosted AT Protocol account with metadata/PAR/PKCE/DPoP/callback, browser nonce, state replay/expired state, wrong Origin, token encryption, session logout/revocation, refresh, server restart and invalid DID/handle. Capture pass/fail evidence without tokens or user data.
- [ ] **Identity boundary:** unauthenticated or forged-DID requests cannot claim/follow/install; stolen/expired/revoked cookies fail, requester cannot claim another DID's existing profile, and an unverified `acct-` placeholder cannot become publicly claimed. A local DID match is not independent PDS handle resolution.
- [ ] **Default follow:** only an already claimed and public local /@chris DID may be followed; user may remove it, opt-out persists across logins, and explicit re-follow removes opt-out. No outgoing PDS follow messages.
- [ ] **Profile actions:** verify exact-origin CSRF rejection, unknown/private target rejection, no self-follow, idempotent follow and unfollow, and no unauthorized account write. Exercise bounded body sizes and rate limits on all sensitive mutations.
- [ ] **Manifest/approval (when PR #29 is actually merged):** published version manifest shows declared scopes and digest, missing/stale acknowledgement fails, immutable version and hash are recorded, approval requires administrator, duplicate decision fails, changed/withdrawn publication cannot be approved; approved status must not create an executable Clear installation.
- [ ] **Alpha invitation:** only server-verified `beta_invites.status='activated'` and an active user can be admitted; revocation immediately removes eligibility. OAuth alone must **never** grant private Clear or alpha rights. Enforce eligibility at every future private endpoint.
- [ ] **Operations:** record healthz, migration idempotency, CI commit SHA, rollout image digest, monitoring signals (5xx, latency, callback errors, invite abuse), alert owner, Turso backup and non-destructive rollback rehearsal. Owner records incident/rollback criteria and signs off.

**Launch blockers unchanged:** route-wide session-bound alpha enforcement, DID↔phone-first Clear account-linking policy, verified live OAuth and handle ownership, final privacy/terms and retention/deletion owner sign-off, abuse controls and production monitoring. This checklist is not proof those items passed.

## Owner-only actions (NOT performed by this documentation PR)

1. **Hosted PDS validation:** use an owner-controlled disposable account, complete authentic browser OAuth flow and produce a redacted staging test record; decide whether DID/handle proof is sufficient for public claims.
2. **Secrets and vendor permissions:** provision, rotate and revoke staging-only OAuth session key, app/admin secrets, Turso and email/SMS credentials through the approved secrets manager. Confirm separate provider scopes and no debug-secret logging.
3. **Staging deployment and rollback:** authorize an operator to deploy a reviewed exact SHA to a separately isolated staging environment, verify migrations and backups, run the smoke matrix, and rehearse reversible rollback. **Do not run these steps from this PR.**
4. **Product/legal gate:** finalize phone/social identity linking, privacy/terms, retention and deletion schedules, invite notice, default-follow disclosure and owner approval for a limited private alpha. Placeholder draft notices are not final.
5. **Go/no-go:** after tests and risk review, record a signed decision to enable specific staging flags. Public launch remains a separate security and legal approval.
