# Audit checkpoint 4: private identity and execution authorization

Date: October 27, 2026. Revision: `12390da9504ae91821851d027556968268fa0290` (`dev`). No runtime source changes or external provider actions. Temporary probes were removed after running against disposable local databases.

## Identity and execution paths

There are three separate representations, not one complete execution system:

1. Legacy Clear run requests reference protocol_installations by slug/version, require administrator access at the HTTP boundary, and only transition approval status. They never execute anything.
2. Social install requests reference a publication ID plus version/spec/manifest digests. Administrator approval changes the request status. The fake-run endpoint consumes one approved request and records a local simulation.
3. Self-serve guides write protocol_installations using publication ID as protocol_key, rather than the legacy slug convention. These installs have no corresponding guide-delivery or execution path yet.

Fake runs require both a private Clear principal and an independently validated local social OAuth session, with an active link matching both identities. Alpha membership is additionally checked when its wrapper is enabled. The handler validates the request actor, version/spec digest, publication visibility, and optional fake paid entitlement, then reserves a unique run before executing local.set/local.assert in memory.

## Reproduced findings

### F14 — P1: legacy run-request authorization is not creator-bound

`internal/clear/run_requests.go:20` matches user installations to the requested publication using only slug and version. It does not bind the installation to publication ID or creator DID. Slugs are unique only within a creator/version tuple, not globally.

Probe: inserted Alice's and Bob's separate publications with slug `weekly`, version 1; created one user's installation keyed `weekly`; Service.Request accepted Bob's publication too. This demonstrates ambiguous authorization by design of the legacy representation, without needing to change any publication content.

Scope: the exposed legacy route is administrator-authorized and creates a pending request, not real execution. This is not a claim that an unauthenticated user can execute arbitrary code. It is an unsafe authorization invariant to preserve when connecting a real runtime.

Fix direction: bind installs and approvals to an exact publication/version identity, and migrate legacy slug-based rows with explicit creator resolution. Ambiguous historical rows must require review rather than silently selecting a creator.

### F15 — P2: legacy request approval ignores later installation disablement

`internal/clear/run_requests.go:27-32` checks only request ID, user, and pending status. It does not recheck enabled installation, publication availability, or its content/version digest when approving.

Probe: created a valid request, disabled its installation, then approved it successfully. Approval is only a recorded state today. A future execution consumer must not treat it as proof of current authority.

Fix direction: revalidate approval prerequisites and pin exact content at the authoritative transaction boundary; revalidate again before any consequential execution. Test uninstall/disable, withdrawal/takedown, and spec changes between request, approval, and use.

### F16 — P2: guide installs are incompatible with the legacy run-request lookup

The self-serve endpoint writes protocol_key=publication ID (`internal/api/self_serve_guide_installs.go:60`); legacy Service.Request expects protocol_key=slug. A publication-ID installation was rejected as `protocol is not published` even though the publication existed and was published.

This is a connection gap between existing APIs, not evidence that guides should run through the local executable simulator. Select one canonical installation identity, then define a separate guide consumption/delivery action with accurate errors.

### F17 — P1: taken-down protocols can complete fake execution

`internal/api/fake_install_run.go:38-50` rechecks published status and nonprivate visibility but never checks creator_takedown_requests. A hidden takedown retains both published status and original visibility.

Probe: prepared a valid approved install, encrypted synthetic OAuth session and matching active identity link; marked the publication hidden; called startFakeInstallRun with the authenticated synthetic social cookie and a trusted test Clear-principal resolver. The handler returned HTTP 201 with `status=completed`, `sandboxOnly=true`, and `realExecution=false`.

Boundary: the probe invoked the handler directly, after supplying a trusted private principal; it did not bypass or test the outer alpha/origin middleware, and used no real identity provider. The defect is the missing takedown check after valid authentication. Only in-memory fake operations executed.

Fix direction: apply the same centralized, transactional publication-access predicate proposed in F10 immediately before reservation/use. Test takedown and moderation holds after approval; ensure denied runs create no executable reservation or side effect.

## Identity controls supported by tests/review

- Private session tokens are random, stored as keyed digests, and associated with active users; expiry, revocation, and fresh-issuance limits have tests.
- Linking requires fresh private authentication, an independently derived social DID, a browser challenge bound to both, and explicit versioned consent.
- Link challenge consumption, one-to-one active link uniqueness, and link audit insertion share a transaction. Wrong principals, expired challenges, replay, and duplicate DID/user linking are tested.
- Disconnect/recovery atomically revoke active links and private sessions, cancel pending link challenges, and write an event. Recovery permits a private user to disconnect a lost social identity without supplying that identity's live cookie.
- Alpha membership tests verify activated/admitted users, revocation, and rejection of spoofed user IDs and missing dependencies.
- Fake-run reservation is unique per install request; existing crash recovery fails interrupted reservations with unknown outcome instead of blindly replaying them.

These tests do not establish real provider reauthentication, deployed durability, or a complete social-user run journey.

## Follow-up observations

- SocialSessionDID accepts an unexpired local session up to 24 hours old. Link records nevertheless set authn_social_at to confirmation time, not the original OAuth authentication time (identity_link.go:83). That field means local revalidation today, not proof of fresh provider authentication. Keep that distinction explicit and define required assurance before enabling consequential linkage.
- Simple unlink/recovery revokes the identity relationship but does not delete social OAuth credentials or invalidate social install requests. Fake runs recheck the active link, so they are denied while unlinked. Decide which grants/approvals should survive relinking, and distinguish standalone social login from connection-scoped credentials.
- Session authentication extends server-side idle expiry, but these reviewed response paths do not refresh the browser cookie lifetime. The browser may discard a cookie 30 days after issuance despite continuing activity. Verify intended session UX during browser review.
- Session expiry SQL uses RFC3339Nano text comparison whereas initial expiry validation parses time. Mixed precision near a same-second boundary needs focused verification; no new expiry-bypass defect is asserted here.
- LocalSandbox's quota executor and Scheduler are separate from the HTTP fake runner, which constructs runner.New(true). Do not imply the API is using the optional scheduler/aggregate byte quota merely because those libraries exist.
- Scheduler is manually ticked and in-memory, with bounded total job storage; completed jobs remain retained. No durable production scheduling is established.

## Verification and next checkpoint

Existing identitylink, clearsession, alpha, clear, and API tests passed. Two temporary probes confirmed cross-creator slug ambiguity, approval after disablement, publication-ID/slug incompatibility, and fake execution after takedown. No production data or source was changed.

Next: marketplace entitlements, refunds, payment simulations, and accounting wiring; then deferred agent grants/interop and remaining frontend/onboarding/operations paths. Prioritize canonical publication/install identity and consistent revocation/access checks before attaching a real executor.
