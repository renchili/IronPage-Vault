# Point-in-Time Recovery

This document defines the supported point-in-time recovery scope for a local-only IronPage Vault deployment.

## Local-only constraint

Recovery must not require cloud services or external network access. Inputs are local artifacts, generated installation volumes, and locally retained database/archive material.

## Supported scope

The implemented recovery boundary is strict logical backup and restore:

- PostgreSQL custom-format dump;
- PDF-storage tar snapshot;
- manifest containing both artifact paths;
- an exclusive application mutation barrier across dump and tar; and
- restore that succeeds only after both artifacts are applied.

Admin-managed scheduled logical backup uses the same boundary. Its enable flag and interval are PostgreSQL configuration, so they are included in the dump and restored with the installation. Automated WAL archiving, scheduled physical base backups, and timestamp-target physical restore are not implemented or claimed.

## Recovery inputs

A complete supported recovery point contains:

- PostgreSQL custom-format dump;
- PDF storage snapshot;
- the retained installation `.env` or an equivalent protected configuration snapshot; and
- backup manifest or backup metadata record.

When an operator independently enables physical PostgreSQL PITR, its base-backup, WAL-retention, timeline, and timestamp procedures remain an operator-owned extension outside the implemented product evidence.

## Database and filesystem consistency

PostgreSQL stores version identity in `document_versions` and file path, digest, size, and parsed page count in `document_files`. The generated filesystem target stores the binary. Both must be restored to the same recovery boundary:

```text
document_versions.id -> document_files.version_id -> restored local PDF file
```

`redaction_confirmations` links proposals to source/result versions. `document_diffs` contains encrypted comparison results and version references. These entities, audit rows, notifications, and backup schedule configuration are restored through the same database dump.

Every supported application mutation acquires a shared advisory lock. Backup acquires the corresponding exclusive lock before `pg_dump` and keeps it through the filesystem tar, so the two artifacts are created while application state is write-quiescent. Restoring database metadata without the corresponding file snapshot is not supported.

## Recovery procedure

1. Retain the installation `.env` and identify its generated storage targets.
2. Select one manifest and both artifact paths.
3. Submit both artifacts to `POST /api/admin/backup/restore`.
4. The service enters code-enforced maintenance, drains active requests, blocks new requests and application mutations, stages the filesystem, and runs `pg_restore --single-transaction`.
5. Verify the restore response and the Completed job/audit state.
6. Verify representative `document_versions`/`document_files` pairs, redaction confirmations, protected document diffs, backup schedule rows, and corresponding filesystem paths/hashes.
7. If startup reports an Interrupted restore, inspect the database and filesystem because the result is unknown; then submit an Admin Completed or Failed resolution with a concrete verification note.
8. Verify the generated health URL after maintenance ends.
9. Verify representative metadata, version downloads, audit records, workflow state, notifications, backup metadata and sampled file hashes.

## Evidence boundary

Strict backup/restore source paths, the exclusive application mutation barrier, persisted schedule, code-enforced maintenance, Interrupted resolution and contracts can be inspected statically. A successful recovery claim requires an executed restore artifact tied to the exact tested revision and recovery inputs. No documentation or static report may be treated as proof of automated physical PITR.
