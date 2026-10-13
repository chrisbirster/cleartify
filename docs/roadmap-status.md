# Cleartify milestone roadmap — 88-task evidence tracker

Updated: 2026-10-12. **Source:** repository `dev` and milestone PRs #5–#24 inspected in this workstream. The original 88-item checklist was not available as a canonical tracked file, so these are **88 normalized task descriptions derived from the agreed M0–M10 milestones**, not a claim of verbatim recovery. Statuses refer to evidence, **not deployment**.

**Legend:** `done` = code or documentation confirmed merged into dev; `in progress` = unmerged PR or pending implementation (explicit PR evidence noted); `blocked-on-owner` = product, legal, identity or launch approval required. An engineering task marked `in progress` without a PR is **not implemented**. CI success of a PR is not a merge. No task is marked done only because of a green PR.

## M0

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Audit existing repository baseline | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 2 | Review accepted architecture ADRs | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 3 | Define three-sided product model | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 4 | Define first-user journey | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 5 | Define public/private data split | **in progress** — PR #6 / M0 doc review; earlier foundation merged but product contract not approved |
| 6 | Agree social identity onboarding | **blocked-on-owner** — owner decision pending |
| 7 | Resolve default-follow consent | **blocked-on-owner** — owner decision pending |
| 8 | Approve pivot and M0 product contract | **blocked-on-owner** — owner decision pending |

## M1

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Land beta activation invitation flow | **done** — merged into `dev` |
| 2 | Fix beta invitation CI regressions | **done** — merged into `dev` |
| 3 | Implement waitlist endpoint | **done** — merged into `dev` |
| 4 | Implement landing and profile preview | **done** — merged into `dev` |
| 5 | Introduce rate limits | **done** — merged into `dev` |
| 6 | Add referral attribution | **done** — merged into `dev` |
| 7 | Commit lockfile and npm ci | **done** — merged into `dev` |
| 8 | Finalize terms/privacy and consent | **blocked-on-owner** — owner decision pending |

## M2

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add stable DID identity primitive | **done** — merged into `dev` |
| 2 | Define OAuth origin allowlist | **done** — merged into `dev` |
| 3 | Add encrypted Turso auth storage | **done** — merged into `dev` |
| 4 | Upgrade Go toolchain | **done** — merged into `dev` |
| 5 | Pin Indigo SDK and hash dependencies | **done** — merged into `dev` |
| 6 | Implement OAuth handlers behind flag | **in progress** — PR #18 merged into `dev`; exact CI/merge proof should be reconfirmed before calling done |
| 7 | Run real hosted-PDS E2E verification | **in progress** — pending implementation |
| 8 | Approve identity and phone account linking | **blocked-on-owner** — owner decision pending |

## M3

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add DID-keyed public profile schema | **done** — merged into `dev` |
| 2 | Add public profile read API | **done** — merged into `dev` |
| 3 | Replace mock profile view | **done** — merged into `dev` |
| 4 | Add DID-based follow model | **done** — merged into `dev` |
| 5 | Authenticate profile creation/claims | **in progress** — pending implementation |
| 6 | Implement owner-scoped follow/unfollow | **in progress** — pending implementation |
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
| 5 | Add authenticated self-service installations | **in progress** — pending implementation |
| 6 | Implement bounded declarative protocol runner | **in progress** — pending implementation |
| 7 | Implement cancel/retry/version-pin and audit | **in progress** — pending implementation |
| 8 | Approve consequential action permission policy | **blocked-on-owner** — owner decision pending |

## M6

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add revocable alpha membership data | **done** — PR #20 merged into `dev` |
| 2 | Tie alpha eligibility to verified beta activation | **done** — PR #20 merged into `dev` |
| 3 | Document staging smoke and gates | **done** — PR #20 merged into `dev` |
| 4 | Enforce session-bound eligibility on all routes | **in progress** — pending implementation |
| 5 | Run staging invite-to-activation smoke | **in progress** — pending implementation |
| 6 | Recruit creator and tester cohort | **in progress** — pending implementation |
| 7 | Measure first run and repeat use | **in progress** — pending implementation |
| 8 | Approve actual alpha invitations and privacy | **blocked-on-owner** — owner decision pending |

## M7

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Add local agent registration schema | **in progress** — PR #21 open, CI must be green; not merged |
| 2 | Create feature-flagged admin registration | **in progress** — PR #21 open, CI must be green; not merged |
| 3 | Test unverified claimed-owner cases | **in progress** — PR #21 open, CI must be green; not merged |
| 4 | Implement verified agent owner binding | **in progress** — pending implementation |
| 5 | Expose safe directory/capability fields | **in progress** — pending implementation |
| 6 | Build third-party self-hosted adapter | **in progress** — pending implementation |
| 7 | Add revocable scoped agent grants | **in progress** — pending implementation |
| 8 | Approve supported-agent listing claims | **blocked-on-owner** — owner decision pending |

## M8

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Define signed message envelope | **in progress** — PR #22 open; not merged |
| 2 | Add public-key and replay registry schema | **in progress** — PR #22 open; not merged |
| 3 | Test local Ed25519 envelope verification | **in progress** — PR #22 open; not merged |
| 4 | Implement key ownership and rotation | **in progress** — pending implementation |
| 5 | Implement replay-safe authenticated transport | **in progress** — pending implementation |
| 6 | Enforce consent and capability grants | **in progress** — pending implementation |
| 7 | Run two-agent isolated interoperability tests | **in progress** — pending implementation |
| 8 | Approve agent collaboration and data limits | **blocked-on-owner** — owner decision pending |

## M9

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Define listing and price-in-cents schema | **in progress** — PR #23 open; not merged |
| 2 | Test pricing constraints | **in progress** — PR #23 open; not merged |
| 3 | Document seller/payment/legal decisions | **in progress** — PR #23 open; not merged |
| 4 | Implement creator seller authentication | **in progress** — pending implementation |
| 5 | Implement entitlements and access control | **in progress** — pending implementation |
| 6 | Integrate processor with refunds and audits | **in progress** — pending implementation |
| 7 | Run limited paid creator pilot | **in progress** — pending implementation |
| 8 | Approve fees payout taxes and seller terms | **blocked-on-owner** — owner decision pending |

## M10

| # | Task | Status / evidence |
|---:|---|---|
| 1 | Write public launch checklist | **in progress** — PR #24 open; not merged |
| 2 | Threat model and security review | **in progress** — pending implementation |
| 3 | Finalize privacy and terms | **blocked-on-owner** — owner decision pending |
| 4 | Validate CI at final release SHA | **in progress** — pending implementation |
| 5 | Prove staging end-to-end core loop | **in progress** — pending implementation |
| 6 | Enable monitoring and incident alerts | **in progress** — pending implementation |
| 7 | Rehearse rollback database recovery | **in progress** — pending implementation |
| 8 | Owner go/no-go for public launch | **blocked-on-owner** — owner decision pending |

## Reconciliation notes

- M1 privacy/terms remains owner-blocked in draft PR #12; approval must precede public signup policy claims.
- M2 OAuth PR #18 was reported merged; flag remains off until real-PDS validation.
- M7 PR #21, M8 PR #22, M9 PR #23 and M10 PR #24 are not considered done until individually merged and reverified.
- M5 approval records do not execute Clear workflows; M6 membership predicate is not authenticated route-wide enforcement.
- Recheck this ledger after **each merge**; statuses cannot predict future merges or infrastructure smoke test results.