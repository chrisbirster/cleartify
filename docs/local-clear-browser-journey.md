# Local Clear browser journey (no deployment)

This recipe is **development only**, based on PR #102's `/clear` frontend. It does not enable production flags or real SMS. It uses a synthetic A publication instead of real creator OAuth. The real A-publishes-in-browser path remains blocked on a verified ATProto OAuth session and a separate creator UI; do not describe the synthetic fixture as a completed public publishing flow.

## Requirements

Go version from `go.mod`, Node 26+ (for `node:sqlite`), npm, and OpenSSL. Run everything from the repo root. **Never use a production or existing personal database.** Only files in `.local/` may be used for fixtures.

The Vite dev server proxies `/api` and `/healthz` to `http://127.0.0.1:8080`; browsers use `http://localhost:3000`. This keeps the cookie and same-origin mutation checks intact. Do not set `VITE_API_BASE_URL`.

## 1. Initialize a disposable local database

In terminal A, run:

```bash
mkdir -p .local
export DATABASE_PATH="$PWD/.local/cleartify-dev.db"
export PORT=8080
export APP_ORIGIN=http://localhost:3000
export APP_URL=http://localhost:3000
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
go run ./cmd/cleartify
```

Once it says it is listening, stop it with Ctrl+C. The database schema is now initialized. In the *same directory*, seed a clearly synthetic fixture:

```bash
node scripts/dev-seed-guide.mjs --i-understand-local-fixtures-only
```

This inserts a test creator A's public guide, a test reader B identity using fictional `+15550002222`, and a guide step already due so the worker can display an in-app nudge after installation. The fixture is **not a real ATProto-authenticated creator publication**. No credential or login session is inserted.

Restart `go run ./cmd/cleartify` from terminal A without changing the environment.

## 2. Start the SPA

In terminal B, from the repository root:

```bash
npm ci
npm run dev -- --host 127.0.0.1
```

Visit <http://localhost:3000/clear>. Vite serves the SPA, forwarding its relative `/api` requests to the local Go backend. If `/api/me/clear-session` returns 401, that is normal before login.

## 3. Browser journey (real handlers, synthetic published guide)

1. Enter `+15550002222` on the Clear login screen. The backend responds generically whether the account exists or not; in this **local fixture**, the Debug Sender prints the one-use code in terminal A. Enter that code to verify the phone challenge. No authentication bypass occurs.
2. On the recipes panel, refresh the real published catalog. Select **Local demo: one small step**, fetch its public manifest and confirm the digest. Tick **create follow-ups** and install. The API validates visibility, status, permissions and acknowledged manifest inside a transaction.
3. Return to the primary Clear conversation and refresh. The fixture's one due guide follow-up appears with `open` status; the worker polls every 30 seconds and records the nudge in the in-app queue. **In-app delivered means queued in the database, not read or sent via SMS.**
4. Use `done <follow-up ID>` (mixed-case `DoNe` is also accepted). Refresh. The follow-up becomes done and drops out of the active nudge inbox. Try `remind me to check my list`: without a due time, it defaults to 24 hours and reports `dueTimeDefaulted:true`.
5. Reinstalling the same publication returns the existing installation rather than duplicating it. Logout is available through `POST /api/me/clear-session/logout`, or discard the entire disposable `.local/` database after stopping the server.

## Security and honest limitations

- Every feature is opt-in by flag in this shell; defaults remain **OFF** in normal application configuration. Debug mode binds the API to loopback, not a public interface.
- The fixed seed user and publication exist **only in a disposable local file**, never in production. Debug Sender codes appear in the local terminal: do not forward those logs. Never commit `APP_SECRET`, database files, or codes.
- The cookie is HttpOnly, SameSite Strict, Secure on HTTPS, and scoped to `/api/me/`. Web client uses `credentials: 'include'` only for same-origin local paths.
- Invalid phone codes return generic challenges, and creation and nudge limits remain enforced. Expect 404/503 when routes are off, 401 when unsigned in, 409 for manifest changes/conflicts, and 429 for limits.
- The frontend is **deterministic commands only**. It does not perform unrestricted language understanding, external execution, or real reminders through SMS. `/social` remains a separately labeled creative mockup.
- True creator A publishing in a browser still needs a verified ATProto OAuth session and creator publishing workflow; this local guide seed is intentionally labeled as a substitute, not evidence the full A-publishes browser flow is working.
- This PR does **not** add or activate a remote staging environment, real OAuth, payments, external SMS, DNS, launch or a production scheduler.

## Tests

Run `npm run build` and `go test -p 1 ./...` for repository-wide checks. The existing Go two-person integration test covers authentication, publication and user isolation. A true Playwright browser test against a seeded local backend has **not been added to CI yet**: the current workflow does not provision a browser runtime or an OAuth creator session. The synthetic seed above provides the repeatable manual browser path without claiming browser automation succeeded.
