# Protocol permission preview policy

**DEFAULT (Oct 13 2026), owner can change.** This is an engineering-only allowlist, not a final consequential-action permission policy.

- `STRICT_PROTOCOL_PERMISSIONS_ENABLED=false` by default; when enabled, this adds fail-closed checks to existing M5 manifest preview/install-request validation.
- Only `local.read`, `local.write.preview`, and `notify.preview` are accepted. Unknown, duplicate, network, payment, message-send, host-execution and unbounded manifests are refused.
- **No execution permission is granted.** These labels are local placeholders only; no runner, PDS writes, email send, provider call, or installation activation is added.
- Owner to approve final permission taxonomy, actual sandbox boundary, granular consent and consequential action policy prior to enabling runtime.
