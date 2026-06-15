# Backup and Recovery

IronPage Vault is designed for local air-gapped operation. Backup and recovery must use local volumes and local files only.

## Data Categories

The system has two data categories:

1. PostgreSQL data
2. Local PDF binary files

PostgreSQL contains:

- users
- sessions
- document metadata
- version metadata
- workflow history
- annotations
- redaction metadata
- Bates job metadata
- audit logs
- notifications
- configuration entries
- backup job metadata

The filesystem contains:

- PDF binaries
- transformed PDF versions
- local backup outputs

## Docker Volumes

The Compose file declares volumes for:

```text
ironpage_pgdata
ironpage_storage
ironpage_backups
```

Inside the container these map to:

```text
/var/lib/postgresql/data
/var/lib/ironpage/storage
/var/lib/ironpage/backups
```

## Logical Database Backup

The Admin backup API creates backup job metadata. A production-ready local backup worker should run a PostgreSQL logical dump into the backup volume.

Recommended local command pattern inside the container:

```bash
pg_dump -U ironpage -d ironpage > /var/lib/ironpage/backups/ironpage_$(date -u +%Y%m%dT%H%M%SZ).sql
```

## Filesystem Snapshot

PDF storage should be copied or snapshotted together with the database backup metadata.

Recommended local pattern:

```bash
tar -czf /var/lib/ironpage/backups/pdf_storage_$(date -u +%Y%m%dT%H%M%SZ).tgz /var/lib/ironpage/storage
```

## Recovery Order

1. Stop the application container.
2. Restore PostgreSQL data or replay the logical dump.
3. Restore PDF storage files to the expected storage directory.
4. Restore backup metadata if it was separated.
5. Start the application.
6. Verify `/healthz`.
7. Verify document metadata and file download for sample records.

## Consistency Requirement

Database metadata and PDF files must represent the same point in time. If metadata points to a file path that was not restored, document download and version inspection will fail.

## Acceptance Checks

- Admin can create backup job metadata.
- Backup job rows are queryable.
- The configured backup directory is local.
- Documentation explains database and filesystem recovery.
- No backup step requires cloud storage or internet access.
