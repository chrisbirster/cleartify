# Audit checkpoint 1: startup, configuration, and storage foundation

Date: October 27, 2026. Revision: `12390da9504ae91821851d027556968268fa0290` (`dev`).

This is a partial audit, not whole-project acceptance. No runtime source was changed. Product baseline follows the owner's social-first description; Clear's eventual role remains unresolved and does not block technical review.

## What this section does

The executable reads environment configuration, opens and bootstraps the database, optionally reconciles interrupted fake runs, creates security and identity services, conditionally registers API features, then starts the HTTP server. Operations hardening optionally adds structured redaction and SQL readiness. SIGTERM/SIGINT initiate shutdown.

All optional feature booleans read by `config.Load` default false. The frontend remains a separate application. Social behavior, protocol publishing and installs require explicit configuration; the existence of their service packages does not make them available to users.

## Confirmed findings

### F01 — P1: error objects bypass hardened log redaction

`internal/operations/logging.go:16-18` only suppresses error/message/detail/query attributes when their slog kind is String. Typical call sites, including `cmd/cleartify/main.go:68`, pass Go error objects, represented as Any. JSONHandler subsequently serializes their error text, allowing any sensitive values in the error to reach logs even when hardening is on.

Reproduction: a temporary package test passed `errors.New("SYNTHETIC_AUDIT_SECRET")` as the `error` attribute through RedactingHandler. The output retained the synthetic marker. No real secrets were used. The probe was removed after execution. Existing TestStructuredRedaction only exercises an error string and therefore misses the bypass.

Recommended fix: redact sensitive field names independently of value kind; resolve LogValuer values before deciding how to handle structured values. Add regression coverage for ordinary errors, nested groups, and resolved values. Review message text separately: the handler currently preserves the top-level message by design.

### F02 — P2: listener failures exit successfully

`cmd/cleartify/main.go:208-218` logs an unexpected ListenAndServe error, shuts down, then returns normally. The process exit code is 0, so scripts and process supervisors can interpret a failed start as successful completion.

Reproduction: built `/tmp/cleartify-audit-bin`; ran with an isolated environment, in-memory database, synthetic APP_SECRET, disabled notifications, and `PORT=invalid-port`. Observed `http server failed` and process exit code 0. No remote service was contacted and no listener opened.

Recommended fix: propagate serve/shutdown errors through a run function that permits deferred cleanup, then exit nonzero at the entrypoint. Test invalid listener configuration and a genuine occupied-port failure.

### F03 — P2: fake-accounting settings do not enable the service

`internal/config/config.go:63-64` reads FAKE_ACCOUNTING_ENABLED and FAKE_SELLER_FEE_BPS, but `cmd/cleartify/main.go:124-182` does not construct or pass api.Options.FakeAccounting. `internal/api/server.go:192-200` requires that service for dispute/operator routes. These settings have no effect in the executable. This is a static wiring finding, not an HTTP reproduction or a claim about real payment support.

Recommended action: decide whether this deferred marketplace feature should be wired as an explicitly local simulation or documented as library-only. Do not enable real commerce as part of this repair.

### F04 — P2: local-transport setting constructs and discards the transport

`cmd/cleartify/main.go:116-117` creates a LocalTransport and immediately discards it. Its constructor only stores dependencies (`internal/interop/local_transport.go:13-15`); it starts no worker or listener. Enabling AGENT_LOCAL_TRANSPORT_ENABLED does not expose usable transport through this executable.

Recommended action: label this integration library-only until there is an authenticated caller and a defined user journey. Adding transport is not a prerequisite for proving protocol sharing.

## Observations requiring subsequent review

- The offline preflight is a separate CLI command, not an automatic startup check. Its case-sensitive OAuth/debug comparisons differ from config.Load's case-insensitive handling. It checks some prerequisites but not every feature dependency. Do not treat a successful preflight as complete configuration validation.
- Database bootstrap executes CREATE IF NOT EXISTS statements individually, without a migration version ledger or one surrounding transaction. Existing installations will not receive future column changes merely by editing CREATE statements. Upgrade/rollback strategy needs explicit review before schema evolution.
- clear_user_sessions and its index are declared twice in schema.go (161-169 and 204-212). Currently identical/idempotent; maintenance duplication rather than a demonstrated runtime failure.
- Schema foreign keys are declared, but enforcement and connection behavior must be verified against the actual Turso driver. No missing-enforcement defect is asserted yet.
- Readiness checks a local SQL connection, not remote Turso durability or provider availability. This is a useful narrow check and should be described accordingly.
- APP_SECRET validation happens after database bootstrap, so some invalid security configurations can mutate local schema before startup aborts.

## Verification

- `go build -o /tmp/cleartify-audit-bin ./cmd/cleartify`: passed.
- Focused operations/database tests: passed (cache available).
- config and security packages: no existing test files.
- Temporary redaction probe: confirmed F01.
- Isolated invalid-listener process probe: confirmed F02.
- No production configuration, database, external identity provider, or deployment was touched.

## Coverage and inventory

- coverage.csv inventories every tracked file, source hash, reviewed spans, and pending state. Full-file reading does not imply behavioral acceptance; service callers/tests remain to be reviewed.
- routes.csv inventories explicit HandleFunc registrations in the main API and OAuth handler; it is a registration index, not an authorization matrix or full list of dynamically wrapped endpoints.
- schema-tables.csv inventories table declarations in the central bootstrap only. Additional tables created inside services remain to be inventoried.
- configuration.csv inventories environment reads in config.Load/main only; service-local reads and deployment settings remain to be reviewed.

Next checkpoint: database integrity/upgrade behavior and social OAuth/account creation, followed by first-follow, public profile claims, and reserved-profile ownership. Review fixtures and tests as well as handlers. Keep the functional distinction between AT Protocol login and local Turso social records explicit.
