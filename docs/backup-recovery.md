# Backup and Recovery

IronPage Vault uses local air-gapped storage only. Backup and recovery preserve PostgreSQL state and local PDF files at one application recovery boundary for the supported single-container deployment.

## Installation storage layout

`scripts/deploy.sh` generates and retains:

```text
POSTGRES_VOLUME_ROOT
PGDATA
IRONPAGE_VOLUME_ROOT
STORAGE_DIR
BACKUP_DIR
```

The schema does not seed a fixed machine path. After migration, startup writes the installation's actual `BACKUP_DIR` to `config_entries`. Operators must use the retained `.env`, not an assumed container path. `backup.local_volume` is deployment-owned and the generic Admin config API rejects attempts to change it.

## Application mutation barrier

Every unsafe HTTP method (`POST`, `PUT`, `PATCH`, `DELETE`) except the operations that own an exclusive boundary acquires a shared PostgreSQL advisory lock for the complete request handler. Manual backup, scheduled backup, and restore acquire the matching exclusive advisory lock on a dedicated database connection.

The exclusive lock waits for active application mutations to complete and prevents new application mutations until the operation ends. Because the supported deployment contains one API process and one local PostgreSQL/filesystem installation, this prevents upload, rollback, redaction, annotation, Bates, workflow, auth-state, config, and other application writes from crossing the dump/tar interval. Direct out-of-band filesystem or database modification is outside the supported operating model and must not occur.

## Strict backup

The Admin backup API succeeds only when all of these exist:

| Item | Required result |
|---|---|
| PostgreSQL custom dump | `pg_dump_custom` |
| Filesystem snapshot | `tar` |
| Manifest | paths/modes and `restore_supported=true` |
| Metadata snapshot | table counts and creation metadata |
| PostgreSQL job | `Completed` full-backup row |
| Audit | encrypted artifact metadata with an acting user |

The manual and scheduled paths acquire the exclusive application mutation barrier before collecting metadata, running `pg_dump`, and archiving `STORAGE_DIR`. The database dump and filesystem tar are therefore taken while application state is write-quiescent and form one recovery boundary.

Dump, tar, manifest and metadata are generated before the database job. The job and audit commit together. If metadata write, job insertion, audit insertion, or commit fails, `backup_cleanup.go` removes the dump, tar, manifest, metadata and error/missing markers. A database record cannot report a completed backup whose files were removed, and generated files cannot remain after failed persistence.

Scheduled backup uses the same strict artifacts, exclusive barrier, cleanup rule, and job/audit transaction. It attributes `created_by` and `SCHEDULED_BACKUP_CREATE` to the protected system scheduler principal rather than `NULL`.

## PostgreSQL subprocess credentials

`pg_dump` and `pg_restore` receive only port, user, database, input, output, and operation flags in their process arguments. The database password is written to a short-lived `PGPASSFILE` under `BACKUP_DIR` with mode `0600`; colons and backslashes are escaped according to the PostgreSQL password-file format. Ambient `PGPASSWORD` and `PGPASSFILE` values are removed from the child environment, and the scoped password file is removed after the command exits.

## Strict restore request

The request requires:

```json
{
  "database_dump_path": "<returned dump path>",
  "file_snapshot_path": "<returned snapshot path>"
}
```

Restore admission and maintenance are separate boundaries:

1. global middleware takes a non-blocking restore-admission mutex; a second restore request receives `409 RESTORE_ALREADY_RUNNING`;
2. the request completes normal authentication and Admin role validation while holding only that admission token; an invalid caller releases it immediately and cannot activate maintenance;
3. route middleware then marks maintenance active before restore handler work;
4. new non-restore requests receive `503 MAINTENANCE_MODE` and the local request gate waits for active reads and writes to leave;
5. the exclusive PostgreSQL advisory lock waits for any mutation owner using another cooperating API process;
6. journal creation, filesystem replacement, `pg_restore`, terminal persistence, and response all remain inside maintenance ownership.

Before external restore work begins, the API creates a restore ID and writes an encrypted lifecycle journal under the installation's generated `BACKUP_DIR/.restore-lifecycle`. The journal retains the requesting Admin, request ID, source IP, artifact paths, and lifecycle metadata as AES-256-GCM ciphertext. It is outside `STORAGE_DIR`, so replacing the document filesystem snapshot does not erase the recovery record.

