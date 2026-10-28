# Audit checkpoint 3: protocol creation, publication, and use

Date: October 27, 2026. Revision: `12390da9504ae91821851d027556968268fa0290` (`dev`). Runtime source unchanged. Temporary local probes were removed; all identities, guides, and takedowns were synthetic.

## Product result

The backend can store versioned drafts, transition them to private/unlisted/public, list metadata, pin an install request to a spec digest, and approve requests. Optional inert `guide` steps allow readable instructions. These are useful foundations, but the social protocol-sharing loop is not complete.

- No Creator Studio or protocol-detail/configuration/export screen is registered in the SPA.
- Public discovery returns title/summary/license/version metadata, not a readable workflow body. The manifest endpoint returns permissions and digests, not guide text. The current frontend does not consume protocol discovery.
- Draft schema supports steps, permissions, remixOf, and attribution. It rejects inputs/defaults, triggers, and other configurable workflow fields. There is no install-specific configuration here.
- General request approval explicitly returns installed=false and executed=false; it changes a request state, not a user's installed collection.
- The narrow self-serve guide endpoint creates an installation for an authenticated private Clear user, with alpha gating when enabled. It does not return guide text, export to another chatbot, or execute the guide.
- The fake runner only supports local.set/local.assert; it is not a guide interpreter or chatbot integration.

Therefore an ordinary creator and reader cannot yet complete publish → discover → configure → use through the website, even with all relevant flags enabled. This is a scope/implementation gap, not a reason to remove the existing permission checks.

## Reproduced findings

### F09 — P1: paid-enforcement discovery drops the takedown filter

`internal/protocols/discovery.go:23` excludes hidden takedowns. The replacement query at line 24, selected by PaidEnforcement, excludes paid listings but omits that takedown condition. A free taken-down protocol reappears when paid enforcement is enabled.

Probe: created/published a valid guide through DraftService; requested and approved its hidden takedown through Takedowns. Reader.Discover omitted it with default options, then returned it with PaidEnforcement=true. No paid listing was needed.

Fix direction: compose independent visibility, takedown, hold, and entitlement predicates rather than replacing complete queries. Test combinations of flags, including hidden free protocols with paid enforcement on.

### F10 — P1: general install flow ignores hidden takedowns

`internal/protocols/install_requests.go:35` checks only publication status. Preview, Request, and Decide use this manifest lookup, without consulting creator_takedown_requests. Hiding via Takedowns does not change the protocol's published status.

Probe: the same hidden guide still passed Preview, Request, and approving Decide. These calls are directly used by the manifest/install-request/administrator-decision endpoints. This is a service-level reproduction, not a full authenticated HTTP/browser reproduction. Approval still does not execute the protocol.

Fix direction: centralize publication accessibility checks, enforce them when previewing/requesting/approving/using, and recheck them within the relevant transaction. The self-serve guide path already checks hidden takedowns, illustrating the inconsistency. Review social moderation holds at the same boundaries.

### F11 — P2: publish validation accepts a guide that install validation rejects

`internal/protocols/draft_validation.go:21-35` permits absent permissions, but install_requests.go:41 requires a non-nil array. The existing guide validation test explicitly treats a guide without permissions as valid.

Probe: created and publicly published `{"steps":[{"operation":"guide","text":"List next actions."}]}` with GuideStepsEnabled; Preview returned ErrInstallUnavailable. Adding `"permissions":[]` avoids this specific failure.

Fix direction: adopt one canonical schema and normalize optional empty fields, or consistently require an explicit array before accepting a draft. Exercise the entire lifecycle from validated input instead of separately hand-seeding downstream fixtures.

### F12 — P2: permission-difference flag breaks readable guides

`internal/protocols/install_requests.go:38` calls AnalyzeManifest when PermissionDiffEnabled. `internal/protocols/manifest_analysis.go` only accepts local.set/local.assert fields; it rejects guide/text and remix metadata even though creator validation accepts those under their supported rules.

Probe: a published guide with explicit empty permissions previews normally but fails with PermissionDiffEnabled=true.

Fix direction: separate permission analysis from executable-plan validation; keep guide steps inert and use the same supported schema across authoring, publication, manifests, and delivery. Verify mixed flag configurations.

## Additional static integration findings

### F13 — P2: creator mutations do not explicitly push Turso sync

DraftService.Create, PublishDraft, and Takedowns.Request/Decide operate on *sql.DB only. The reviewed API wrappers do not invoke Store.Push after these changes. In contrast, install requests and self-serve installs explicitly push. Startup currently passes only DB to the creator services.

When remote sync is configured, success from these endpoints does not itself establish that their writes have been pushed remotely; a later unrelated push may include them. Local commits remain real. No remote data-loss or sync fault test was performed.

Fix direction: define one consistent commit/sync acknowledgment contract and wire it through creator writes, including safe retry semantics if local commit succeeds and remote push fails. Review all write services for the same pattern.

## Existing behavior worth preserving

- Drafts are version-keyed and owner-bound, with bounded JSON/step sizes and constrained operations.
- Publishing repeats validation and writes visibility/audit changes transactionally; cross-owner publishing is rejected.
- Unlisted publications stay out of normal discovery; private drafts stay unpublished.
- Remix publication checks source existence and conservative license compatibility, and records lineage.
- Install acknowledgments pin a digest and version; changed specs invalidate stale acknowledgments/decisions.
- Self-serve guides reject executable steps, permissions, nonpublic and hidden publications, and scope installations to the authenticated private user.

## Open issues for later verification

- The publication flow only permits increasing visibility; there is no creator self-service return to private/unlisted. The separate takedown path needs an administrator decision. Decide what publish/unpublish controls creators should have.
- Guide installs store publication ID as protocol_key, while the older Clear Request path matches installations by slug. Trace intended relationships before connecting guide installation to run requests; slug alone also needs creator scoping.
- General manifest lookup does not check creator visibility/takedown metadata or creator moderation holds. Paid access differs between preview and request; inspect paid-user journey during marketplace review.
- Private → public requires two explicit API transitions. The UI must explain or handle this when authoring becomes available.
- ValidateDraftSpec accepts an empty steps array. Decide whether empty drafts are allowed and whether publication must require at least one useful step.
- Publication CreatedAt sorting uses RFC3339Nano text; mixed precision and version ordering require separate pagination/discovery review.

## Verification

Existing protocols, API, and Clear package tests passed. A temporary lifecycle probe exercised actual Create → private → public → takedown → discovery/manifest/request/approval services and confirmed F09-F12. It also confirmed configurable inputs are currently rejected. No external chatbot compatibility or real execution was tested or claimed.

Next checkpoint: private identity linking and execution authorization, including the two installation representations, followed by marketplace and deferred agent infrastructure. Core product fixes should prioritize a consistent protocol schema, publication accessibility, and one readable/configurable delivery path.
