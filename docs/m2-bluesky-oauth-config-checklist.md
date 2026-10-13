# M2 Bluesky OAuth configuration checklist — NOT READY TO EXECUTE

**Docs-only / internal verification notes (October 13, 2026). A public HTTPS origin (staging or HTTPS tunnel) and Bluesky throwaway account are both still missing. Deploy is NOT authorized. No credentials are included.**

| Config key | Required for later real Bluesky OAuth test | Acceptance / safety |
|---|---|---|
| `ATPROTO_OAUTH_ENABLED` | Set true only after owner approves isolated live-test environment | Keep OFF by default everywhere else |
| `ATPROTO_PUBLIC_ORIGIN` | Yes | Publicly reachable HTTPS API origin; no path/query/fragment or remote explicit port; derived client ID `/oauth/client-metadata.json` and callback `/oauth/callback` must be directly reachable |
| `APP_ORIGIN` | Yes | Approved public HTTPS frontend origin, no path/query/fragment; callback returns browser to `/?atproto=connected` beneath it |
| `ATPROTO_SESSION_KEY_B64` | Yes | Secret-manager reference ONLY; valid base64 for a 32-byte symmetric key; never print/copy into docs |
| `APP_SECRET` | If existing app security requires it | Reference by name only; test-only value managed outside repository |
| `DATABASE_PATH` | Yes | Dedicated isolated storage; no production records or connections |
| `TURSO_REMOTE_URL`, `TURSO_AUTH_TOKEN` | Only if approved isolated remote sync is used | Test-only references; omit for local database; do not share production credentials |
| `IDENTITY_LINKING_ENABLED` | No (optional separate linked-account test) | OFF unless approved; independent Clear and social authentication required |
| `CLEAR_SESSIONS_ENABLED` / `PRIVATE_ALPHA_ROUTES_ENABLED` | Only for separately approved linked private-alpha flow | OFF unless specifically authorized and backed by test-only services |
| `IDENTITY_RECOVERY_ENABLED` | No | OFF unless separate recovery validation approved |

**Metadata checklist:** anonymous HTTPS GET of `/oauth/client-metadata.json`; exact publicly resolvable `client_id`; exact `redirect_uris` callback; proper JSON Content-Type; declared `atproto` scope; SDK-consistent public-client auth, PKCE S256, DPoP/PAR/authorization properties; no wildcard redirect, credential content, localhost callback, surprise HTTP redirect, or internal hostname.

**Go/no-go:** owner test approval; two real-world prerequisites supplied; reviewed route/metadata consistency; distinct isolated datastore and secrets; no production flags or deploy changes; disposable account control; reviewed rollback path. If any condition is missing, **do not begin real OAuth verification**.

Refer to `docs/m2-bluesky-real-oauth-test-plan.md` for the step-by-step test, negative cases, evidence and rollback. This checklist records requirements; it does not turn anything on.
