# Design: link a verified AT Protocol DID to a private Clear account

**Status:** Design proposal, October 13, 2026 — **not implemented, not enabled, not a launch approval**. This is a documentation-only slice. The public social identity remains separate from the private phone-first Clear identity until the owner explicitly links both. Neither existing OAuth nor a telephone number on its own is sufficient to perform a link.

## 1. Goals, boundaries, and evidence currently available

- **Social side:** The Indigo AT Protocol OAuth flow (behind `ATPROTO_OAUTH_ENABLED=false` until real hosted-PDS verification) produces a stable subject DID and encrypted server-held OAuth session. `internal/atproto/social_session.go` checks an opaque browser cookie, expiry and the persisted decryptable OAuth session. **DID is the identity key; a mutable handle is not an identity credential.**
- **Clear side:** `internal/beta/service.go` can activate an invitation via an email link and one-time phone challenge, creating an active `users.id` with `phone_identities` and an activated `beta_invites` record. This proves completion of onboarding **at that moment**; it does not yet provide a complete, long-lived, authenticated *end-user Clear session* for subsequent sensitive link/unlink requests.
- **Private Clear assets:** `users`, `phone_identities`, `protocol_installations`, `clear_run_requests`, alpha membership, private run data and future local Clear memory must remain indexed and authorized by **Clear user ID**, not a user-supplied DID.
- **Public social assets:** `social_profiles`, local follow graph, creator publications and future AT Protocol records are separately subject to social disclosure rules. **Linking must never automatically make private Clear information public**, publish a follow, or push private data to a PDS.
- No automatic linking by matching email, handle, display name, phone, invite email, inferred identity, agent name, network address or device. No automatic account merging.

**Non-goals for this slice:** code changes, API routes, OAuth enablement, real-PDS requests, private memory synchronization, payment entitlements, agent delegation, public profiles, secret provisioning, deployment or production configuration changes.

## 2. Security invariants / authorization model

1. A link may be created only with **both independent proofs in the same short-lived server-bound linking ceremony**: (a) a valid Indigo OAuth session whose authenticated subject DID is read from server state, and (b) a **fresh Clear authentication/re-authentication** tied to the specific active `users.id`, requiring a verified phone challenge or an equivalent owner-approved higher-assurance Clear session. Merely having an old `activated_user_id` record is not an active Clear session.
2. User input may select an operation but may **never supply the authoritative DID or Clear user ID**. Those come from the validated server-side principals.
3. Prior to commitment, show **both identities and exact consequences**, collect explicit, affirmative consent, and write an auditable consent version. A later cancellation or timeout writes **no active link**.
4. A proposed conservative default is **one active social DID per Clear user and one active Clear user per DID**, enforced with transactional unique indexes. A replacement or recovery flow requires re-authentication, conspicuous confirmation and policy review; never silently transfer a DID or merge data.
5. Every Clear resource and private-alpha route continues to require **authenticated Clear identity plus its own authorization**. A linked DID is an association, **not permission** to enter Clear, receive an alpha membership, access private tasks or approve consequential actions. Linking does not create social OAuth login bypass for phone-first Clear.
6. Do not use a social handle as a join key. Validate handle-to-DID ownership against a genuine PDS before showing a public verified-handle badge; a local `social_profiles.did` match alone is insufficient to prove current handle ownership.
7. On unlink or account deletion, atomically disable link-based privileges, revoke link-related Clear and social sessions where applicable, and delete connection-scoped OAuth tokens on disconnect; do not delete an unrelated social account or AT Protocol public records without a separate explicit action.

## 3. Proposed data model (not a migration)

Create `identity_links` (or equivalent), storing only necessary metadata:

| Field | Proposed semantics |
|---|---|
| `id` | Random opaque server ID |
| `clear_user_id` | FK to `users.id` (private, active account) |
| `social_did` | Verified AT Protocol OAuth subject DID, normalized and indexed |
| `status` | `pending`, `active`, `revoked` (never use inactive association to authorize) |
| `linked_at`, `revoked_at` | Server UTC timestamps |
| `consent_version`, `consented_at` | Server-owned versioned explicit agreement |
| `authn_clear_at`, `authn_social_at` | Re-authentication freshness evidence (no raw credentials) |
| `created_at`, `updated_at` | Audit timestamps |

- Constrain *active* `clear_user_id` and `social_did` to one-to-one (SQLite/Turso partial unique indexes if supported; verify remote dialect and migration compatibility).
- Store short-lived **hashed** linking challenge tokens in `identity_link_challenges` with `nonce_hash`, `clear_user_id`, `social_did`, state, expiration, attempts, consent version and consumed timestamp. Keep raw secrets solely in HttpOnly Secure cookie or approved short-lived channel, never in logs.
- Append `identity_link_events` (create, consent, activate, reject, unlink, revoke, recovery) with actor type, minimal identifiers, event time and redacted evidence; apply owner-approved retention and sensitive-data access controls.
- **Do not copy** phone/email/contact fields into social profile tables or DID documents, and do not copy OAuth token ciphertext to Clear run or creator publication tables. OAuth credentials remain in the dedicated encrypted OAuth store and are removed when disconnected under the signed-off policy.
- Migration design should consider deletion/foreign keys and protect against races on duplicate links; never cascade private Clear resource deletes when merely unlinking.

