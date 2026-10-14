# M10 Round 4 local operational readiness

**DEFAULT OFF.** No deployment, network provider, secret changes or public SLA is approved. The existing `GET /healthz` remains process liveness. New readiness and log behavior is active only with `OPERATIONS_HARDENING_ENABLED=true` (default absent/off).

## Operator preflight (no values)

Run `go run ./cmd/cleartify config-preflight` in the operator's existing environment. The command validates required variable **names and shapes** and prints only each variable name with a generic error code; it does not print their values, hostnames, secrets or connection strings. Required names: `APP_SECRET`, `APP_ORIGIN`, `DATABASE_PATH`. Optional `PORT` must be 1–65535. With `ATPROTO_OAUTH_ENABLED=true`, require `ATPROTO_PUBLIC_ORIGIN` and `ATPROTO_SESSION_KEY_B64`. `TURSO_REMOTE_URL` requires `TURSO_AUTH_TOKEN`. Debug notifications are rejected by preflight. Passing this offline preflight **does not validate actual database connectivity, secret entropy, credential authority, real OAuth or production readiness**. The preflight intentionally reads existing environment values internally to check validity but never displays them.

## Opt-in health, readiness, logging and shutdown

With `OPERATIONS_HARDENING_ENABLED=true`, `GET /readyz` performs a short (300 ms) SQL ping and returns either `{"ready":true}` (200) or `{"ready":false}` (503). It reports **only** a Boolean, not connection data. `GET /healthz` retains the existing lightweight liveness response and is not a database probe. Readiness is set to draining on SIGTERM/SIGINT before the pre-existing HTTP server's 10-second graceful shutdown; requests already in progress are handled using standard `http.Server.Shutdown`. There is no new handler, listener or external endpoint except the opt-in route on the existing server.

Logs under the opt-in mode use JSON `slog` and redact structured attributes whose names indicate secrets, tokens, credentials, cookies, passwords, keys, URLs or auth material. Untrusted error/detail/message/query attribute strings are suppressed. **Caveat:** redaction by attribute name is not a general-purpose guarantee against secrets placed in the primary log message or arbitrary unrecognized field names. Do not log raw values or include user inputs in event messages. Debug sender is rejected in hardened mode because it prints beta secrets.

## Acceptance and rollback

Run `go test -p 1 ./internal/operations ./...` and the existing Go/web CI. Verify preflight success and missing-name errors using synthetic environment data only; test readiness before/after Drain and liveness preservation; test JSON redaction with synthetic marker. Rollback: unset `OPERATIONS_HARDENING_ENABLED`, which removes `/readyz` and returns logging/shutdown to existing behavior. The preflight command remains an explicit optional operator tool. No deployment or public readiness claim follows from passing local tests.
