# M6 Clear private-alpha session enforcement

**DEFAULT (Oct 13 2026), owner can change.** Internal engineering defaults, not an external promise.

- `PRIVATE_ALPHA_ROUTES_ENABLED` remains false by default. Only when **both** it and `CLEAR_SESSIONS_ENABLED` are enabled is the private Clear session status route wrapped by the alpha gate; alpha eligibility status is also protected by that same gate.
- The gate uses only server-authenticated, unrevoked Clear cookies from the session validator; query, path and request body user IDs never establish a principal. Eligible membership additionally requires an active user, activated beta invite, and non-revoked alpha membership.
- Missing validator or eligibility service fails closed (503), absent/expired/revoked session returns 401, revoked/ineligible membership returns 403. Never leak session tokens or private membership data to other users.
- Logout remains available even when membership is revoked so users can clear sessions. Existing admin-only run-request APIs are **not** reclassified as private user routes in this change; admin authorization and their own flags remain separate.
- No deployment, production toggles, secret rotation, payment or real-PDS calls. Future owner review: extend protection to any additional **new** private Clear user routes, confirm CSRF/cookie behavior and atomic revocation semantics, and independently verify external DID links.
