# Point-in-Time Recovery

This document describes the intended point-in-time recovery model for IronPage Vault in a local-only deployment.

## Purpose

Point-in-time recovery protects legal document metadata, audit records, workflow history, notification records, configuration, and backup metadata from accidental local data loss.

## Local-only Constraint

PITR must not require cloud services or external network access. WAL archives and logical dump outputs should be written to a local backup volume.

## Recovery Inputs

A full recovery point should include:

- PostgreSQL base backup or logical dump
- WAL archive when physical PITR is enabled
- PDF storage snapshot
- configuration snapshot
- backup metadata record

## Relationship Between Database and Filesystem

PostgreSQL stores the file path and hash for each PDF version. The filesystem stores the actual PDF binary.

A correct recovery must restore both sides consistently:

```text
document_versions.file_path -> restored local PDF file
```

## Recommended PITR Strategy

For a stricter production deployment:

1. Enable PostgreSQL WAL archiving to the local backup volume.
2. Schedule periodic base backups.
3. Snapshot PDF storage at the same recovery boundary.
4. Record the recovery target timestamp in backup job metadata.
5. Document the restore command used for each backup job.

## Prototype Status

The current prototype records backup job metadata and documents the local recovery strategy. Full automated WAL archiving and restore orchestration are planned hardening work.

## Acceptance Questions

- Are PostgreSQL data and PDF storage both included?
- Is the target recovery time documented?
- Does the restore procedure avoid external services?
- Can restored metadata find the restored PDF files?
- Are audit logs preserved?
