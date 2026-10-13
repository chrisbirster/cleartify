# M5 fake local runner: restart recovery and durable audit

**Engineering default (October 13, 2026): OFF.** `FAKE_RUN_RECOVERY_ENABLED=false`. Recovery runs during application startup only when both `FAKE_RUN_RECOVERY_ENABLED=true` and existing `FAKE_INSTALL_RUN_ENABLED=true`. The fake runner's existing OAuth, alpha, Clear-session and publication restrictions remain intact; this change adds no endpoints, execution privileges or external services.

## Recovery rule

Before starting the HTTP listener, query every durable `clear_fake_install_runs.status='reserved'`. Such a reservation may have been interrupted before, during or after local fake execution; its true outcome cannot be inferred after a crash. **Never automatically replay or resume execution.** Transition each reserved record to `failed` with an explicit `outcome=unknown` recovery diagnostic and an append-only `recovered_failed` audit event in a single DB transaction. Do not modify already completed or failed runs. Re-running startup reconciliation is idempotent, and the unique install-request reservation still prevents duplicate execution attempts.

After commit, push to the configured database sync interface; **if enumeration, transaction, audit write or push fails, startup exits without serving traffic.** A failed remote push may leave local committed data, but service does not start and operators must resolve sync before restart; the recovery function will not replay a fake run. No external providers or network-only services are added.

## Local verification

`go test -p 1 ./internal/api` covers restart by closing/reopening a file-backed local database, multiple interrupted reservations, unchanged terminal rows, exactly-once audit events, repeat startup, transactional rollback under injected audit failure, missing dependencies and push failure.

This is fail-safe reconciliation rather than an actual scheduler or cross-process continuation. There is **no real runner**, no automatic task execution on restart, no endpoint to replay unknown outcomes, and no staging/deploy, payment, real-PDS, secrets, or legal-policy change. Production-grade recovery would require stronger multi-instance coordination, recovery runbooks and approved execution isolation.
