# M5 bounded local preview runner — engineering defaults

**DEFAULT (Oct 13 2026), owner can change.** These are internal technical defaults, not public promises.

- `runner.New(false)` is the default: no production API route, background process, or external workflow invokes this runner. Enabling requires a future explicit, reviewed flag-gated integration.
- Plans are declarative and restricted to at most 32 steps, with only `local.set` and `local.assert` operations against an isolated in-memory map. Values are capped at 4096 bytes; identifiers allow only safe ASCII. No JS, shell, dynamic code, network, filesystem, remote agents, providers, payments, or actual user data operations.
- Execution is synchronous, bounded by step count, and checks context cancellation between steps. Failed assertions can be retried at most three times; the failed step is re-evaluated. Cancelled/completed runs cannot restart.
- In-memory per-run audit entries are ordered, include UTC time, step and attempt, and are **not durable compliance audit logs**. In-memory state is lost on process exit. The fake is intentionally not a security boundary for untrusted host plugins: no untrusted plugin code is ever run.
- Explicitly out of scope: persistent scheduling and audits, concurrency-safe external cancellation during a running step, authenticated user authorization, installation-to-run wiring, worker isolation, token usage, real PDS, real agent execution, and production enablement. These require separate design/security review and owner sign-off.
