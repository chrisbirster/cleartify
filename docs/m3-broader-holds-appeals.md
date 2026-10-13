# M3 broader hold enforcement and self-service appeals

**Default OFF**: `MODERATION_BROADER_ENABLED=false`. Turning it on requires the already flag-gated moderation runtime and a configured moderation DB. The existing narrower `MODERATION_ENFORCEMENT_ENABLED` remains independent.

When enabled, currently authenticated social profile claim/follow/unfollow, creator draft creation and creator publication-visibility changes deny held **actor DIDs** (HTTP 403); database failures fail closed (503). Catalog discovery excludes public protocols authored by actively held DIDs; version lookup also excludes them. A released hold automatically restores access/visibility without deleting the original hold or audit history. Non-held users' actions remain unchanged.

The two new authenticated and flag-gated routes are `GET /api/me/moderation/hold` (onHold, appealPending) and `POST /api/me/moderation/appeal` (reason, 5–1000 characters). The latter accepts the verified social session DID only, requires the configured Origin, and deduplicates pending appeals via existing DB uniqueness. The existing **admin-only** case listing and decision endpoint handles release with an atomic hold update and audit event. The old admin-submitted appeal remains distinct.

These are API primitives for a future client appeal form, **not a complete rendered frontend UX**. Other publishing, agent, commerce, social and private actions may still need policy-specific hold enforcement; no independent real-PDS claim or production moderation approval is implied. There are no keys, external endpoints, legal wording, deployments or workflow changes.
