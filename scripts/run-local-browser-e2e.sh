#!/usr/bin/env bash
set -euo pipefail
mkdir -p .local artifacts/screenshots
export PORT=8080
export APP_ORIGIN=http://127.0.0.1:3000
export APP_URL="$APP_ORIGIN"
export DATABASE_PATH="$PWD/.local/cleartify-dev.db"
export APP_SECRET="$(openssl rand -hex 32)"
export NOTIFICATION_MODE=debug
export CLEAR_SESSIONS_ENABLED=true
export RETURNING_LOGIN_ENABLED=true
export CLEAR_CONVERSATION_ENABLED=true
export SELF_SERVE_GUIDE_INSTALLS_ENABLED=true
export GUIDE_STEPS_ENABLED=true
export FOLLOWUP_DELIVERY_CHANNEL=in_app
export PRIVATE_ALPHA_ROUTES_ENABLED=false
export ATPROTO_OAUTH_ENABLED=false
backend_pid=""
web_pid=""
cleanup() {
 if [[ -n "$web_pid" ]]; then kill "$web_pid" 2>/dev/null || true; fi
 if [[ -n "$backend_pid" ]]; then kill "$backend_pid" 2>/dev/null || true; fi
 wait 2>/dev/null || true
}
trap cleanup EXIT
wait_health() {
 local url="$1"
 for _ in $(seq 1 60); do
  if curl --max-time 2 -fsS "$url" >/dev/null 2>&1; then return 0; fi
  sleep 1
 done
 echo "Local server did not become ready: $url" >&2
 return 1
}
./.local/cleartify-e2e > .local/setup-server.log 2>&1 &
backend_pid=$!
wait_health http://127.0.0.1:8080/healthz
kill "$backend_pid"
wait "$backend_pid" 2>/dev/null || true
backend_pid=""
node scripts/dev-seed-guide.mjs --i-understand-local-fixtures-only
./.local/cleartify-e2e > .local/browser-backend.log 2>&1 &
backend_pid=$!
wait_health http://127.0.0.1:8080/healthz
npm run build > .local/browser-build.log 2>&1
npm run preview -- --host 127.0.0.1 --port 3000 > .local/browser-vite.log 2>&1 &
web_pid=$!
wait_health http://127.0.0.1:3000/
npx playwright test -c playwright.config.ts
