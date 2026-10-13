# M3 anti-impersonation: local engineering defaults

**DEFAULT (Oct 13 2026), owner can change.** Internal implementation policy, NOT public promises, legal advice or a claim of independent account verification.

- New `social.Impersonation` is disabled unless instantiated explicitly with `Enabled:true`; there is **no production API route, automatic job, or server initialization**. Existing profile and moderation endpoints are unchanged.
- Reserved platform-like handle labels: admin, support, security, official, cleartify, moderator, staff. Reject direct matches and conservative ASCII-leetspeak lookalikes (0/o, 1/i, 3/e, 5/s; punctuation ignored). This is incomplete for Unicode and AT Protocol confusables; false positives must be reviewed.
- **Labels:** `unverified` by default; `locally_claimed_not_independently_verified` for claimed person profiles only. Neither label means external PDS verification. Active holds downgrade display classification to `unverified`; no verified badge is minted.
- **Report-driven holds:** at least two distinct local reporters with `impersonation` reports may trigger a reversible local hold; repeated submissions by one reporter count once. A reserved/lookalike handle may trigger a hold without reports. Hold/release operations do not delete an account or alter ATProto records.
- Hold release and evaluation are **unrouted library methods**; before enabling, require authenticated moderator authorization, appeal review, atomic report/hold mutations, appeal and audit retention, reporting abuse controls, independent handle/DID resolution, and enforcement of holds in discovery/follow/create actions. Current holds do not suppress existing public routes.
- No external moderation requests, real-PDS calls, credential changes, payments, deployment, secrets or launch.
