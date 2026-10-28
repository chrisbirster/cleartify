# Local remediation progress

October 27, 2026. These changes follow the audit of `12390da9504ae91821851d027556968268fa0290`. They are local working-tree changes, not a production deployment. Original audit findings and file hashes remain historical evidence.

| Finding | Local repair | Regression evidence |
| --- | --- | --- |
| F22: activation changes verified identity | Save name, normalized email and phone in a persistent invite identity snapshot. Start, resend and verify use that snapshot. Updates to the snapshot are rejected by a database trigger. | Change the waitlist identity, restart the database, then verify the original challenged phone and name. Snapshot updates fail. |
| F23: revoked activation creates account | Validate challenge state and atomically claim it in the same transaction that creates the account and starter installs. Check affected rows before creating account data. | Revoked invitations reject correct/incorrect codes and resends without changing state. A fault-injected zero-row claim leaves no account/install side effects. |
| Adjacent activation attempt limit | Preserve failed attempts when resending; reject locked or non-challenge sessions. | Four failures, resend, fifth failure locks the session and rejects the replacement code. |
| F24: login remains locked after cooldown | Reset the per-challenge counter when issuing a new challenge. Persistent phone/user throttles retain the shared budget across resends. | Existing resend/lockout assertions still pass; a fresh challenge after the throttle window can authenticate. |
| F01: error objects leak in logs | Redact sensitive field names regardless of value type; resolve lazy log values before recursively checking groups. | Error objects, bound attributes, maps, nested groups and lazy values are redacted while event names and safe status fields remain. |

The four initial activation regression tests, the login cooldown regression, and the logging regression failed against their previous implementations before the fixes were applied.

## Database upgrade

A versioned migration ledger now supplements the existing bootstrap schema. Migration 1 creates `beta_invite_identities` and its update-protection trigger. Schema creation and version recording are transactional. Tests cover upgrading an unversioned database, repeated migration, rollback on a deliberately conflicting trigger, and retry after removing the conflict.

**Existing pending beta invitations must be reissued after this upgrade.** They have no trustworthy historical identity snapshot. The application deliberately does not backfill one from the mutable waitlist. Old links/challenges fail closed; issuing a replacement invitation creates a snapshot and restores the normal activation flow. Existing activated accounts and login sessions are not migrated or deleted.

Account creation remains local/private beta behavior; this does not implement the social onboarding gap in F06. Activation delivery/refresh recovery in F29 remains open. Real provider delivery and remote database replication have not been exercised.

## Verification

Focused activation/database, login and logging tests pass. After the final source edits, `go test -p 1 -count=1 ./...` passed across all packages (with permission for existing loopback tests), `go vet ./...` passed, and `git diff --check` passed. That first repair batch included no frontend changes.


## Second repair batch: browser routing and publication access

| Finding | Local repair | Evidence |
| --- | --- | --- |
| F28: public profile route is blank | Use a supported dynamic segment filtered to an `@` prefix, remove the prefix before the API lookup, and add a catch-all not-found page. Return the profile effect cleanup as required by Solid 2 so stale fetches abort. | Production web build passed. Local browser rendered a synthetic public profile with its handle/bio, a missing-profile message, and the not-found page for an unmatched URL. A final profile-to-home navigation produced no browser warnings/errors. |
| F27: split-origin credentials/preflights incomplete | Frontend fetches include credentials. API allows credentials and GET/POST/PUT/DELETE/OPTIONS only for the configured app origin. CORS responses vary by Origin, including responses to other origins. | New preflight regression failed before the fix and passes afterward, including rejection of credentialed CORS for other origins. Existing session/cookie and mutation-origin tests pass. |
| F09: paid discovery drops hidden takedowns | Compose paid and hold filters onto a shared publication accessibility predicate instead of replacing the query. | Full creator lifecycle test confirms hidden free guides stay hidden for all paid/hold flag combinations and reappear after restoration. |
| F10: general install ignores hidden takedowns | Share the accessibility predicate with direct-version reads and manifests; check request/approval eligibility and write within one database transaction. | Hidden preview/request/approval are denied, pending approval remains unchanged, restoration permits approval, private access is denied, and unlisted direct access remains available. |

The lifecycle regression failed against the previous discovery/install implementation. The full uncached Go suite, `go vet ./...`, `npm run build`, and `git diff --check` pass after these changes. Locked frontend dependency installation reported zero known vulnerabilities in this run. Temporary browser/API servers and generated build/dependency files were removed after verification.

F27 is repaired for the documented **same-site** deployment. Strict cookie policy and existing Origin/Sec-Fetch-Site checks remain enforced; unrelated-domain authenticated deployments are not supported. Browser verification exercised public profile reads, not a live OAuth/provider session or complete activation flow. README and deployment docs now spell out matching development hostnames and exact origin configuration.

At the end of the second batch, social creator holds at install boundaries, F11/F12 schema compatibility, legacy runtime authorization, and the remaining findings were still open. The takedown fix does not authorize real execution or implement missing social UI.


## Third repair batch — October 28, 2026

All remaining 21 numbered findings have local implementation changes. See [the complete repair/evidence matrix](remaining-findings-repairs.md), which covers F02–F08, F11–F21, F25–F26 and F29. Together with the first two batches, all 29 numbered findings have local remediation.

Database startup now enables foreign keys for every checked pooled connection and performs explicit historical relationship checks. Migrations 2/3 add pinned legacy request and installation identity. Unknown historical slug-only installs must be explicitly reinstalled; unpinned legacy requests cannot be approved. The migration deliberately grants no inferred authorization. Existing pre-snapshot beta invitations still need reissue.

A disposable local browser acceptance completed authenticated profile setup, public viewing, returning to saved settings, making the profile private and confirming that public viewing returns unavailable. No console warnings/errors appeared during the public-profile check. The real provider OAuth exchange was not exercised.

Fake payment result-write failures now roll back their reservation and purchase/refund side effects together. Key ownership remains distinct from administrator-reviewed agent activation. Existing feature flags remain disabled by default; no production integration was enabled.


Final verification for the third batch passed: `go test -p 1 -count=1 ./...` across all packages with local loopback permission, `go vet ./...`, `npm run build`, `git diff --check`, and `docker build` using the repository's pinned Go 1.26.1 builder. The final local image is `cleartify-audit:remaining-findings` (image ID `sha256:4a217e2f66a94ae1dc092be85c85676819585738761dcd8209a1be065473198b`). No container was deployed.

Creator sync failure retries were exercised across draft creation, publishing, takedown request and decision: identical retries recover the existing result, retry Push, and preserve one immutable version and the expected audit count. Conflicting draft payloads remain rejected. The explicit upgrade test preserves legacy ambiguous installs without creating a guessed binding.

The disposable browser fixture database/source, local servers, browser tab, `node_modules`, web build output and TypeScript build cache were removed after validation. Audit coverage hashes continue to describe the original reviewed revision rather than this changed working tree.