The API then transactionally stores the `Requested` backup-job state and `BACKUP_RESTORE_REQUESTED` audit. The same requesting Admin remains the job creator. After the strict restore returns, the journal is atomically advanced to `Completed` or `Failed` before the matching database state and audit are attempted.

Because `pg_restore` replaces PostgreSQL state, terminal persistence idempotently restores both the Requested audit and the matching Completed or Failed audit into the restored database. The `backup_jobs.created_by` value is never replaced with `NULL`. A `200` response is returned only after the Completed job state and both acting-user audit records exist in PostgreSQL.

## Restore lifecycle reconciliation

A terminal database or audit failure cannot be rolled back across a PostgreSQL restore and filesystem swap. The encrypted journal is therefore the durable reconciliation source:

1. `Requested` is written before external restore work;
2. `Completed` or `Failed` is written only after the platform result is known;
3. the terminal job state and audits are committed together;
4. a Completed or Failed journal is removed only after that commit;
5. startup reconciles every remaining journal before the backup scheduler or HTTP API starts.

A retained Completed or Failed journal is replayed idempotently. A journal still in Requested means the process ended before it durably recorded the platform result. Startup changes that record to `Interrupted`, not Failed. Its audit is attributed to the protected system principal and records `outcome=unknown` plus `operator_verification_required`. The encrypted Interrupted journal remains on disk.

After checking representative restored database rows and filesystem objects, an Admin resolves the record without rerunning restore:

```http
POST /api/admin/backup/restore/:id/resolve
```

```json
{
  "status": "Completed",
  "verification_note": "Verified document rows, version paths, audit rows, and sampled PDF hashes"
}
```

`status` must be `Completed` or `Failed`; `verification_note` is mandatory. The resolver becomes the acting user for `BACKUP_RESTORE_RECONCILED_COMPLETED` or `BACKUP_RESTORE_RECONCILED_FAILED`. The journal is removed only after resolution state and audit commit. Malformed, undecryptable, or unpersistable journals fail startup rather than silently discarding an ambiguous restore.

If final journal deletion fails after a successful database commit, the API returns `lifecycle_reconciliation_pending=true`; the next startup replays and removes a terminal journal idempotently.

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

The PostgreSQL command and filesystem rename cannot be one cross-system ACID transaction. The implementation therefore combines authenticated restore admission, code-enforced maintenance, an application mutation barrier, database single-transaction restore, reversible filesystem installation, an encrypted lifecycle journal, explicit Requested/Completed/Failed/Interrupted states, operator resolution, and fail-closed success reporting.

## Operational recovery order

1. Retain `.env` and the generated database/storage identity.
2. Select a completed backup and both paths returned by its manifest/API response.
3. Submit both paths to the Admin restore route; after authorization, the API enforces maintenance and drains other requests.
4. Verify the returned restore ID is `Restored` and the jobs list contains its Completed restore row with the requesting Admin in `created_by`.
5. Verify Admin audit results contain both `BACKUP_RESTORE_REQUESTED` and `BACKUP_RESTORE_COMPLETED` for that acting user and request ID.
6. If the process exited with an unknown result, restart with the same installation configuration, inspect the retained Interrupted record and restored data, then use the resolution route with a concrete verification note.
7. Verify representative documents, versions, files, audit records and notifications before resuming normal use.

Restoring only the dump or only the filesystem archive is unsupported. Reconciliation also requires the requesting Admin identity to exist in the restored same-installation snapshot so the audit foreign-key contract can be preserved.

## Static and execution evidence

Static inspection can verify strict mode checks, application mutation barrier definitions, authenticated maintenance admission, cleanup paths, safe archive extraction, rollback functions, PostgreSQL single-transaction flags, PGPASSFILE handling, encrypted lifecycle persistence, acting-user preservation, Interrupted semantics, operator resolution, startup ordering, and failure handling. It does not claim that an execution occurred. Existing runtime evidence, when available, applies only to its exact revision and inputs; static review neither requires nor creates it.
