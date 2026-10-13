# M10 Local staged E2E and alert hooks

**DEFAULT (Oct 13 2026), owner can change.** This test-only staging harness runs via `go test -p 1 ./internal/e2e ./internal/alerts`; the regular Go CI discovers it as part of `go test -p 1 ./...`.

It constructs an in-memory local Turso/SQLite database with normal schema migrations, an `httptest` server handler, deterministic fixed clock, and the existing fake-only marketplace provider. Stages check DB availability, local health, disabled features returning 404, the fake-only provider, and synthetic failure signaling. Every stage runs in-process: no listener, external network, third-party provider, deployment, keys or agent execution.

`alerts.Sink` is an internal interface with `Noop` and bounded `Local` implementations. It holds typed, timestamped events only in memory. There is no external exporter, address, webhook or notification configuration. A stage failure emits an alert event; sink failures propagate. These hooks are **not** global monitoring or a production alerting system.

**Boundary:** this is a staged local smoke/integration harness, not a complete cross-system user journey. Real ATProto OAuth and PDS, multi-device interop, production session authorization, paid entitlement enforcement, crash recovery, deployed database sync and real monitoring remain owner-blocked or incomplete. All production feature flags remain OFF; there are no public API changes.
