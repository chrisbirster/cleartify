# M9 marketplace and creator monetization — data model only

The M9 slice models a protocol version's creator listing with visibility state, creator DID, currency, pricing kind and an integer price in **minor units** (USD cents) to avoid floating-point currency bugs. `draft` is the default, and no listing write or checkout routes are exposed. `free` and `one_time` are permitted kinds for planning; subscriptions, bundles and usage fees are deferred. Application-level validators must later require a free listing to have a zero price and verify that the claimed seller controls both the DID and protocol publication.

**Not implemented:** checkout, payment processing, creator payouts, customer entitlements, subscriptions, taxes, chargebacks, refunds, payment provider integration, marketplace ranking or financial records. A listed price does **not** imply a valid offer or ability to transact. Do not publish product purchase controls or capture payment details.

Owner decisions before commerce:
1. Select supported countries, payment/payout provider and identity verification/KYC requirements.
2. Define platform commission, inclusive/exclusive tax treatment, fees, refund period and seller support.
3. Define creator copyright/license warranty, moderation, prohibited content and takedown/dispute workflow.
4. Approve terms/privacy updates, billing retention, tax reporting and contractual seller onboarding.
5. Validate willingness to pay with a small manual pilot and calculate margins before adding a processor.

Future authenticated create/update routes must verify DID ownership, validate the referenced immutable published version and use audited states. This schema does not establish any right to sell another creator's protocol.
