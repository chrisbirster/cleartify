# M4 creator publishing workflow

**DEFAULT (Oct 13 2026), owner can change.** New `CREATOR_PUBLISHING_ENABLED=false` gate, additionally requires existing creator drafts and AT Protocol OAuth enabled. No feature is activated, no external PDS contact or launch approval.

Authenticated OAuth DID owners alone may POST `/api/me/protocol-drafts/{id}/visibility` with `visibility` and a non-empty `provenance` note. Transitions are immutable and monotonic: `draft -> private -> unlisted/public`, and `unlisted -> public`. The underlying protocol spec and edition/version are never rewritten by the transition. Every successful change writes a transactionally consistent actor-DID audit record. Invalid and duplicate transitions fail closed.

Supported internal, non-legally-approved license allowlist: CC0-1.0, CC-BY-4.0, MIT, Apache-2.0. Remix references must resolve to a published, nonprivate upstream publication, have attribution and meet conservative compatibility: CC0 upstream or matching CC-BY-4.0 downstream. A bare upstream ID and author's note do not independently establish license ownership, consent or provenance, and cannot replace counsel/owner review. Unsupported or absent sources fail closed.

Private records remain in draft status, never enter public discovery. Unlisted and public records become published; unlisted records remain absent from list discovery (direct version lookup can access them). Public listings use the existing discovery endpoint. The existing global published-protocol consumers are not equivalent to per-user permission; production install and moderation boundaries remain open.

**Remaining:** approval/rights verification, takedown/moderation integration, full remix lineage and license compatibility, rights-holder attestations, 10 approved seed protocols, real PDS smoke and staging. Public legal wording and launch are owner-only.
