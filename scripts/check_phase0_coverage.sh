#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GO_COVER="$ROOT/.tasktmp/phase0-go.cover"
MIN_COVERAGE="90.0"

mkdir -p "$ROOT/.tasktmp"

cd "$ROOT"

go test -coverprofile="$GO_COVER" ./ ./cmd/cleartify ./data ./internal/app ./internal/auth/... ./internal/config ./internal/httpx ./internal/server ./internal/store

GO_TOTAL="$(go tool cover -func="$GO_COVER" | awk '/^total:/ {gsub("%", "", $3); print $3}')"
if ! awk "BEGIN { exit !(${GO_TOTAL} >= ${MIN_COVERAGE}) }"; then
  echo "Go line coverage is ${GO_TOTAL}%, expected at least ${MIN_COVERAGE}%."
  exit 1
fi

(
  cd web
  bun run test:coverage
)

WEB_TOTAL="$(node -e "const fs=require('fs'); const summary=JSON.parse(fs.readFileSync('./web/coverage/coverage-summary.json','utf8')); process.stdout.write(Number(summary.total.lines.pct).toFixed(1));")"
if ! awk "BEGIN { exit !(${WEB_TOTAL} >= ${MIN_COVERAGE}) }"; then
  echo "Web line coverage is ${WEB_TOTAL}%, expected at least ${MIN_COVERAGE}%."
  exit 1
fi

echo "Phase 0 coverage meets the ${MIN_COVERAGE}% minimum for the touched Go packages and web source."
