# M1 smoke checks and rollback runbook

Status: operational checklist / pre-production validation. This document does **not** establish that production deployment or external SMS/email delivery has passed.

## Audit context (2026-10-12)

Reviewed `feat/m1-beta-activation` (PR #7) and `dev` documents:

| M1 requirement | Evidence | Still needed |
| --- | --- | --- |
| Lockfile and reproducible frontend install | CI workflow uses `npm install`; `package-lock.json` not found on PR #7 branch at inspected path | Confirm lockfile strategy; switch to `npm ci` only after a committed, validated lockfile |
| Real waitlist API/dedupe/validation/consent | `POST /api/waitlist` and Go tests in PR #7 | Confirm live deployed API and error paths |
| Memorable landing and profile preview | `dev` README documents `/` and `/@chris` | Browser/mobile acceptance review |
| Abuse prevention | Request body limit verified in handler | Signup IP/account velocity limits and bot strategy not verified |
| Privacy/terms | Activation docs require acceptance | Published policies, version capture and actual forms not verified |
| Referral tracking | Waitlist docs describe optional source metadata; API hardcodes source `landing` | Explicit validated referral code/source fields, privacy-safe analytics |
| CI | PR #7 reports previous green CI; workflow builds web, Go and Docker | Check new head SHA, deploy environment and smoke checks |
| Rollback | Deployment docs describe topology, no step-by-step rollback | This runbook adds rollback procedure; rehearse with actual target infrastructure |

These are evidence-based gaps, not proof that components are missing on every branch; inspect open PR #7 before duplicating code.

## Smoke test usage

```bash
API_BASE_URL=https://api.cleartify.com WEB_BASE_URL=https://cleartify.com bash scripts/smoke-m1.sh
```

The script makes only safe, synthetic HTTP requests: GET health, GET landing, and invalid waitlist POST that must return 400. It does **not** prove DB writing, invite email, phone verification, CORS in a real browser, transport delivery or consent persistence.

### Before any production rollout

1. Record the previous known-good release tag, image digest, Cloudflare asset version and Fly machine configuration. Back up/verify the Turso local database and remote sync state. Never put tokens in incident reports.
2. Confirm only one writer against the sync-backed local database. Check volume attachment and database migration compatibility before starting.
3. Deploy to staging and run the read-only smoke script; perform a manual test account signup and consent persistence using an approved disposable phone/email, then test invite+activation under strict limits.
4. Verify 201 for valid signup and no duplicates; inspect activation token/code handling, rate limits, delivery-provider receipts and error paths.
5. Confirm cross-origin browser signup from the actual `APP_ORIGIN`, logs without PII/secrets, and waitlist records recoverable after restart.
6. Deploy production gradually, run read-only smoke, monitor errors/latency/provider failures, and record the deployed digest.

### Rollback trigger

Rollback on sustained 5xx, schema incompatibility, incorrect consent handling, misrouted verification codes, accidental data exposure, duplicated sends, or irrecoverable queue failures. Pause outbound notifications immediately on abuse/privacy incidents.

### Rollback sequence

1. Disable invite issuance and outbound message jobs if implicated; preserve audit logs without tokens.
2. Restore previous known-good Fly image/machine configuration using the recorded digest, **not** a floating tag. Keep persistent volume and secrets; do not delete the DB.
3. Restore previous Cloudflare static deployment if frontend/API compatibility is broken.
4. Re-run `scripts/smoke-m1.sh`; then manually verify the critical account path on staging. Confirm deploy version and error metrics.
5. **Never blindly roll back destructive database migrations.** Apply a reviewed forward-compatible repair or restore from a verified backup only after assessing data loss and remote sync divergence.
6. Record timeline, affected users, decisions and a follow-up PR with prevention tests.

### PR gate

CI success on a PR means code checks passed; it does not mean staging/production deployment, external delivery or live rollback rehearsal passed. Paste links to the exact checks and any staging smoke output in the PR before merging.
