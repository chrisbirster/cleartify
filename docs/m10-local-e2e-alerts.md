# M10 Local staged E2E and alert hooks

**DEFAULT (Oct 13 2026), owner can change.** This test-only staging harness runs via `go test -p 1 ./internal/e2e ./internal/alerts`; the regular Go CI discovers it as part of `go test -p 1 ./...`.

It constructs an in-memory local Turso/SQLite database with normal schema migrations, an `httptest` server handler, deterministic fixed clock, and the existing fake-only marketplace provider. Stages check DB availability, local health, disabled features returning 404, the fake-only provider, and synthetic failure signaling. Every stage runs in-process: no listener, external network, third-party provider, deployment, keys or agent execution.

`alerts.Sink` is an internal interface with `Noop` and bounded `Local` implementations. It holds typed, timestamped events only in memory. There is no external exporter, address, webhook or notification configuration. A stage failure emits an alert event; sink failures propagate. These hooks are **not** global monitoring or a production alerting system.

**Boundary:** this is a staged local smoke/integration harness, not a complete cross-system user journey. Real ATProto OAuth and PDS, multi-device interop, production session authorization, paid entitlement enforcement, crash recovery, deployed database sync and real monitoring remain owner-blocked or incomplete. All production feature flags remain OFF; there are no public API changes.


## Round 3 comprehensive local journey

Run `go test -p 1 ./internal/e2e -run TestProductionShapedLocalJourney -count=1` locally or through the repository's **existing** Go CI (`go test -p 1 ./...`). No workflow changes are required.

The new staged integration test exercises actual local service/database primitives in this order: signup and consent placeholder persistence; activated fake Clear user, validated session and alpha membership; local fake social DID consent/challenge and database identity linking; immutable creator draft publication; unpaid/paid fake entitlement enforcement, repeat checkout idempotence and reconciliation; permission manifest acknowledgement and install approval; strictly in-memory declarative `runner` plan followed by persisted run snapshot/audit; moderation hold hiding/release restoring publication visibility; fake refund revoking access; disabled HTTP routes, local health and Noop alerts. Each stage uses the existing `Harness.Stage` with typed local alert signaling if it fails.

**Accuracy of the boundary:** This is a production-*shaped* local integration journey rather than a literal single HTTP browser journey. Signup/session/link/publication/install/checkout/moderation orchestration uses real repository service functions and database tables, with local fake identities and fake provider. The final execution is an in-memory sandbox plan with a durable fake-run record; no real executor. The separate API, OAuth-fake, loopback TCP and signed replay tests remain independently covered by `go test -p 1 ./...`; this one test does not start the external-facing service, perform real OAuth, contact PDS, place payments, deliver alerts or enable production feature flags. CI has no external networking prerequisites beyond its existing module/package setup.

Behavioral flags remain **OFF by default**. This test never turns on production routing, alters deployment settings or creates legal-policy commitments.
