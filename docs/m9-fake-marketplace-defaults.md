# M9 fake-only commerce foundation

**DEFAULT (Oct 13 2026), owner can change.** Internal engineering defaults only; not legal, tax, pricing, chargeback, or public-policy promises.

- Service is disabled by default (`Enabled=false`) and not wired to any server route. API callers must someday supply a verified, server-owned authenticated DID; accepting arbitrary caller-supplied DIDs at network boundaries would be insecure.
- `Provider` documents charge/refund behavior, but **only concrete `FakeProvider` is accepted** by the service. No provider credentials, SDK, HTTP, money movement or remote integrations.
- Seller ownership requires both listing creator DID and published-protocol creator DID to match the trusted actor. A fake purchase of a listed one-time USD offering atomically records an active buyer entitlement and audit event. Only the buyer or verified listing/protocol seller may fake-refund, invalidating the entitlement and auditing the action.
- One entitlement per buyer/listing; simulated receipts are unique. Audit table is a local event record, not a financial ledger and not proof of payment. Existing pricing and seller-identity decisions await owner sign-off.
- Before production: vetted payment-provider implementation, durable idempotency, financial reconciliation, refunds/chargebacks, seller verification, taxes/compliance, revenue splits, policies, authorization review and entitlements integrated with actual installation and access checks.
