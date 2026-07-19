# Point-in-Time Recovery

This document defines the supported point-in-time recovery scope for a local-only IronPage Vault deployment.

## Local-only constraint

Recovery must not require cloud services or external network access. Inputs are local artifacts, generated installation volumes, and locally retained database/archive material.

## Supported scope

The implemented recovery boundary is strict logical backup and restore:

- PostgreSQL custom-format dump;
- PDF-storage tar snapshot;
- manifest containing both artifact paths; and
- restore that succeeds only after both artifacts are applied.

Automated WAL archiving, scheduled physical base backups, and timestamp-target physical restore are not implemented or claimed.

## Recovery inputs

A complete supported recovery point contains:

- PostgreSQL custom-format dump;
- PDF storage snapshot;
- the retained installation `.env` or an equivalent protected configuration snapshot; and
- backup manifest or backup metadata record.

When an operator independently enables physical PostgreSQL PITR, its base-backup, WAL-retention, timeline, and timestamp procedures remain an operator-owned extension outside the implemented product evidence.

## Database and filesystem consistency

PostgreSQL stores each PDF version path and hash. The generated filesystem target stores the binary. Both must be restored to the same recovery boundary:

```text
document_versions.file_path -> restored local PDF file
```

Restoring database metadata without the corresponding file snapshot is not supported.

## Recovery procedure

1. Stop application writes.
2. Retain the installation `.env` and identify its generated storage targets.
3. Select one manifest and both artifact paths.
4. Restore PostgreSQL from the custom-format dump.
5. Restore PDF storage from the corresponding tar snapshot.
6. Start the service with the retained installation configuration.
7. Verify the generated health URL.
8. Verify representative metadata, version downloads, audit records, workflow state, notifications, and backup metadata.

## Evidence boundary

Strict backup/restore source paths and contracts can be inspected statically. A successful recovery claim requires an executed restore artifact tied to the exact tested revision and recovery inputs. No documentation or static report may be treated as proof of automated physical PITR.
