# M1 signup abuse and consent record keys

**DEFAULT (Oct 13 2026), owner can change.** `DB_SIGNUP_ABUSE_ENABLED=false` and `CONSENT_KEYS_ENABLED=false` unless explicitly enabled. These are internal defaults and not public policy promises or legal advice.

When enabled, each waitlist request consumes one attempt in a transactional DB 15-minute bucket; the conservative default is five requests per peer per window. Peer IP is read from the network connection's RemoteAddr, not untrusted X-Forwarded-For, and represented as a keyed HMAC digest. Missing DB or sufficiently long secret fails closed. Legacy in-memory signup throttling still runs. Outstanding work includes reverse-proxy trusted-peer IP parsing, retention/cleanup of limiter buckets, device-independent bot challenges, email normalization and operational alerts. No raw IP is saved in the abuse table.

Opt-in consent-key plumbing adds two append-only records to an existing successful waitlist transaction: `policy_acceptance / policy_version_pending` and `email_marketing / marketing_consent_version_pending`. These are internal identifiers **only**. No consent statement, actual legal-policy version, or binding public wording is created or changed. The existing policy acceptance and marketing consent rows remain intact. Owner/counsel must approve policy text and stable version identifiers before product use.

No configuration for deployment or any external provider is touched.
