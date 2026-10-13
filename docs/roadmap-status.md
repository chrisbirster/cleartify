# Cleartify milestone roadmap — 88-task evidence tracker

Updated: 2026-10-13. **Source:** verified GitHub merged PR states and dev commit c2b22ec16c1c662b29b86b65d287d06828a762b6 at audit refresh. PRs #7–#24 and #26–#31 are merged (including #12 and #23); #6 and #25 remain open. This is an **audit of merged work**, not a prediction that open PRs will merge. The original 88-item checklist was not available as a canonical tracked file, so these are **88 normalized task descriptions derived from the agreed M0–M10 milestones**, not a claim of verbatim recovery. Statuses refer to evidence, **not deployment**.

**Legend:** `done` = code or documentation confirmed merged into dev; `in progress` = unmerged PR or pending implementation (explicit PR evidence noted); `blocked-on-owner` = product, legal, identity or launch approval required. An engineering task marked `in progress` without a PR is **not implemented**. CI success of a PR is not a merge. No task is marked done only because of a green PR.


## Remaining owner-only actions (not completed by merged code)

These require decisions, credentials, independent verification, controlled staging or legal sign-off. CI and documentation do not constitute sign-off.

| Workstream | Owner action required | Gate |
|---|---|---|
| M0 product contract (#6) | Review/merge recorded October 13 decisions and approve or revise the expressly PROPOSED retention/deletion policy with category-specific periods | No final retention durations or public deletion promises |
| M1/M10 legal | Finalize and review draft /terms and /privacy, access/deletion procedure and public notice | PR #12 is development-only placeholder text |
| M2 social OAuth | Supply disposable **real hosted-PDS** account, isolated staging secrets and authorization for login, callback, refresh, logout/revocation, nonce/replay, session restart, DID/handle checks | AT Protocol OAuth flag remains off; tests do not substitute for real PDS |
| M2/M3 identity | Decide DID-to-phone-first Clear identity linking, independent handle ownership and rotation, profile disclosure, default-follow notice and impersonation safeguards | Local DID-matched claims/follows are not independent PDS handle proof |
| M4 creators | Approve creator licenses, attribution, remix, report/takedown and permission taxonomy | Discovery schema is not official endorsement |
| M5 Clear | Sign off consequential permissions, runner sandbox, approved install activation, revocation, cancellation and audit | Requests/approval only; **no real execution or usable install** |
| M6 private alpha | Provision staging secrets/providers/Turso, authorize a staged deploy/rollback, run hosted PDS/beta smoke, approve tester cohort, identity-bound access and alpha go/no-go | Checklist merged, live smoke not performed |
| M7/M8 agents | Decide allowed agent identity claims, delegation, privacy, consent and abuse controls | Agent registry/envelope are local groundwork only |
| M9 commerce | Decide provider, country scope, seller terms, fees, refunds, taxes, KYC and payouts | Pricing schema only; **no payment processing** |
| M10 public launch | Approve legal/security review, alerts and incident ownership, backup/rollback proof, exact release-SHA acceptance and public go/no-go | Launch not authorized |

**Status rule:** The 88 tasks below are a normalized reconstruction, not the recovered original checklist. A row remains **in progress** if only its foundation is merged, and **blocked-on-owner** if the mandatory remaining step is an owner decision, secret, staging action or approval.

## M0

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Audit existing repository baseline | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 2 | Review accepted architecture ADRs | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 3 | Define three-sided product model | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 4 | Define first-user journey | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 5 | Define public/private data split | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 6 | Agree social identity onboarding | **in progress** — owner said "at protocol login required" Oct 13; flag-off social-account PR #27 merged, but M0 product contract #6 remains open |
| 7 | Resolve default-follow consent | **in progress** — owner said "Following is automatic but removable"; local default-follow/unfollow PR #26 merged with its flag default off, while M0 PR #6 remains open |
| 8 | Approve pivot and M0 product contract | **in progress** — owner said "Social-first"; PR #6 document still open, so contract not yet merged |

## M1

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Land beta activation invitation flow | **done** — merged into `dev` |
| 2 | Fix beta invitation CI regressions | **done** — merged into `dev` |
| 3 | Implement waitlist endpoint | **done** — merged into `dev` |
| 4 | Implement landing and profile preview | **done** — merged into `dev` |
| 5 | Introduce rate limits | **done** — initial signup throttling merged; PR #31 merged extra CSRF / in-process mutation limits. Distributed abuse protection and trusted-proxy staging validation remain outstanding |
| 6 | Add referral attribution | **done** — merged into `dev` |
| 7 | Commit lockfile and npm ci | **done** — merged into `dev` |
| 8 | Finalize terms/privacy and consent | **blocked-on-owner** — PR #12 merged **draft** policy acceptance/version and placeholder pages for `dev` only; legal wording, retention and public-launch approval still pending |

## M2

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add stable DID identity primitive | **done** — merged into `dev` |
| 2 | Define OAuth origin allowlist | **done** — merged into `dev` |
| 3 | Add encrypted Turso auth storage | **done** — merged into `dev` |
| 4 | Upgrade Go toolchain | **done** — merged into `dev` |
| 5 | Pin Indigo SDK and hash dependencies | **done** — merged into `dev` |
| 6 | Implement OAuth handlers behind flag | **done** — PR #18 merged into `dev`; feature flag remains off pending real-PDS verification |
| 7 | Run real hosted-PDS E2E verification | **blocked-on-owner** — requires owner-controlled hosted-PDS account, staging secrets and authorized end-to-end exercise; not yet verified |
| 8 | Approve identity and phone account linking | **blocked-on-owner** — owner decision pending |

## M3

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add DID-keyed public profile schema | **done** — merged into `dev` |
| 2 | Add public profile read API | **done** — merged into `dev` |
| 3 | Replace mock profile view | **done** — merged into `dev` |
| 4 | Add DID-based follow model | **done** — merged into `dev` |
| 5 | Authenticate profile creation/claims | **in progress** — PRs #27/#28 merged OAuth-DID-bound private registration and local DID-match claim; public handle ownership is not independently verified against a real PDS; flags remain off |
| 6 | Implement owner-scoped follow/unfollow | **done** — PRs #26/#28 merged local session-owned follow/unfollow with durable /@chris opt-out; no PDS graph writes, flag off |
| 7 | Enforce reports/block/anti-impersonation | **in progress** — pending implementation |
| 8 | Validate profile disclosure and creator attribution | **blocked-on-owner** — owner decision pending |

## M4

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Define immutable protocol version schema | **done** — merged into `dev` |
| 2 | Add public discovery read endpoint | **done** — merged into `dev` |
| 3 | Test hidden draft/withdrawn records | **done** — merged into `dev` |
| 4 | Add authenticated creator publishing | **in progress** — pending implementation |
| 5 | Validate workflow permissions and input contracts | **in progress** — pending implementation |
| 6 | Implement remix and attribution policy | **in progress** — pending implementation |
| 7 | Seed 10 original licensed protocols | **in progress** — pending implementation |
| 8 | Approve creator license/IP policy | **blocked-on-owner** — owner decision pending |

## M5

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add durable Clear run request schema | **done** — merged into `dev` |
| 2 | Require exact version installation for requests | **done** — merged into `dev` |
| 3 | Add admin-only request/approval APIs | **done** — merged into `dev` |
| 4 | Test idempotent one-time decisions | **done** — merged into `dev` |
| 5 | Add authenticated self-service installations | **in progress** — PR #29 merged permission-manifest preview, explicit acknowledgement, pinned version/spec digest and approval request. No executable installation record or verified DID-to-private-Clear-user mapping; approval returns installed:false/executed:false |
| 6 | Implement bounded declarative protocol runner | **in progress** — pending implementation |
| 7 | Implement cancel/retry/version-pin and audit | **in progress** — PR #29 merged immutable version pin and approval metadata for install requests; real execution retry/cancel and auditable outcome lifecycle do not exist |
| 8 | Approve consequential action permission policy | **blocked-on-owner** — owner decision pending |

## M6

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add revocable alpha membership data | **done** — PR #20 merged into `dev` |
| 2 | Tie alpha eligibility to verified beta activation | **done** — PR #20 merged into `dev` |
| 3 | Document staging smoke and gates | **done** — PRs #20 and #30 merged eligibility and updated default-off feature flag/staging smoke/owner-only checklist; no staging test or deployment was performed |
| 4 | Enforce session-bound eligibility on all routes | **in progress** — pending implementation |
| 5 | Run staging invite-to-activation smoke | **blocked-on-owner** — PR #30 checklist merged, but owner-controlled staging secrets, actual staging deployment, hosted-PDS and beta activation smoke evidence are still missing |
| 6 | Recruit creator and tester cohort | **in progress** — pending implementation |
| 7 | Measure first run and repeat use | **in progress** — pending implementation |
| 8 | Approve actual alpha invitations and privacy | **blocked-on-owner** — owner decision pending |

## M7

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add local agent registration schema | **done** — PR #21 merged into `dev` (unverified local/admin-only registration) |
| 2 | Create feature-flagged admin registration | **done** — PR #21 merged into `dev` (unverified local/admin-only registration) |
| 3 | Test unverified claimed-owner cases | **done** — PR #21 merged into `dev` (unverified local/admin-only registration) |
| 4 | Implement verified agent owner binding | **in progress** — pending implementation |
| 5 | Expose safe directory/capability fields | **in progress** — pending implementation |
| 6 | Build third-party self-hosted adapter | **in progress** — pending implementation |
| 7 | Add revocable scoped agent grants | **in progress** — pending implementation |
| 8 | Approve supported-agent listing claims | **blocked-on-owner** — owner decision pending |

## M8

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Define signed message envelope | **done** — PR #22 merged into `dev` (offline signature/registry only) |
| 2 | Add public-key and replay registry schema | **done** — PR #22 merged into `dev` (offline signature/registry only) |
| 3 | Test local Ed25519 envelope verification | **done** — PR #22 merged into `dev` (offline signature/registry only) |
| 4 | Implement key ownership and rotation | **in progress** — pending implementation |
| 5 | Implement replay-safe authenticated transport | **in progress** — pending implementation |
| 6 | Enforce consent and capability grants | **in progress** — pending implementation |
| 7 | Run two-agent isolated interoperability tests | **in progress** — pending implementation |
| 8 | Approve agent collaboration and data limits | **blocked-on-owner** — owner decision pending |

## M9

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Define listing and price-in-cents schema | **done** — PR #23 verified merged into `dev` (schema/tests/docs only; no payment processing) |
| 2 | Test pricing constraints | **done** — PR #23 verified merged into `dev` (schema/tests/docs only; no payment processing) |
| 3 | Document seller/payment/legal decisions | **done** — PR #23 verified merged into `dev` (schema/tests/docs only; no payment processing) |
| 4 | Implement creator seller authentication | **in progress** — pending implementation |
| 5 | Implement entitlements and access control | **in progress** — pending implementation |
| 6 | Integrate processor with refunds and audits | **in progress** — pending implementation |
| 7 | Run limited paid creator pilot | **in progress** — pending implementation |
| 8 | Approve fees payout taxes and seller terms | **blocked-on-owner** — owner decision pending |

## M10

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Write public launch checklist | **done** — PR #24 merged into `dev` (launch checklist document only) |
| 2 | Threat model and security review | **in progress** — PR #31 merged origin/CSRF defense and in-process mutation throttling with tests; full OAuth, proxy, agent, data-handling, dependency and launch threat reviews not yet signed off |
| 3 | Finalize privacy and terms | **blocked-on-owner** — owner decision pending |
| 4 | Validate CI at final release SHA | **in progress** — pending implementation |
| 5 | Prove staging end-to-end core loop | **in progress** — pending implementation |
| 6 | Enable monitoring and incident alerts | **in progress** — pending implementation |
| 7 | Rehearse rollback database recovery | **in progress** — pending implementation |
| 8 | Owner go/no-go for public launch | **blocked-on-owner** — owner decision pending |

## Reconciliation notes

- PR #12 has **merged into dev** with explicitly draft `/privacy` and `/terms` content and `CurrentPolicyVersion` draft marker. Owner approved this development-only foundation, **not** finalized legal wording or public launch. Final privacy/terms and retention/deletion sign-off remain blocked on owner/legal review.
- M2 OAuth PR #18 and flag-off social creation PR #27 are verified merged; ATProto login remains disabled until real hosted-PDS verification. Neither OAuth DID ownership nor local claims proves arbitrary handle ownership.
- PRs #7–#24 (including #12 and #23) and #26–#31 are verified merged to dev. PR #6 (owner contract/proposed retention policy) and this audit PR #25 remain open. M9 PR #23 is now done **only for listing/pricing schema, tests and decision documentation**, not payments or entitlements. M0 PR #6 contains the October 13 owner statements plus an explicitly unapproved proposed retention policy; it is not merged.
- M5 approval records do not execute Clear workflows; M6 membership predicate is not authenticated route-wide enforcement. Social owner decisions (required AT login; automatic but removable Chris follow; social-first; protocol authorship rule) are documented in open PR #6, not proof that OAuth real-PDS login or public claims are verified. Flag-off PRs #26 and #27 are merged, and PR #28 implements local DID-scoped claim/follow; these do not replace real PDS handle verification. PR #29 (install request and approval data), #30 (alpha flag/smoke documentation), and #31 (CSRF/rate-limit tests) have merged. Their bounded slices are done, but no Clear runtime execution, real staged alpha acceptance, distributed abuse limits or production launch is authorized.
- Recheck this ledger after **each merge**; statuses cannot predict future merges or infrastructure smoke test results.