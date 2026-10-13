#!/bin/sh
set -eu
# Only uses disposable temporary databases created inside Go tests.
# This script accepts no database paths, credentials, or production targets.
if [ "$#" -ne 0 ]; then
  echo "usage: scripts/m10-local-drill.sh (no arguments)" >&2
  exit 2
fi
go test -p 1 ./internal/database -run '^TestLocalBackupRestoreDrill$' -count=1
go test -p 1 ./internal/alerts -run '^TestDurableLocalRulesAndFlagOff$' -count=1
