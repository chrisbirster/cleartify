# M9 round 5 — immutable fake sales and local operator boundary

All behavior remains **fake-only**; no money movement, real provider/PDS access, deployment, legal publication, or owner approval is implied.

- Checkout's server-resolved listing amount/currency are validated as positive USD (up to 1,000,000 minor units), and stored in an append-only `fake_sale_snapshots` row in the same transaction as entitlement and purchase audit. Input is limited to listing ID; client amounts and currency are not accepted.
- Reconciliation uses the purchase-time snapshot, **not the listing's current price**, including after a listing price change or refund; immutable ledger rows remain unique by entitlement and entry kind. Existing records predating this migration lack snapshots and must be backfilled through an explicitly reviewed migration before reconciliation (no guessed historical price).
- Admin bearer + exact Origin are required for opt-in fake dispute resolution/refund routes. They register only when `FakeCheckoutEnabled`, `FakeAccounting.Enabled` and fake service dependencies are explicitly set in the API options; no default runtime wiring is added. An open buyer-originated dispute is required to refund through this operator API. Retry is rejected.
- Buyer disputes are opened through `POST /api/me/fake-entitlements/{id}/dispute` using a validated OAuth session DID and exact Origin. The operator cannot supply or impersonate a buyer DID. Resolution/refund require a previously opened dispute.
- No new public claims, fee promises, provider credentials, payout functionality, or real chargebacks.
