# Audit checkpoint 7: onboarding, private login, and moderation

Revision: `12390da9504ae91821851d027556968268fa0290`. October 27, 2026. Full first-party service/handler/test reads; synthetic local probes only. Notification senders are disabled/debug, so these are pre-launch code defects rather than evidence of compromised production users.

## Reproduced findings

### F22 — P1: verified phone can differ from the phone receiving the challenge

waitlist.Service.Join publicly upserts an existing entry by normalized email, including phone. Beta.Start sends a code to the phone loaded at that moment; Beta.loadActivation later joins the current waitlist row, and Verify creates phone_identities using its current phone. No immutable challenged-phone binding is stored or checked.

Probe: join with phone A → invite → start and capture SMS sent to A → resubmit the same waitlist email with phone B → verify using the correct code sent to A. An active account was created with B marked verified. The created identity did not match the SMS destination. No control of B was proved.

Impact: public waitlist mutation can alter the credential attached to another email's in-progress activation. A prospective attacker would need the relevant timing/email and a victim to complete activation; no live takeover was attempted. This must be fixed before real onboarding.

Fix direction: bind invite/activation to immutable identity details (or a versioned pending identity change), prohibit unauthenticated changes to an in-progress/activated account's credentials, and require a new challenge for a new phone. Revalidate the exact challenged destination atomically on activation.

### F23 — P1: revoked beta activations still create users

Invite revokes earlier sent/challenge invites, but Verify rejects only activated/locked/expired states, not all non-challenge states. Its final conditional UPDATE checks status=challenge without checking RowsAffected; it can commit user/phone/starter rows even when that update changes zero rows.

Probe: start an invitation, issue a replacement invite (revoking the old challenge), then verify the old activation token with its correct code. Verify succeeded and created an account. This bypasses invitation revocation; it does not automatically confer alpha membership because that predicate independently checks activated invites.

Fix direction: require a current challenge state, current expiry/attempt state and exact phone binding inside the activation transaction. Claim the activation with a conditional update whose affected-row count must be one; rollback every account side effect otherwise. Add revocation/concurrency/failure tests.

### F24 — P2: returning-login lockout does not recover after cooldown

Login.Start preserves clear_login_challenges.attempts on conflict (`internal/clearsession/login.go:108-112`). The hourly throttle resets, but the challenge's per-user failed-attempt count remains at five forever. New codes cannot succeed.

Probe: request a code → five wrong guesses → advance clock 61 minutes → request a new code. A new SMS was delivered, but verifying that correct new code returned ErrChallenge. No administrative reset path is exposed.

Fix direction: separate rolling abuse counters from challenge-local attempts. Preserve abuse limits across quick resends but reset a new challenge appropriately after the defined lockout expires. Verify wrong-code concurrency, restart and cooldown recovery.

### F27 — P1/integration: split-origin authenticated browser paths are incomplete

The deployment docs prescribe separate SPA/API origins. BetaActivate fetch omits credentials=include, so cross-origin activation does not retain its session cookie. API CORS omits Access-Control-Allow-Credentials and allows GET,POST,OPTIONS only, excluding existing PUT/DELETE routes for follow/block/unlink.

HTTP preflight probe confirmed allowed Origin receives allow_methods=GET,POST,OPTIONS with empty allow_credentials even when PUT or DELETE is requested. Source review confirms the frontend's fetch credential omission. This is static/protocol-level evidence, not a full credentialed live-OAuth browser test.

Fix direction: decide a supported same-origin proxy or credentialed same-site deployment, then align cookie attributes, request credentials, Origin/CSRF checks and allowed methods. Do not permit wildcard credentialed origins or relax identity checks. Test real browser cookies and preflights, not only Go's cookiejar.

### F29 — P2: activation refresh/retry cannot recover its session

BetaActivate stores activationToken only in a component signal. Start consumes the invite token before SMS delivery; if sending fails the response omits the newly created activation token. Reloading after a successful start also loses that token, while the URL still contains the already-used invite token.

Evidence: source trace through Beta.Start's conditional update/push/send ordering and BetaActivate's signal-only state. Existing one-use-link test proves reusing Start is rejected. This scenario was not manually exercised with a live delivery provider.

Fix direction: define a resumable, browser-bound activation session and safe delivery retry/outbox. Preserve single-use/replay guarantees while allowing ordinary refresh and transient delivery failure recovery.

## Moderation review

Reports/blocks are local records. Holds are evaluated through an explicit admin action; report insertion does not automatically evaluate. Hold/case/audit changes span separate transactions during evaluation, while a decision and hold release share a transaction. Creator draft creation does not call enforceSubjectHold; publishing does so only with the broader gate. Existing cross-milestone tests often call the gate helper directly rather than the actual draft/install/run endpoint.

Published protocol takedown/hold enforcement remains inconsistent (F09/F10/F17); stored blocks have no interaction enforcement (F07). Dismissed reports are still counted by impersonation evaluation because it does not filter status, and a later evaluation can recreate a released hold from unchanged evidence. These need a defined report lifecycle and endpoint-level integration tests, rather than treating helper tests as comprehensive enforcement.

A creator can request takedown only for their own publication. This is not a complete third-party rights-holder complaint/correction workflow for reserved celebrity profiles or community adaptations. No moderation staff UI or final public policy is present.

## Positive evidence and scope

Local tests cover normal invite/start/verify, hashed secrets, code expiry/lockout mechanisms, normalized contacts, consent recording, private-session ownership, unlink/recovery, and administrator/Origin checks. Those tests passed, while the new transition probes exposed F22-F24. Temporary probes were removed.

Signup's process-local limiter groups proxy peers unless trusted proxy handling is explicitly configured; sensitive mutation limits always use the socket peer. These are deployment availability considerations alongside the already recorded signup finding. Public-facing notification adapters, durable delivery jobs, final consent wording and retention/deletion implementation remain incomplete.
