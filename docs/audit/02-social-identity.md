# Audit checkpoint 2: database integrity and social onboarding

Date: October 27, 2026. Revision: `12390da9504ae91821851d027556968268fa0290` (`dev`).

Read-only runtime audit; no application source changes. Temporary tests used only synthetic identities and disposable local databases, and were removed after execution. No real PDS login, external account creation, or deployment was attempted.

## What the current journey actually does

1. OAuth login is an identifier form served by the API. It connects an existing AT Protocol identity; it does not create a hosted-PDS account for a new visitor.
2. Indigo handles authorization, and callback creates a 24-hour local browser session and redirects to the frontend home page.
3. If the first-follow hook is enabled, it creates a private `acct-...` placeholder and follows the existing claimed/public local `chris` profile. If Chris is absent, login succeeds without the follow.
4. The separately flagged social-account endpoint can also create the private placeholder. Neither path assigns the user's actual handle, bio, or public identity.
5. The claim endpoint can only publish a profile whose non-placeholder handle is already present in the database under the exact authenticated DID. It cannot convert a fresh placeholder into the user's chosen profile or transfer a celebrity reservation to a newly verified owner.
6. The public profile frontend shows identity/bio and claimed status, but has no follow, claim, edit, or protocol-catalog controls. Those user journeys cannot currently be completed through that screen.

Public profiles/follows remain local Turso records. OAuth authentication is not evidence that the graph or protocols are published into users' PDS repositories.

## Confirmed findings

### F05 — P1: declared foreign keys are not enforced

`internal/database/database.go:26-80` opens connections without enabling foreign-key enforcement. Probes against both the in-memory database and a fresh file-backed database reported `PRAGMA foreign_keys=0` after migration. Both accepted a social_follows row referencing two nonexistent social profiles, despite foreign keys declared at schema.go:320.

Impact: the database does not provide the referential-integrity guarantees suggested by REFERENCES declarations. Service checks may prevent some invalid writes, but cannot substitute for database constraints across every lifecycle, maintenance action, or future caller. This is not evidence of an unauthenticated HTTP exploit.

Recommended fix: configure enforcement for every connection using a supported driver mechanism; verify pool/reopen behavior, parent-before-child creation and deletion order, inspect existing data for orphans, and run the full suite before enabling it on an existing database. A one-time PRAGMA on an arbitrary pooled connection is insufficient.

### F06 — core product gap: social signup cannot reach a public profile

`internal/social/accounts.go:24-38` and default_follow.go:31-50 create private placeholder handles. `internal/social/actions.go:24-36` only claims an already assigned non-placeholder handle under the same DID. There is no production handle-assignment/update step in these reviewed services or their API routes.

Probe: EnsureAuthenticated for a synthetic valid-shaped DID succeeded, then ClaimExisting for `alice.test` returned `handle has not been independently linked to this DID`.

The existing actions test starts by manually inserting Alice with a real-looking handle; it proves ownership protection once prelinked, not new-user onboarding. Existing API tests establish rejection without authentication, not a successful browser signup-to-profile journey.

Recommended work: a supported existing-identity sign-in path plus a new-identity signup handoff; verified DID/handle resolution; explicit profile-publication choice; profile editing; browser UI. Decide separately whether Cleartify profile slugs may differ from AT Protocol handles. Do not weaken the existing ownership check to make onboarding appear to work.

### F07 — P2/incomplete feature: stored blocks have no social effect

`internal/social/moderation.go:31-39` records/removes block rows. Neither Actions.SetFollow, DefaultFollow.OnVerifiedLogin, nor Reader.PublicByHandle consults them. Repository search found no production readers of social_blocks outside its storage operations.

Probe: Alice blocked Chris and then successfully followed Chris. Blocking also does not remove an existing follow. This is consistent with docs/m3-social-moderation-defaults.md explicitly deferring filtering; it is an unfinished product capability, not a newly discovered contradiction of that document.

Recommended work: define block semantics first (at minimum follow creation, existing relationships, default re-follow, and eventual interaction/feed behavior), then enforce them transactionally across applicable paths. Public anonymous access to inherently public content is a distinct question; do not promise invisibility from a local block.

### F08 — P2: hiding a target prevents unfollowing it

`internal/social/actions.go:49` requires public_visible=1 and claimed=1 for both follow and unfollow. default_follow.go:60 applies the same restriction to the dedicated Chris removal path.

File-backed probe: Alice followed Chris; the test made Chris private; Alice's unfollow returned `public profile not found` and the relationship remained. The dedicated default-follow removal also rejects hidden Chris. A user cannot record an opt-out through these routes while the target is hidden.

Recommended fix: require public/claimed status when adding a relationship; permit authenticated users to remove their existing relationships and persist default-follow opt-out even when the target becomes private. Stable DID-based relationship identifiers avoid ambiguity after handle changes.

## Behaviors supported by current tests/review

- Public profile lookup normalizes handles and hides private profiles.
- Claims reject another DID's profile and unlinked handles.
- Repeated follows do not duplicate rows; self-follow is rejected.
- First follow is local, requires an existing claimed/public Chris profile, and avoids following oneself.
- Explicit opt-out survives repeated logins; explicit re-follow clears it.
- Social registration/mutation endpoints reject absent sessions and invalid origins; disabled routes are absent.
- OAuth request state is browser-nonce-bound and consumed transactionally before token exchange; replay and wrong-browser tests exist.
- OAuth payload encryption uses AES-GCM with random nonces and record-bound associated data; ciphertext tampering/context substitution tests exist.

These are local guarantees, not live provider or browser acceptance.

## Other identity observations to follow up

- Logout ignores local deletion/push and provider logout errors (oauth_handler.go:137-145). It may return 204 despite incomplete revocation; fault-injection verification remains pending.
- The dedicated Chris unfollow uses LocalAuthenticatedDID, which only joins session rows, while other social mutations use SocialSessionDID, which additionally decrypts/validates the OAuth payload. Consolidate or document the different assurance levels.
- OAuth POST origin checks require the API public origin. A frontend-hosted login/logout button on a different origin cannot simply POST directly; the current API-hosted form matches the check. Cross-origin UX requires deliberate design and browser validation.
- SocialSessionDID checks local browser expiry and decryptability, not fresh provider-side revocation. Define expected logout/revocation behavior before launch.
- Reserved creator claims need a separate proof/transfer ceremony and provenance model. The schema keys profiles by a non-null DID; a placeholder reservation cannot be treated as a verified creator DID. Do not infer Tiago Forte/Huberman seeds from design documents or test fixtures.
- Existing bootstrap remains CREATE IF NOT EXISTS without tracked upgrade versions. Full upgrade/restore behavior is still pending review.

## Verification and next checkpoint

Existing database, social, AT Protocol, and API package tests passed. Two removed probe tests confirmed F05-F08 and the placeholder claim boundary using synthetic local data. Full hosted-PDS signup, session cookies across origins, and real handle ownership remain unverified.

Next: protocol definitions, guide/configuration support, draft-to-public lifecycle, profile discovery and install/export paths. Determine whether an ordinary creator can publish a useful runbook and another user can actually configure/use it. Fixes remain separate from this inventory/review pass.
