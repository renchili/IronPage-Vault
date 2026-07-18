# Backup and Recovery

IronPage Vault uses local air-gapped storage only. Backup and recovery must preserve PostgreSQL state and local PDF files at the same recovery boundary.

## Data categories

PostgreSQL contains users, sessions, document metadata, version metadata, workflow history, annotations, protected redaction metadata, Bates allocation, audit logs, notifications, configuration, and backup-job metadata.

The local filesystem contains PDF binaries, transformed versions, and backup artifacts.

## Installation storage layout

`scripts/deploy.sh` generates storage targets into `.env`:

```text
POSTGRES_VOLUME_ROOT
PGDATA
IRONPAGE_VOLUME_ROOT
STORAGE_DIR
BACKUP_DIR
```

`docker-compose.yml` mounts one PostgreSQL volume at `POSTGRES_VOLUME_ROOT` and one product-data volume at `IRONPAGE_VOLUME_ROOT`. PDF storage and backup outputs are subdirectories of the generated product-data root.

Do not use assumed fixed container paths when operating an installation. Read its retained `.env`.

## Strict backup behavior

The Admin backup API creates a local full-backup job only when both strict artifacts succeed.

Implementation path:

```text
internal/app/backup_file.go -> internal/platform/backup_strict.go -> internal/platform/backup_exec.go
```

| Artifact | Required mode | Purpose |
|---|---|---|
| PostgreSQL custom dump | `pg_dump_custom` | Database restore input |
| Filesystem tar snapshot | `tar` | PDF storage restore input |

If either artifact is missing or produced through a fallback mode, the API must not report a restore-supported backup.

A successful manifest records:

```text
backup_id
created_at
database_dump_path
file_snapshot_path
database_dump_mode
file_snapshot_mode
restore_supported
```

## Operational prerequisites

The single service image includes PostgreSQL dump/restore tools and `tar`. The generated backup and storage directories must remain writable and readable by the service. No backup or restore step requires cloud storage, a remote service, or internet access.

## Strict restore behavior

Restore requires both:

```text
database_dump_path
file_snapshot_path
```

Empty, missing, or unreadable artifact paths are rejected. A successful strict restore reports the PostgreSQL and filesystem restore modes only after both operations complete.

## Recovery order

1. Stop application writes.
2. Retain the installation `.env` and identify its generated database and storage targets.
3. Identify the backup manifest and both artifact paths.
4. Restore PostgreSQL from the custom-format dump.
5. Restore PDF storage from the tar snapshot.
6. Start the application with the same installation configuration.
7. Verify the generated health URL.
8. Verify representative document metadata, version records, file downloads, audit records, and notifications.

## Consistency requirement

Database metadata and PDF files must represent the same recovery point. Restoring only one side is not a supported recovery.

## Acceptance evidence

Runtime acceptance must prove:

- Admin creates a strict full backup;
- the backup job and manifest are queryable;
- both artifacts are present and restore-supported;
- restore rejects an incomplete request;
- restore consumes the returned artifact paths;
- representative state remains readable after restore; and
- no external network or cloud dependency is used.

Static source inspection can verify the strict code path and configuration ownership but cannot prove that a backup or restore executed successfully for the current revision.
