# M6 private Clear route authorization matrix — Round 3

**Default OFF:** existing `PRIVATE_ALPHA_ROUTES_ENABLED=false` keeps previous behavior; when enabled, the Clear-session validator must produce a server-validated principal and the alpha eligibility service must verify active membership on every protected private request. Missing/invalid session -> 401; nonmember/revoked membership -> 403; missing dependency or database error -> 503. Client-supplied userId/DID fields never identify a Clear principal.

| Protected private route (when registered) | Required action |
|---|---|
| `GET /api/me/clear-session` | Read private authenticated session |
| `GET /api/me/alpha-eligibility` | Read private alpha eligibility |
| `POST /api/me/fake-install-runs` | Create a fake Clear run |
| `GET /api/me/fake-install-runs/{id}` | Read a fake Clear run |
| `POST /api/me/identity-link/start` | Start private Clear/social link |
| `GET /api/me/identity-link/preview` | Preview link |
| `POST /api/me/identity-link/confirm` | Confirm link |
| `DELETE /api/me/identity-link` | Disconnect linked identity |

**Separate trust boundaries:** `POST /api/me/clear-session/logout` stays callable after membership revocation to allow secure sign-out. `/api/admin/clear/*` retain their existing independent administrator authorization; these are **not** private-user endpoints. Social-only `/api/me/*` actions use OAuth-based social authorization and are not private Clear resources. Public discovery, waitlist, beta, health and OAuth bootstrap remain outside alpha gating.

Route registration is still subject to its existing independent feature flags and dependencies; wrapped routes continue their own identity-link, approval, ownership and access validations after membership checks. This covers current private Clear routes; new routes require explicit classification and guards. No legal text, real PDS, secrets or deployment.
