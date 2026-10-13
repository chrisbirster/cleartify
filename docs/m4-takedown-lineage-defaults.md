# M4 local creator takedown and remix ancestry (engineering defaults)

**Default OFF**: `CREATOR_TAKEDOWNS_ENABLED=false`. The routes are not registered unless explicitly enabled alongside social OAuth. All reason values are internal *keys* (`rights_review`, `identity_review`, `other_review`); no public policy, legal assessment or final wording is created.

Owner DID from an authenticated social session can submit `POST /api/me/protocols/{id}/takedown-requests` with a reasonKey. Pending requests are recorded with an immutable audit event, and duplicate open requests are rejected. An existing admin bearer token is required for `POST /api/admin/creator-takedowns/{id}/decision` with `hidden`, `rejected`, or `restored`. Decisions are transactional; `hidden` suppresses public protocol discovery and version lookup without deleting content; `restored` removes the suppression while preserving audit history. A pending request alone does not hide content.

Validated remix publications store their direct parent and attribution text in `creator_remix_lineage` when first published privately. Authenticated `GET /api/me/protocols/{id}/lineage` follows bounded parent references and rejects cycles. Parentage remains after takedown and restoration. This is *technical lineage*, not independent copyright/licensing verification.

Tests cover creator-vs-other-DID ownership, repeated requests, hide/restore, durable audit row counts and remix lineage. Boundaries: no external notices, independent rights adjudication, moderation UI, takedown of installed copies, professional legal review, real PDS, network/deployment or keys. The flag does not authorize production enablement.
