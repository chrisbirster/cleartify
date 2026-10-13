# M7 agent registration, slice 1

Set `AGENT_REGISTRATION_ENABLED=true` explicitly to expose `POST /api/admin/agents`; default is **off**. This administrative-only route requires the existing `ADMIN_TOKEN` bearer. It accepts `ownerDid`, `displayName` and `description` and stores a local **unverified** agent registration linked to an already public claimed person profile. This is **not** proof of ownership of an external AI assistant. No provider SDK, API call, token handling, remote agent verification or autonomous work is performed.

When enabled, do not publicly mark a listing as connected or verified. Admin registration is a temporary operator workflow: owner-authorized self-service registration must wait for fully verified OAuth identity linking, session permissions, revocation, moderation and appropriate rate limits. Do not use admin bearer tokens in the browser. The route is intentionally unavailable when the flag is off.

Manual staging checklist: flag off → 404; flag on with no admin token → 503; wrong bearer → 401; invalid or unclaimed owner → 400; valid claimed owner → 201 and unverified record; no external network requests. Staging must use only disposable test owner identities.
