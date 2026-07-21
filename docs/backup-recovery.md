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

Before external restore work begins, the API creates a restore ID and writes an encrypted lifecycle journal under the installation's generated `BACKUP_DIR/.restore-lifecycle`. The journal retains the requesting Admin, request ID, source IP, artifact paths, and lifecycle metadata as AES-256-GCM ciphertext. It is outside `STORAGE_DIR`, so replacing the document filesystem snapshot does not erase the recovery record.

The API then transactionally stores the `Requested` backup-job state and `BACKUP_RESTORE_REQUESTED` audit. The same acting Admin is retained for every later state. After the strict restore returns, the journal is atomically advanced to `Completed` or `Failed` before the terminal database state and audit are attempted.

Because `pg_restore` replaces PostgreSQL state, terminal persistence idempotently restores both the Requested audit and the matching Completed or Failed audit into the restored database. The `backup_jobs.created_by` value is never replaced with `NULL`. A `200` response is returned only after the Completed job state and both acting-user audit records exist in PostgreSQL.

## Restore lifecycle reconciliation

A terminal database or audit failure cannot be rolled back across a PostgreSQL restore and filesystem swap. The encrypted journal is therefore the durable reconciliation source:

1. `Requested` is written before external restore work;
2. `Completed` or `Failed` is written after the platform result is known;
3. the terminal job state and audits are committed together;
4. the journal is removed only after that commit;
5. startup reconciles every remaining journal before the backup scheduler or HTTP API starts.

A retained Completed or Failed journal is replayed idempotently. A journal still in Requested means the process ended before it could durably record the platform outcome; startup converts it to Failed with explicit metadata requiring operators to verify restored data before retrying. Malformed, undecryptable, or unpersistable journals fail startup rather than silently leaving an ambiguous restore lifecycle.

If the final journal deletion fails after a successful database commit, the API still returns the completed restore together with `lifecycle_reconciliation_pending=true`; the next startup verifies and removes the idempotent journal.

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

The PostgreSQL command and filesystem rename cannot be one cross-system ACID transaction. The implementation therefore uses database single-transaction restore, reversible filesystem installation, an encrypted lifecycle journal, explicit restore states, startup reconciliation, and fail-closed success reporting.

## Operational recovery order

1. Stop application writes.
2. Retain `.env` and the generated database/storage identity.
3. Select a completed backup and both paths returned by its manifest/API response.
4. Submit both paths to the Admin restore route.
5. Verify the returned restore ID is `Restored` and the jobs list contains its Completed restore row with the requesting Admin in `created_by`.
6. Verify Admin audit results contain both `BACKUP_RESTORE_REQUESTED` and `BACKUP_RESTORE_COMPLETED` for that acting user and request ID.
7. If `lifecycle_reconciliation_pending` is returned or the process exits during restore, restart with the same installation configuration; startup must reconcile the journal before serving requests.
8. Verify representative documents, versions, files, audit records and notifications.

Restoring only the dump or only the filesystem archive is unsupported. Reconciliation also requires the requesting Admin identity to exist in the restored same-installation snapshot so the audit foreign-key contract can be preserved.

## Static and execution evidence

Static inspection can verify strict mode checks, cleanup paths, safe extraction, rollback functions, PostgreSQL single-transaction flags, encrypted lifecycle persistence, acting-user preservation, idempotent audit reconciliation, startup ordering, collection visibility, and failure handling. It does not claim that an execution occurred. Existing runtime evidence, when available, applies only to its exact revision and inputs; static review neither requires nor creates it.
