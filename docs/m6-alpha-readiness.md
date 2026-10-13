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
