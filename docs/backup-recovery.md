# Backup and Recovery

IronPage Vault uses local air-gapped storage only. Backup and recovery preserve PostgreSQL state and local PDF files at one recovery boundary.

## Installation storage layout

`scripts/deploy.sh` generates and retains:

```text
POSTGRES_VOLUME_ROOT
PGDATA
IRONPAGE_VOLUME_ROOT
STORAGE_DIR
BACKUP_DIR
```

The schema does not seed a fixed machine path. After migration, startup writes the installation's actual `BACKUP_DIR` to `config_entries`. Operators must use the retained `.env`, not an assumed container path.

## Strict backup

The Admin backup API succeeds only when all of these exist:

| Item | Required result |
|---|---|
| PostgreSQL custom dump | `pg_dump_custom` |
| Filesystem snapshot | `tar` |
| Manifest | paths/modes and `restore_supported=true` |
| Metadata snapshot | table counts and creation metadata |
| PostgreSQL job | `Completed` full-backup row |
| Audit | encrypted artifact metadata |

Dump, tar, manifest and metadata are generated before the database job. The job and audit commit together. If metadata write, job insertion, audit insertion, or commit fails, `backup_cleanup.go` removes the dump, tar, manifest, metadata and error/missing markers. A database record cannot report a completed backup whose files were removed, and generated files cannot remain after failed persistence.

Scheduled backup uses the same strict artifacts, cleanup rule, job/audit transaction, and no external network.

## Strict restore request

The request requires:

```json
{
  "database_dump_path": "<returned dump path>",
  "file_snapshot_path": "<returned snapshot path>"
}
```

The API creates a restore ID and transactionally records `Requested` plus `BACKUP_RESTORE_REQUESTED` before external restore work. It later records exactly one `Completed` or `Failed` state with the corresponding encrypted audit metadata. A `200` response is returned only after `Completed` and `BACKUP_RESTORE_COMPLETED` are stored.

## Staged filesystem recovery

The tar archive is opened by the Go platform adapter rather than extracted directly into live storage. Restore:

1. creates a sibling staging directory;
2. accepts regular files and directories only;
3. rejects absolute paths, `..` traversal, symlinks, hard links, devices and other special entries;
4. moves the existing storage directory to a rollback path;
5. atomically renames the staged directory into the configured storage path;
6. restores the previous directory if PostgreSQL restore fails; and
7. removes the rollback directory only after PostgreSQL succeeds.

A cleanup failure is returned as restore failure and recorded in the Failed lifecycle metadata rather than silently ignored.

## PostgreSQL recovery

`pg_restore` is required and is invoked with:

```text
--clean --if-exists --single-transaction
```

The database either commits the restored archive or rolls back that database restore. If it fails, the staged filesystem installation is rolled back to the previous directory.

The PostgreSQL command and filesystem rename cannot be one cross-system ACID transaction. The implementation therefore uses database single-transaction restore, reversible filesystem installation, explicit restore states and fail-closed success reporting.

## Operational recovery order

1. Stop application writes.
2. Retain `.env` and the generated database/storage identity.
3. Select a completed backup and both paths returned by its manifest/API response.
4. Submit both paths to the Admin restore route.
5. Verify the returned restore ID is `Restored` and the jobs list contains its Completed restore row.
6. Restart with the same installation configuration when required.
7. Verify representative documents, versions, files, audit records and notifications.

Restoring only the dump or only the filesystem archive is unsupported.

## Static and execution evidence

Static inspection can verify strict mode checks, cleanup paths, safe extraction, rollback functions, PostgreSQL single-transaction flags, restore lifecycle state and audit requirements. It does not claim that an execution occurred. Existing runtime evidence, when available, applies only to its exact revision and inputs; static review neither requires nor creates it.
