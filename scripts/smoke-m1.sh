#!/usr/bin/env bash
set -euo pipefail

# Read-only smoke test. Never submits real waitlist or beta activation records.
# Usage: API_BASE_URL=https://api.cleartify.com WEB_BASE_URL=https://cleartify.com bash scripts/smoke-m1.sh

: "${API_BASE_URL:?Set API_BASE_URL to the running API origin}"
: "${WEB_BASE_URL:?Set WEB_BASE_URL to the deployed web origin}"

api="${API_BASE_URL%/}"
web="${WEB_BASE_URL%/}"

case "$api" in https://*|http://localhost:*|http://127.0.0.1:*) ;; *) echo "API_BASE_URL must use HTTPS outside localhost" >&2; exit 2;; esac
case "$web" in https://*|http://localhost:*|http://127.0.0.1:*) ;; *) echo "WEB_BASE_URL must use HTTPS outside localhost" >&2; exit 2;; esac

echo "[smoke] API health"
health="$(curl --fail --silent --show-error --location --max-time 15 "$api/healthz")"
if ! printf '%s' "$health" | grep -Eq '"ok"[[:space:]]*:[[:space:]]*true'; then
  echo "health response missing ok:true" >&2
  exit 1
fi

echo "[smoke] landing page"
curl --fail --silent --show-error --location --max-time 15 --output /dev/null "$web/"

echo "[smoke] API rejects malformed waitlist data (no writes expected)"
status="$(curl --silent --show-error --max-time 15 --output /dev/null --write-out '%{http_code}' \
  --header 'Content-Type: application/json' \
  --data '{"name":"","email":"bad","phone":"invalid"}' "$api/api/waitlist")"
if [ "$status" != "400" ]; then
  echo "expected 400 for invalid signup; got $status" >&2
  exit 1
fi

echo "[smoke] PASS (read-only checks only; no signup, messaging or beta activation exercised)"
