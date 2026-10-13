# Private alpha route authorization boundary

**DEFAULT (Oct 13 2026), owner can change.**

- `PRIVATE_ALPHA_ROUTES_ENABLED=false` by default. No existing admin endpoints change their authorization semantics.
- All newly registered private alpha handlers must use `Server.privateAlpha`: principal from authenticated Clear session, then `alpha.Gate.Eligible`. Ignore client-supplied user IDs.
- Fail closed: unauthenticated is 401, inactive/revoked membership is 403, missing dependency or database outage is 503.
- The initial GET `/api/me/alpha-eligibility` is a boundary test/status endpoint. With no Clear principal resolver configured it intentionally returns 503 even if flag enabled.
- **Integration follow-up:** after PR #33 lands on `dev`, wire its server-side `clearsession.Service.Authenticate` into `ClearPrincipalResolver`, add end-to-end cookie and membership tests. This PR does not claim route-wide enforcement on future private endpoints that do not yet exist. Owner controls flag enablement.
