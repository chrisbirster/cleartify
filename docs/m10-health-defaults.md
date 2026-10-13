# M10 observability and release-source verification

**DEFAULT (Oct 13 2026), owner can change.** Internal engineering defaults only.

- Existing `GET /healthz` remains unchanged. Monitoring hooks are an **unwired opt-in library**, disabled unless a caller explicitly uses `monitoring.Wrap(..., true, counter)`. No new network routes, metrics exporters, alerts, providers, URLs, credentials, or infrastructure changes.
- In-memory request and HTTP 5xx counts are process-local, non-durable, and never carry URL paths, identities, query strings, tokens, or payloads. No logging of private user data.
- `scripts/check-release-sha.sh` is a **local checkout consistency** check, not artifact signing, provenance attestation, deployment approval, or protection against malicious workflow changes. CI validates the full commit SHA against checked-out HEAD in both existing jobs, no new job or deployment action.
- Before launch: owner must approve metric scope, retention, authentic readiness probes, alert destinations and thresholds, log/trace redaction, backup recovery objectives, artifact attestation and actual deployment-specific rollback steps.
