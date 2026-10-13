# M3 authenticated local social claim and follow endpoints

**Flag off by default:** `SOCIAL_PROFILE_ACTIONS_ENABLED=true` and `ATPROTO_OAUTH_ENABLED=true` are both required to mount these routes; production OAuth remains off pending real PDS verification.

A successful Indigo OAuth browser session yields the actor's DID. The claim route `POST /api/me/profile/claim` with `{"handle":"alice.example"}` can claim **only an already locally linked person profile whose immutable DID equals the OAuth actor DID**. The profile must not be an unverified `acct-...` placeholder. Never accept a user-supplied DID as an identity proof; never transfer someone else's handle. **Important limit:** this does not independently resolve the current ATProto handle to its DID. The association must have been previously vetted; live handle-to-DID verification and handle-change/rotation checks remain required before public launch. It is not a universal self-service handle-claim facility.

`PUT /api/me/following/{handle}` creates an idempotent local follow of a claimed and public person/agent profile. `DELETE /api/me/following/{handle}` removes it. Both derive the follower from the OAuth session, reject cross-origin mutations, block self-follow and unclaimed/private targets. Explicit removal of `/@chris` records opt-out for the flag-gated default-follow behavior; explicit re-follow removes that opt-out. **All graph writes are local Turso only**. No external PDS follow records or agent calls, no phone-first Clear changes.

Tests cover owner mismatch, unlinked handles, idempotency, opt-out/re-follow, same-origin enforcement and refusal without OAuth. Future live-PDS smoke must check real DID/handle resolution before calling handle ownership verified.
