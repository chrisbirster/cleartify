# Owner decision 1 — AT Protocol login required for social accounts

The owner's October 13 wording was **"at protocol login required"**. A Cleartify **social** account can only be created after a successfully verified Indigo OAuth browser session. A waitlist signup, phone number, private Clear account, user-supplied handle or submitted DID does not grant social identity.

Implementation: `POST /api/social/accounts` is registered only when `SOCIAL_ACCOUNT_REGISTRATION_ENABLED=true` **and** `ATPROTO_OAUTH_ENABLED=true`. Both default **off**; do not enable OAuth until it passes real hosted-PDS verification and security review. The API derives the DID from an unexpired, locally stored OAuth browser session with a decryptable Indigo OAuth session; rejects unauthenticated and cross-site requests; creates an unclaimed, nonpublic, DID-keyed placeholder only. Public handle claiming and display identity checks require additional work, and neither are implied by this endpoint. This route neither calls an outside API nor accepts caller-supplied identity for account creation.

**The private phone-first Clear path and existing waitlist/beta activation remain independent of AT Protocol login.** Do not add an OAuth requirement to `POST /api/waitlist` or beta activation routes. The existing `ATPROTO_OAUTH_ENABLED` default stays false; no production configuration or deployment changes are in this PR.

Before turning on social signup: complete real-PDS end-to-end login, expired/revoked session tests, supported handle resolution/ownership proof, public disclosure permissions, and account-linking decisions. This is a feature-flagged foundation for the social product contract, **not** a production release switch.
