# Round 4 item 5 — M1/M3/M4/M6 cross-milestone local integration

**Local-only, fake/test data, all runtime flags still OFF by default.** No staging deploy, real PDS, real commerce, final legal wording or launch approval.

The consolidated in-memory SQL integration test exercises M1 waitlist/invite activated-user state; a real locally issued Clear session and M6 alpha admission, cross-user/no-session denial, and revocation; M4 author draft publishing through private/unlisted/public, permission-pinned installation preview, digest rejection and admin approval, unauthorized creator takedown rejection and reversible hidden/restored decision; and M3 impersonation-hold write denial followed by reversible release. Final Clear session revocation must return 401. All identities, data and tokens are synthetic.

This is **not** a complete authenticated HTTP publish/install/run end-to-end journey. Publishing and install approval use the existing service primitives, while alpha authorization is exercised through the actual HTTP router; the fake-run route additionally requires a genuine signed social OAuth browser session and an identity link. No mock PDS was added to this PR, and no assertion of actual runtime execution is made. Existing `internal/api/fake_install_run_test.go` and protocol runner tests cover isolated fake-run boundaries. M1 counter rules and consent versions retain their existing tests; no real bot-defense or approved legal policy text is implied.

Remaining gaps: unify the authenticated synthetic OAuth/browser identity with private Clear sessions for HTTP-only publish/install/run, test moderation action holds in those same live handler requests, test the appeals decision path together with takedown and discoverability in one journey, and validate multi-instance database/restart behavior. These are engineering follow-ups, not owner-verified staging evidence.

Run `go test -p 1 ./internal/api ./internal/protocols ./internal/runner` or the existing full CI. No feature is switched on, no runtime code changes and no restricted repository files or secrets are changed.
