# M3/M7 owner-approved internal enforcement defaults — October 13, 2026

**INTERNAL DEFAULTS APPROVED BY OWNER (Chris). NOT FINAL LEGAL WORDING. COUNSEL REVIEW PENDING.** These are conservative engineering decisions, not public terms, a moderation SLA, a legal finding, or permission to enable production features. Existing feature flags stay OFF.

## Moderation policy and response targets

| Signal | Default action | Enforcement status |
|---|---|---|
| Reserved platform identity or ASCII lookalike | Reversible impersonation hold; queue operator review | Implemented in local evaluator, existing moderation flags |
| 2 distinct impersonation reporters | Reversible impersonation hold; queue review | Implemented as `ImpersonationDistinctReportHoldThreshold=2` |
| 3 distinct spam reporters | Flag for manual review; do not auto-ban or suspend | **Policy default only**; no automated spam escalation built |
| 3 distinct harassment/abuse reporters | Flag for manual review; do not auto-ban or suspend | **Policy default only**; no automated abuse escalation built |
| Severe verified repeat abuse | Human-approved account disable/ban only after evidence review | **Manual policy only**; no ban workflow/API built |
| Individual block or a single uncorroborated report | Record signal; do not automatically hold or ban | Existing block/report storage |

Count **distinct verified reporter identities** rather than report rows; do not rely on volume alone for permanent restrictions. Identity claims and handle screenings are local until real DID/PDS verification is approved. Failed database checks should deny protected actions rather than silently allow. Do not automatically punish a reporter solely for a mistaken report. Owners/reviewers must preserve evidence and allow reversibility.

**Reviewer:** Owner (Chris) until named, independently authenticated moderation operators and role separation are installed. An appeal is filed through authenticated self-service API, assigned to an authorized human reviewer, and must not be auto-rejected or auto-accepted. Proposed **internal targets, not guarantees:** initial moderation triage within **24 hours** and appeal decision within **5 calendar days**, excluding cases requiring external evidence. These durations are constants for future instrumentation; **no queue scheduler, enforced deadline, notification or SLA implementation exists today**. Users can obtain reversal via existing admin hold-release and audit workflow. Released cases keep audit history; underlying heuristic/report signals may cause a later reevaluation, which requires operator oversight to avoid repeat false positives.

## Agent ownership, delegated access and revocation

Before an agent acts on a user's behalf, require a **server-authenticated owner DID**, an explicitly registered agent bound to that owner, **Ed25519 possession proof over an agent ID + DID + public key + one-use challenge**, and current verified registration/active owner binding. Challenge TTL is **5 minutes**; reused or expired challenges are invalid. **This is not independent real-PDS attestation.** Agent-to-agent delegation also needs verified bindings for both agents, the authenticated owner, explicit scope, and expiry. No ambient/wildcard authority.

Permitted grant scopes (allowlist only): `message.preview`, `protocol.preview`, `local.read`. They permit only the named local low-risk operation; no posting, purchasing, payment initiation, deletion, credential access, filesystem writes, network sending or arbitrary executor privileges. **Default suggested TTL: 1 hour; enforced maximum: 24 hours.** Existing Grant API requires caller to supply a TTL rather than automatically using the suggested value. Deny omitted, zero, negative or over-maximum TTL; check binding, scope, expiry and revocation on *every* protected operation. Owner may revoke one grant or revoke binding; binding revocation invalidates dependent access. An unverified/disabled agent cannot receive trusted delegated access. Audit proof and decisions without retaining signing private keys.

Existing enforcement reads central `internal/policy` constants for impersonation threshold, agent challenge lifetime, maximum TTL and scope allowlist. The moderation triage/appeal targets, spam/abuse review thresholds, suggested TTL and ban prohibition are **documented/tested policy constants only**, not falsely represented as operationally enforced controls.

No production rollout, real PDS, external service, authentication keys, workflow, deployment or legal wording changes are authorized by this document.
