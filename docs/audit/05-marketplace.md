# Audit checkpoint 5: fake marketplace, entitlements, and accounting

Date: October 27, 2026. Revision: `12390da9504ae91821851d027556968268fa0290` (`dev`). No runtime source changed. All probes used synthetic identities, a disposable local database, and FakeProvider; no purchases, refunds, or funds moved externally.

## What exists

The service deliberately accepts only the concrete FakeProvider. It creates buyer/listing entitlements, immutable sale-price snapshots, and purchase/refund audit rows. A separate accounting library creates seller ledger entries, reverses refunded sales, records disputes, and supports manually ticked reconciliation jobs with leases/retries. This is local simulation infrastructure, not live commerce.

The executable wires fake checkout, but does not wire FakeAccounting (F03), KeyedOperations, or ReconciliationJob. No marketplace listing creation/management endpoint or marketplace frontend was found in the route inventory. Fixtures and direct database setup are not a creator selling experience.

## Reproduced findings

### F18 — P2: checkout does not enforce published content or claimed seller eligibility

`internal/marketplace/fake.go:48-53` requires a listed positive-price offer and matching listing/protocol creator DIDs, but never verifies publication status, visibility, takedown, creator hold, or claimed profile. AuthorizeSeller only checks the same ownership correspondence, not creator eligibility.

Probe: seeded a private/unclaimed seller profile, a draft protocol, and a listed offer. Purchase succeeded and created an active entitlement. This is service-level behavior against seeded data; no listing-creation HTTP path exists for an attacker to create that state through the reviewed app.

Fix direction: enforce authoritative publication availability and the agreed seller eligibility at offer publication and checkout. Keep buyer refunds available after takedown; do not indiscriminately block refund paths when checking availability.

### F19 — P2/policy gap: unlisting an offer removes the paid-access requirement

`internal/marketplace/fake.go:97-99` treats absence of a currently listed priced offer as free access. The same rule applies when a formerly paid offer is changed to unlisted.

Probe: purchase → refund denied PublicationAccess as expected; changing only marketplace_listings.status to unlisted made PublicationAccess return true for the refunded buyer. It also removes the requirement for other nonbuyers. Publication status was published for this phase of the probe.

Boundary: docs/m9-fake-paid-access-enforcement.md explicitly identifies withdrawn/repriced listings as unfinished policy/testing. This is a reproducible access-state gap, not a claim that the docs promised that case already worked. PublicationAccess is an entitlement predicate, not a complete publication-visibility authorization check, and even returns true for nonexistent IDs with no priced offer.

Fix direction: distinguish free content from a withdrawn paid offer. Decide whether prior buyers retain use, and deny new paid access without an active entitlement unless the creator explicitly makes that content free. Apply the resulting rule consistently to discovery, manifests, installs, and runs.

### F20 — P2/library-only: idempotency can become permanently stuck after commit

KeyedOperations reserves an empty result row, commits Purchase/Refund through a separate service transaction, then updates the result in another statement (`internal/marketplace/idempotent_ops.go:49-66`). If the operation commits but saving its idempotency result fails, later reads reject that key as in-flight indefinitely. There is no lease/recovery step for this table.

Probe: injected a trigger rejecting UPDATE on fake_payment_idempotency. Keyed purchase returned an error after creating an active entitlement. After removing the trigger, retrying the same key still returned ErrDenied while the entitlement remained active.

Fix direction: commit effect/result together where possible, or implement durable recovery that reconciles a reserved operation with its committed entitlement/audit. Recovery must never blindly repeat a future non-idempotent payment. Keep this distinct from the reconciliation job's lease recovery, which does not repair keyed-operation reservations.

### F21 — P2/library-only: reconciliation accepts reversal rows for active sales

`internal/marketplace/fake_accounting.go:59-74` validates reversal entries only when entitlement.status is refunded. RunReconciliationJob first checks purchase/refund audits, but does not reject a reversal ledger row attached to an active entitlement.

Probe: reconciled valid simulated sales/refunds, then directly injected a balanced reversal for a still-active sale without a refund audit. RunReconciliationJob succeeded and SellerBalance became zero despite that active sale.

Boundary: this required direct database mutation, not a demonstrated buyer-accessible HTTP action. It is a gap in the accounting integrity checker—the same class of corruption the existing ledger regression tests are intended to detect. No real seller balance is involved.

Fix direction: require each reversal to correspond to a refunded entitlement and matching refund audit, and reject extra/orphan ledger rows as well as incorrect arithmetic. Retain original sale-fee snapshots during legitimate reversals.

## Existing behavior supported by tests/review

- Fake provider rejects nonfake receipts, invalid currency/amounts, and replacement provider implementations at the service boundary.
- Repeated purchases for an active buyer/listing reuse the entitlement; refunded entitlements are not silently reactivated.
- Refunds require the buyer or matching listing/protocol seller; disputes require the original buyer.
- Immutable sale snapshots preserve purchase-time price despite later listing changes.
- Legitimate reconciliation is idempotent; reversals negate original fees even when configured fees later change.
- Existing tests detect wrong ledger arithmetic and corrupted reversals on refunded entitlements, and fail closed on missing snapshots/audit mismatch.
- Job tests cover lease reclamation, stale-worker fencing, exhausted retries, concurrency, and fixed-width timestamp ordering.
- API operator tests and registration guards require configured dependencies/admin credentials; buyer identity is derived from OAuth rather than caller-submitted DID.

## Remaining integration observations

- Fake purchase/refund/accounting mutations receive *sql.DB and do not explicitly Push Turso sync; API wrappers do not add a push. This extends F13's local-commit versus remote-acknowledgment inconsistency.
- KeyedOperations is not used by the HTTP checkout/refund handlers. Repeated HTTP refund returns denial rather than the keyed wrapper's successful replay behavior.
- Preview denies paid manifests without considering the purchaser, while request/run checks can accept a purchaser entitlement. A buyer cannot complete the documented preview/acknowledgment loop through those endpoints alone unless another authorized manifest-delivery path is added.
- Moderation holds/block semantics need a consistent marketplace policy; the reviewed purchase actor helper only checks local OAuth authentication.
- Sale, refund, dispute, and reconciliation operations remain separate from listing UI, seller review, actual processor integration, and any seller payout mechanism.

## Verification and next checkpoint

Existing marketplace/API package tests passed. One removed fault-injection probe confirmed F18-F21. Full-file coverage records only outputs read without truncation; partially displayed tests remain pending/partial in the ledger.

For a social-protocol launch, recommend deferring real commerce until free publishing/discovery/configuration/use works. Preserve this code for later repair and evaluation rather than treating simulations as launch-ready billing.

Next: agent directory, ownership challenges, grants, and local interoperability. Then finish onboarding/frontend, moderation, persistence/operations and remaining tests, and consolidate a fix backlog against the social-first launch scope.
