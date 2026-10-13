# Social, Clear and agent mutation hardening pass — October 13

This pass wraps **registered** sensitive local mutation routes and adds defense in depth to existing per-endpoint authentication:
- Methods POST, PUT, PATCH and DELETE under `/api/me/*`, `/api/social/accounts`, admin agent listings, existing Clear admin run requests, and future protocol install decisions are subject to a bounded per-IP/per-route limit: 12 requests per rolling 60-second fixed window. The limiter is in-process, best effort only (not a distributed Cloudflare rate limit); staging multi-instance results must be reviewed before public launch.
- **Cookie-authenticated social routes** require an exact configured `APP_ORIGIN` Origin header; requests with `Sec-Fetch-Site: cross-site` are rejected. For admin bearer-token routes, non-browser requests without Origin remain supported, but supplied mismatching origins and cross-site fetch metadata are rejected. Existing admin bearer checks are still mandatory.
- The source IP is parsed from the server-observed `RemoteAddr`, **never** client-supplied `X-Forwarded-For` without a configured trusted reverse proxy. Behind a load balancer this can group users behind one proxy: owner must decide trusted proxy configuration before using these limits as production abuse protection.
- Limiter memory is bounded, and old windows expire. `Retry-After` is returned with 429. The checks only operate on registered routes so feature flags off continue returning 404, and don't turn on ATProto OAuth or mutate production settings.

Tests cover cross-site requests, wrong Origin, no authenticated social session, forged cookie, per-route burst throttling, X-Forwarded-For spoofing, and disabled endpoints. This does not implement global distributed rate limiting, a real-PDS security review, or end-to-end approval permission validation.

**Integration notice:** M5 install flow PR #29 and M6 readiness PR #30 are separate and may not yet be merged. This generic guard will apply to their recognized registered sensitive routes once their PRs are integrated into `dev`, but regression tests must be rerun after the actual merge. No deployment, secrets or payments were changed.
