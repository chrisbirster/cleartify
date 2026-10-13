# M2 identity link persistence defaults

**DEFAULT (Oct 13 2026), owner can change.** Private storage foundation only. No feature enabled, API route, browser ceremony, automatic account merge or PDS access.

- At most one active Clear user per social DID and one active DID per Clear user, enforced by SQLite partial unique indexes.
- Expiring hashed challenge records carry server-selected principal identifiers and consent-version metadata; no raw token is stored.
- A transactional commit consumes the challenge, writes the link and appends minimal audit metadata. Rollback protects against duplicate claims.
- The storage API is deliberately **not an authentication boundary**. It must only be called after the separately verified active Clear session, fresh phone re-authentication, Indigo subject DID, OAuth state/replay validation and explicit final consent described in `docs/m2-atproto-clear-account-linking-design.md`.
- No public projection of private Clear data; no claims that a handle is PDS-verified.
- The owner may revise cardinality, recovery, consent text, session freshness policy or retention. None is a final legal promise.
- **Remaining work:** hook into fresh Clear auth and OAuth, consent preview/confirmation, unlink/session invalidation, rate limits, deletion and real-PDS owner gate. These storage primitives must not be treated as account-linking completion.
