# M10 public launch — release checklist

**Status:** Planning only. Nothing in this document constitutes approval to launch.

## Security review (blocking)
- [ ] Run a threat-model review for ATProto OAuth PKCE/DPoP/PAR, callbacks, trusted metadata, issuer/subject/DID binding, token encryption/key rotation, CSRF, cookie policies, refresh and revocation.
- [ ] Test real hosted-PDS login/logout, expiry and invalid callback in staging; keep `ATPROTO_OAUTH_ENABLED` off until signed off.
- [ ] Enforce authenticated per-user authorization and scoped approvals on profiles, follows, installations, run requests, outcomes, agent registration and publication.
- [ ] Prove alpha eligibility is fail-closed for all private routes, including revoked and inactive users, and test invitation rate limiting.
- [ ] Audit prompt-injection boundaries, untrusted protocol content and agent-originated messages before allowing execution or interop.
- [ ] Verify no private inputs, contact details, tokens, payments or memories reach ATProto records, logs, analytics, error responses or public pages.
- [ ] Review dependency licenses, vulnerability results, CI supply chain, permissions, secrets injection and reproducible container builds.
- [ ] Complete spam/report/block/impersonation/takedown escalation, audit logs, retention and abuse throttling.

## Owner/legal decisions (blocking)
- [ ] Approve and publish final Terms of Service, Privacy Notice and versioned user consent (PR #12 is a draft).
- [ ] Specify privacy contact, supported user age/jurisdictions, deletion/export procedure, data retention periods, backups and subprocessors.
- [ ] Review creator publishing license, protocol attribution, claimed profiles, trademark/IP disputes and the agent connection disclaimer.
- [ ] If payments launch, approve merchant-of-record/processor, seller onboarding, fees, payouts, refunds and taxes. **M9 schema is not payment enablement.**
- [ ] Approve exact launch feature set, public onboarding identity model, Chris-first-friend opt-in and ability to remove/decline connections.

## Operability, monitoring and rollback
- [ ] Identify and record release commit SHA, artifact digests, configuration, database migration plan and rollback image.
- [ ] Run full Go/web/Docker checks on current dev and the final release candidate (not old green heads only).
- [ ] Rehearse Turso backup and restore; verify single-writer assumptions, migrations, durable sessions and sync consistency.
- [ ] Validate health, readiness, browser CORS, genuine waitlist signup, phone activation, end-to-end protocol lifecycle and OAuth in staging.
- [ ] Define dashboards and alerts: API errors/latency, alpha invite/verification abuse, callback failures, run approval backlog, database durability, uptime and message provider faults.
- [ ] Document on-call escalation, pause switches, incident communication and audit-safe logs, excluding PII/secrets.
- [ ] Rehearse rollback of Fly image, Cloudflare assets and feature flags independently; **do not blindly reverse destructive DB migrations**.
- [ ] Verify legal links, mobile accessibility, privacy consent, user support, 404/unclaimed profiles and publicly visible disclosure of incomplete features.
- [ ] Do a limited consented beta soak with explicit go/no-go criteria before opening public signups.

## Exit gate

Owner signs off legal/security; engineering records green CI on the exact release SHA; staging smoke shows authenticated users can safely discover, install and complete a bounded protocol; abuse controls and rollback drills are proven. Otherwise **no public launch**. M10 remains blocked even if this documentation PR is green.
