# Point-in-Time Recovery

This document describes the supported point-in-time recovery documentation scope for IronPage Vault in a local-only deployment.

## Purpose

Point-in-time recovery planning protects legal document metadata, audit records, workflow history, notification records, configuration, backup metadata, and PDF storage references from accidental local data loss.

## Local-only constraint

PITR guidance must not require cloud services or external network access. Recovery inputs must be local artifacts, local volumes, or locally retained database/archive material.

## Current supported scope

The current project supports documented local recovery strategy and strict backup/restore artifact semantics.

The current repository does not prove automated physical PITR orchestration. Documentation must not claim automated WAL archiving, scheduled physical base backups, or timestamp-target restore automation unless those features are implemented and validated.

## Recovery inputs

A complete recovery point should include:

- PostgreSQL dump or base backup material.
- WAL archive when physical PITR is enabled by the operator.
- PDF storage snapshot.
- configuration snapshot.
- backup manifest or backup metadata record.

## Relationship between database and filesystem

PostgreSQL stores the file path and hash for each PDF version. The filesystem stores the actual PDF binary.

A correct recovery must restore both sides consistently:

```text
document_versions.file_path -> restored local PDF file
```

If metadata is restored to a point where a referenced PDF file does not exist, document download and version inspection will fail.

## Documented PITR strategy

For a stricter local deployment, operators should document:

1. where database backup material is retained.
2. where any local archive material is retained.
3. how PDF storage snapshots are aligned to the same recovery boundary.
4. the target timestamp or backup ID.
5. the exact restore command used.
6. the post-restore verification checks.

## Verification checks

After a recovery exercise, verify:

- `/healthz` returns successfully.
- representative document metadata is present.
- representative document version files are downloadable.
- audit logs expected for the recovery point exist.
- workflow state is consistent with the selected recovery point.
- backup metadata and artifact paths are understandable to the operator.

## Evidence boundary

The re-audit report records strict backup/restore evidence, but current documentation must not overstate PITR. Full automated PITR remains unproven unless a future implementation and validation run demonstrate it.

The project must not simultaneously claim both `no tracked limitations` and `planned PITR hardening`. This document treats automated physical PITR orchestration as not currently proven.