## 4. Proposed linking ceremony and data flow

```text
Owner's browser
  | (A) authenticates private Clear account with fresh phone challenge/session
  v
Clear session -> server resolves active users.id; does NOT trust request userId
  | (B) "Connect my AT Protocol identity" -> read-only disclosure + consent
  v
Server creates expiring single-use link nonce bound to Clear session + origin
  | (C) Indigo OAuth login, state + PKCE + DPoP + browser-nonce validation
  v
Server resolves verified OAuth subject DID; checks encrypted session, expiry
  | (D) displays "Private Clear account X <-> AT Protocol DID Y"
  |     user explicitly confirms consent and disclosures
  v
Server rechecks both fresh principals + nonce + uniqueness under transaction
  | -> inserts/activates link and versioned consent + audit event
  v
Response: linked:true, publicExposureChanged:false
       /                   \
Private Clear authorizer    Public social authorizer
requires Clear session     requires social OAuth session
       \                   /
         NEVER automatically reveal Clear private data or grant alpha access
```

1. **Initiate:** Request comes from an authenticated Clear session with an allowed exact `Origin`, CSRF controls, bounded body, rate limits and no `userId` or `did` body fields. Challenge TTL suggested **10 minutes for design discussion only**; owner approval needed for actual value.
2. **Authenticate social:** Redirect through the current Indigo OAuth flow using the validated state/nonce ceremony; the browser returns to a server-held linking transaction with a DID from the OAuth result, not from a URL parameter. Treat OAuth error, replay, or cookie mismatches as fail-closed.
3. **Consent preview:** Confirm to the same signed-in user which private Clear account and DID will become associated, what this *does not* do, unlink and deletion behavior, privacy boundaries, and the current version of the consent notice. Consent must be an explicit action, not prechecked by visiting callback.
4. **Commit:** Within one transaction verify the active Clear session, fresh phone proof, decrypted current OAuth session, unexpired unconsumed challenge, unchanged consent version, eligibility policy and unique one-to-one constraints. Atomically consume nonce and write link plus audit event. If either identity has an existing different active link, return a generic conflict and offer a deliberate recovery/relink flow.
5. **Return:** Show a private connection status, never a public profile assertion without independent ownership validation. No PDS write, outbound sync or Clear runner execution is triggered.
6. **Failure recovery:** Keep the two accounts separate if any step fails; rate-limit repeated attempts and avoid leaking whether someone else's DID or phone is already linked.

**Cross-device linking:** Bind the ceremony to the initiating Clear browser session by default. Cross-device QR or deep-link linking requires a separate out-of-band challenge design and should not be silently added.

## 5. Consent, unlinking, deletion and privacy

- Consent language: “Connect my verified AT Protocol identity to this private Clear account” with clear explanations that linked DID is used for account association, is not proof of a currently held handle, does **not** publish private Clear data, and may be unlinked. Record notice version and timestamp.
- **Unlink:** An authenticated Clear session with fresh re-authentication requests unlink, shows affected capabilities and explicitly confirms. For recovery when Clear access is lost, require a separate owner-approved high-assurance account recovery path; a social session alone must not reveal or authorize private Clear deletion.
- Use a transaction to mark link revoked and invalidate/expire link-scoped claims, pending installation permissions and any sessions that depended on the link. Revoke/delete OAuth tokens associated with the disconnected relationship where appropriate; define what happens if the OAuth login is also used as an independent social session. Avoid inadvertently deleting an unrelated social login session if it represents an independent social account, but **never retain connection-scoped OAuth tokens on disconnect**.
- After unlink, previously authorized Clear workflows must not remain implicitly available to the DID. Existing local follows and public social records should follow separately documented social delete/unfollow controls, not hidden Clear settings.
- Account deletion: verified Clear deletion initiates the engineering-default timeline **only after owner and counsel sign-off**; deactivate the link, revoke sessions, remove tokens and process private Clear resources under a documented procedure. Backups, external providers and public AT Protocol replication have distinct limitations. No universal or immediate PDS deletion promise.
- Limit access to the link table to security/account code, audit reads to authorized operators, and exclude DID/phone joins, clear messages, OAuth secrets, raw tokens, consent text and identity-link URLs from analytics, server logs and public APIs.

## 6. Threat model and required mitigations

