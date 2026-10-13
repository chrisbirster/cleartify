# M6 live Clear session and alpha route integration

**DEFAULT (Oct 13 2026), owner can change.**

- When both `CLEAR_SESSIONS_ENABLED` and `PRIVATE_ALPHA_ROUTES_ENABLED` are enabled, private alpha routes now resolve their principal from a live, revocable server-side Clear session cookie and then verify current `alpha.Gate.Eligible`.
- Both flags remain **false by default**. With either dependency missing, authorization fails closed.
- Query/body `userId` is ignored. Forged cookies, revoked sessions, different Clear users, and revoked alpha membership are denied. No admin bearer privilege or AT Protocol session alone confers Clear eligibility.
- This integration currently protects the `/api/me/alpha-eligibility` status endpoint; future private Clear endpoints must be registered through the same middleware. It does **not** claim completed coverage of routes that are not yet implemented.
- Fresh re-auth for sensitive DID linking, absolute session lifetime, phone-based login for previously activated users, owner staging verification and full identity ceremony remain separate steps.
