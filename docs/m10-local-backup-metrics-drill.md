# M10: isolated local backup/restore drill and durable local metrics

**Default OFF; internal engineering test only (October 13, 2026).** No deployment, production connection, external alert export, keys, money, PDS or legal-policy changes.

## Run the drill

From repository root, run `sh scripts/m10-local-drill.sh`, or `go test -p 1 ./internal/database ./internal/alerts`. The current Go CI runs all tests through its existing `go test -p 1 ./...` step; **no workflow edits**.

`TestLocalBackupRestoreDrill` creates an isolated file-backed Turso/SQLite database under an OS temp directory, migrates and inserts synthetic waitlist data, **closes the writer**, copies it to a new temp backup, restores to another new temp path, checks the copy SHA-256 and verifies SQLite integrity plus expected row count. The drill refuses existing destination files, non-temp paths, empty or nonregular source files and WAL/journal sidecars; never copy a live or uncheckpointed database. Its narrow temp-only scope is an explicit safety constraint, **not** a production backup implementation or point-in-time restore system. The script takes no source path or credentials.

`alerts.Config` has `enabled: false` by default and a bounded list of threshold rules using local metric names, integer thresholds and internal event keys. `NewDurable(tempMetricsJSONL, config, sink)` persists append-only JSONL cumulative counts with file sync; it loads those counts when restarted and emits an alert **only on reaching** a configured threshold. Sink is `alerts.Noop` by default or bounded `alerts.Local` for inspection. Files must be inside OS temp and have a `.jsonl` suffix. Rules are tested for durability, disabled behavior and single threshold emission. It does not expose HTTP, perform remote writes, schedule monitoring, guarantee concurrent-process consistency, or replace production observability.

## Stop conditions / limitations

Do not pass a production database to the drill. Do not extend this script to accept arbitrary file paths without a separate security review. Do not assume live-copy consistency, schema compatibility across different versions, a real recovery time objective, alert delivery, multi-instance durability, or compliance with backup retention obligations. Owner action is still required for independent backups, scheduling, encryption, tested off-host recovery and production alert endpoints. All service runtime feature flags remain OFF; the metrics component must be opted into explicitly by a future caller.
