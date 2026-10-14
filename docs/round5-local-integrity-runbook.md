# Round 5 local integration / fake ledger integrity runbook

## Demonstrated gap and narrow fix
Review of current dev found that `FakeAccounting.Reconcile` used `INSERT ... ON CONFLICT DO NOTHING` for existing immutable sale and reversal ledger rows, without validating their contents on repeat reconciliation. A corrupted existing active sale could be silently accepted; a previously recorded reversal could differ from its source sale without triggering an error. This matters when fake purchase, refund, mutable fee configuration and accounting retries intersect.

The explicit fake accounting reconciliation now verifies the identity and arithmetic of every recorded sale against its immutable sale snapshot, and every reversal against its recorded sale, even when an idempotent insertion conflicts. Historical fee basis points are not recomputed for existing sales: changing the configured fee must not invalidate a correctly recorded historic sale or change a refund reversal. Corruption denies the reconciliation and rolls back that transaction, without replacing records.

## Local verification
Run `go test -p 1 ./internal/marketplace ./internal/database ./internal/atproto ./internal/interop` and `go test -p 1 ./internal/marketplace -run TestRound5ExistingFakeLedgerIntegrityCrossBoundary -count=10`. The focused regression exercises fake purchase, first accounting insertion, reprocessing after fee configuration changes, rejection of inconsistent sale arithmetic, repair of the synthetic fixture, refund and reversal creation, and rejection of altered reversal accounting without duplicate or lost rows. It does not claim to test a real payment provider, remote settlement, production key rotation or PDS verification.

## Boundaries and remaining blockers
This is local fake-provider bookkeeping only, with no automatic scheduler changes, new privileged endpoints or enabled-by-default flags. Production payment settlement is still deliberately mocked; a real provider, isolated executor hardening, independent review, staging deployment, legal/counsel decisions, and launch approval are not achieved by this PR. Other boundaries (authenticated local worker, local OAuth migration) are covered by their own tests; this patch does not change them. Operator intervention for a damaged ledger requires investigation and a verified local backup; do not silently rewrite ledger history.