| Attack / failure | Required mitigation or test |
|---|---|
| Someone connects a victim's DID using their phone session | Bind verified OAuth subject to challenge; refuse user-supplied DID; consent preview and explicit confirmation |
| Someone connects their own DID to someone else's Clear account | Fresh authenticated Clear session/phone re-challenge; no raw userId selection; CSRF and session fixation tests |
| OAuth callback login CSRF, state replay, DPoP/issuer manipulation | Indigo state/nonce validation; PKCE/DPoP; exact redirect/origin allowlists; one-time consumed nonce; real hosted PDS integration tests |
| Handle rename, stolen handle, DID rotation or impersonation | DID is stable join key; separately resolve DID document/PDS handle with expiration, cache policy and DID change review; no verification badge from local matching alone |
| Concurrent link/replace requests, retry after timeout | Transactional partial unique indexes; single-use challenge; idempotent safe result without private identity disclosure |
| Session theft, OAuth token exfiltration or stale revoked grant | HttpOnly/Secure/SameSite cookies; encrypted OAuth store, fixed origin, idle and absolute session limits, revocation/rotation, server-side freshness checks |
| CSRF and API abuse, phone challenge brute force | Exact Origin and Fetch Metadata checks; rate-limit by user/session/IP with trusted-proxy policy and global limits; challenge attempt bounds; privacy-safe logs |
| Link creates unauthorized access to Clear runs, alpha or commerce | Separate Clear authentication on every private route; explicit scoped authorizations and alpha eligibility; link is metadata only |
| User unlinks yet access continues via cached token/job | Revoke active link authorization and sessions, invalidate caches/queued jobs, check link status on each privileged operation |
| Private Clear phone, tasks or run payload leak to public AT Protocol | No PDS writes from link, allowlisted public projections only, negative disclosure tests, no implicit synchronization |
| ATProto services or phone provider unavailable | No partial active link; maintain separate account access; audited and retry-safe recovery without privileges widening |
| Deletion vs legal hold / backup replication | Owner-reviewed category-specific retention, documented exceptions and backup policy; no unapproved time guarantees |

## 7. Code tasks unlocked, ordered PR plan

1. **Owner decisions and real-PDS validation gate:** Approve Clear session identity assurance, DID-to-user cardinality/recovery, explicit linking consent, public disclosure and unlink behavior; run hosted-PDS OAuth tests before the flag can be enabled.
2. **Database + tests:** Add `identity_links`, short-lived `identity_link_challenges`, audit events, unique active-link constraints and migration/race tests in `internal/database/schema.go` and dedicated `internal/identitylink` package. Keep private FK references internal.
3. **Real Clear authenticated sessions:** Extend `internal/beta`, `internal/security` and `internal/api` to create/revoke an end-user Clear session after fresh activation/re-authentication. Introduce session middleware bound to `users.id`. Do not use `ADMIN_TOKEN` or a phone number submitted in JSON as an end-user credential.
4. **Link UI/API (flag off):** Introduce `IDENTITY_LINKING_ENABLED=false` (proposed), `POST /api/me/identity-link/start`, `GET /api/me/identity-link/preview`, `POST /api/me/identity-link/confirm`, `DELETE /api/me/identity-link` backed by both validated identities and explicit consent. Add frontend private account connection status and re-auth prompts.
5. **OAuth binding:** Integrate the server-side challenge with Indigo `internal/atproto/oauth_handler.go` and `internal/atproto/social_session.go`. Do not leak nonce or let ordinary OAuth login create an implicit private link. Keep current `ATPROTO_OAUTH_ENABLED` off until real PDS approval.
6. **Authorization and Clear install integration:** Resolve `social DID -> active, authenticated Clear user ID` only after a verified link and live Clear session, then carefully implement installation activation for `protocol_install_requests` (M5) and private-alpha `alpha.Gate.Eligible` (M6). No execution until runner/security design approvals.
7. **Unlink, revocation and data minimization:** Delete connection-scoped OAuth tokens, revoke dependent sessions/capabilities, mark links inactive, handle queued approvals and document deletion/backups. Never auto-delete independent social content.
8. **Hardening CI and staging:** Write deterministic tests for forged DID/user ID, missing/expired cookies, stale consent, callback/state replay, concurrent link, two Clear accounts claiming one DID, unlink/relink, session revocation, CSRF, throttling and PDS disconnect. Owner runs real hosted-PDS and isolated staging smoke with redacted evidence.

## 8. Open questions for owner and counsel

- Must every link be strictly one DID ↔ one Clear account, and may a user change DID through a verified recovery ceremony?
- Is phone re-challenge required on every link and unlink, or may a recently authenticated Clear session with a maximum age suffice?
- When a user disconnects OAuth, should an independent social login remain intact while the Clear link and connection-scoped tokens are always removed?
- If a person loses a DID, what evidence is sufficient for Clear recovery without transferring someone else's private data?
- Can users retain a social profile when deleting private Clear, and what happens to locally hosted follow/creator content?
- How should AT Protocol handle rotation, DID document changes and suspended/deleted PDS accounts affect the displayed connection?
- What is the final versioned consent copy, retention/deletion schedule, contact for privacy requests, backup exception policy and legal jurisdiction?
- Are Clear users allowed multiple approved social DIDs later? What safeguards are needed for agents and delegated identities?
- Which actual private Clear actions will a link authorize, if any, and which always require fresh, individually scoped approval?

**Exit criteria:** Owner approves policy and counsel approves necessary disclosures; real hosted-PDS OAuth verification succeeds; Clear user sessions exist; link/unlink implementation passes abuse, authz, replay and deletion tests; no private data escapes to AT Protocol; staging smoke demonstrates revoked links cannot access any Clear resource. **Until then, this document authorizes no feature flag changes, deployment or external connections.**
