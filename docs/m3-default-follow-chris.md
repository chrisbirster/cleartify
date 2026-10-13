# M3 decision 3 — local automatic follow of /@chris

This slice implements the October 13 owner decision **"Following is automatic but removable"** for newly authenticated **social** accounts only. It does not alter phone-first private Clear onboarding.

- **Disabled by default:** set `SOCIAL_DEFAULT_FOLLOW_ENABLED=true` *and* `ATPROTO_OAUTH_ENABLED=true` (the latter MUST remain disabled until real-PDS verification). Disabled means no onboarding hook and no unfollow endpoint.
- After Indigo successfully verifies the OAuth subject, the server creates an **unclaimed, nonpublic placeholder profile** keyed by that DID, if absent. It adds a **local Turso** DID-keyed follow for the existing, claimed, publicly visible `/@chris` profile. Handles are not stable identities, and no specific DID is hardcoded.
- A signed-in person may send `DELETE /api/me/following/chris` from the configured `APP_ORIGIN`; the follower DID is obtained solely from the verified local OAuth browser session. This removes the follow and records a durable opt-out to prevent relogin from restoring it. No third-party follows can be removed using a caller-provided DID.
- Never follow a placeholder, private, unclaimed or non-existent Chris profile. If `/@chris` is absent, the sign-in remains valid but default follow cannot occur until a subsequent verified login after that profile is claimed.
- **No AT Protocol PDS writes, graph updates, agent network requests or outbound calls are added by this feature**. The existing Indigo OAuth login itself performs standard OAuth network operations when deliberately enabled. No automatic social publishing, no fake endorsements.
- Public profile claim and handle proof remain separate M3 work. **Do not enable OAuth or the follow flag in production** before real PDS tests, owner trust/consent review and staging checks.
- CI tests verify new account follow, private placeholder visibility, repeated-login idempotence, user removal and persistent opt-out, and absence of follows for missing creators.
