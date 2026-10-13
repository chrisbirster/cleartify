# Concurrency and lifecycle regression hardening (internal)

**Default OFF for all existing optional behavior.** This patch adds regression coverage; no production feature flag, secret, deploy, network integration or legal policy change.

- **M9:** parallel duplicate fake checkout on one buyer/listing must produce the same entitlement; concurrent local reconciliation must book the sale exactly once with the configured fee. Tests serialize this local SQLite connection intentionally to exercise application idempotency; it does **not** prove behavior under multi-region replica conflicts.
- **M5:** concurrently submitted identical scheduler job IDs must not duplicate work; parallel Tick calls must execute a single job and the submitted plan must be defensively copied.
- **M2:** encrypted OAuth state must reject key changes, alternate associated IDs and invalid keys. Rotating the symmetric key without a migration or re-authentication invalidates existing encrypted sessions; this is an expected fail-closed boundary.
- **M8:** pre-existing outbox tests cover lease fencing and revocation-induced dead letters. At-least-once delivery remains: a crashed worker may redeliver to an idempotent sink.
- **M1/M7:** existing session expiry and revoke, agent ownership/key revocation, and scoped-grant expiry tests remain part of the full Go suite. Real adapter revocation and end-to-end PDS testing remain separate owner-blocked work.

Run `go test -p 1 ./...` (the existing CI command). No new workflow, provider or background service is introduced. This is scoped local regression evidence, not a proof of distributed linearizability or arbitrary-code isolation.
