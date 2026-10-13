# M10: rollback and backup recovery runbook (documentation only)

**DEFAULT (Oct 13 2026), owner can change.** Internal engineering procedure; NOT a deployment authorization, operational SLA, or public promise. Owner must review against actual deployed infrastructure.

## Precondition / stop conditions
Do not use this document to perform production changes automatically. Identify the deployment owner, current release Git SHA, last-known-good SHA, database engine and backup policy, runtime flags, and incident commander. Halt if an environment, credential, backup restore point or owner approval is missing. Do not paste secrets or user data in tickets or CI logs.

## Code rollback (owner-executed later)
1. Freeze releases and external write operations using an owner-approved procedure; capture read-only symptoms, timestamps, and commit SHA.
2. Compare failing release with the last-known-good immutable Git SHA and changes to schema/data migrations. Never assume a code rollback reverses migrations.
3. Confirm rollback compatibility, obtain explicit approval, and use the deployment platform's existing release rollback mechanism (not provided here).
4. Verify readiness, error rate, relevant request flows and authorization boundaries against the selected version. If verification fails, halt and escalate; do not repeatedly deploy.
5. Record timeline, version, accountable owner, and follow-up actions without secrets or personal data.

## Backup recovery (owner-executed later)
1. Check backup existence, integrity, timestamp, database version and restoration permissions with the owner before touching any environment.
2. Prefer a disposable isolated recovery environment for an encrypted restore rehearsal; validate schema, row counts, foreign keys, and critical data samples using redacted output.
3. Record recovery-point uncertainty and write-loss implications. Explicitly obtain owner sign-off before any destructive restore.
4. For a real incident, freeze writers, take a pre-restore snapshot, follow the database provider's documented point-in-time recovery procedure, and validate integrity before resuming writes.
5. Preserve incident evidence, audit access, evaluate replays/reconciliation, and obtain owner go-ahead before accepting traffic again.

## Open decisions
Actual RPO/RTO, backup retention, encryption/key custody, region, incident access, migrations, status-page copy, platform-specific commands, and recovery drill schedule need owner and counsel/security review. None is established here.
