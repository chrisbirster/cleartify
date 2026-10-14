# M5 manifest static analysis and version permissions (Round 4)

**Default OFF:** `MANIFEST_PERMISSION_DIFF_ENABLED=false`. When explicitly enabled, install preview performs strict offline static analysis of declarative manifest permissions and steps; no arbitrary code, process, filesystem, environment, secrets or network access is executed.

Only `local.read`, `local.write.preview`, and `notify.preview` permissions and the `local.set` / `local.assert` declarative steps are permitted. Unknown operations, unexpected fields, paths in step keys, unsupported privileges and over-limit plans fail closed. This is **not** a production OS sandbox or proof of safe untrusted code.

For a newer published version with the same creator DID and slug, the preview includes sorted `permissionChanges.added`, `removed`, and `changed` fields relative to the closest earlier version. The existing canonical manifest digest includes the publication ID, version, permissions and spec digest. **A request acknowledging an older version's digest fails** with `ErrManifestChanged`; a new preview/explicit acknowledgement is needed even when the newer version only changes permissions. The admin approval step also rechecks pinned version and spec digest. No automatic upgrade or silent carryover of approval is introduced.

The diff is advisory preview metadata, not an independent authorization token. The HTTP/UI must show the new preview and require explicit consent before sending the matching digest; no new frontend prompt is implemented here. Re-prompt is enforced at the server acknowledgement boundary. An old version cannot be treated as permission to install a changed new version.

Tests cover dangerous manifest field and operation rejection, permission add/remove differences, stale acknowledgements across versions, explicit fresh approval, flag-off behavior, and sandbox rejection of shell/filesystem/network/env/secret operations. No real sandbox execution, staging, deployments or external services are added.
