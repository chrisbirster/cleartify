# M2 — in-process fake PDS/OAuth lifecycle tests (Round 4)

**INTERNAL, FAKE ONLY, DEFAULT OFF.** This implementation does not make calls to Bluesky, a real PDS, a hosted OAuth provider, or any network endpoint. It does not authorize deployment, change existing production OAuth session behavior, or supply OAuth credentials.

`InProcessFakePDS` is a deterministic test double with synthetic DID/handle mappings, fake refresh tokens, one-time replacement on refresh, expiry, and revocation. `FakeLifecycle` requires explicit `Enabled:true`; otherwise its refresh, re-verification, and logout methods fail closed. Handle re-verification compares the freshly resolved DID against the already trusted DID, preventing silent account substitution after a handle transfer.

`FakePDSHandler` provides synthetic `/fake/oauth/refresh`, `/fake/pds/resolve`, and `/fake/oauth/revoke` routes, exercised **in process** using `httptest.NewRecorder` and `ServeHTTP`, with no bound socket or live network. This is a miniature local test protocol, **not** an implementation or conformance test of AT Protocol / OAuth wire semantics. Tests cover initial handle lookup, token refresh, stale refresh replay, exact expiry, re-assigned handle rejection, logout and revoked refresh, malformed routes and wrong methods.

**Not wired into the application's real OAuth routes:** production Indigo refresh and real PDS handle resolution remain to be verified separately with an owner-approved public HTTPS staging origin and dedicated throwaway Bluesky account. Token rotation persistence, OAuth provider refresh-token failure semantics, live session-cookie renewal, and provider-specific revocation/re-authentication need additional integration. This fake harness is not a substitute for the real-PDS test plan in `docs/m2-bluesky-real-oauth-test-plan.md`.

Use `go test -p 1 ./internal/atproto`. No workflow, `fly.toml`, `go.mod`, `go.sum`, secrets, staging or deploy changes are included.
