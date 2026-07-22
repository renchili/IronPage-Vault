# Design Document

## Product boundary

IronPage Vault is an air-gapped legal PDF lifecycle backend. The product boundary is the Go/Echo API, PostgreSQL metadata and security state, and local filesystem PDF storage.

`public/index.html` is the only acceptance browser probe. It is not a production frontend and does not change the backend-only product scope.

## Single-container deployment

The standalone deployment packages PostgreSQL and the Go API in one container for an offline machine with no required external service. `scripts/deploy.sh` creates a protected installation file containing the complete local runtime configuration before the image is built.

Database identity, ports, host exposure, API binding, application asset roots, PostgreSQL data roots, product storage roots, credentials, signing material, encryption material, bootstrap identity, and acceptance fixtures are deployment-owned values. The schema does not seed a machine-specific backup path; startup persists the generated `BACKUP_DIR` into `config_entries`.

Fresh configuration accepts only an IPv4 loopback host binding and selects a host port that is not currently accepting loopback TCP connections. Compose remains the final bind authority because port availability can change after the probe.

## Data ownership and mutation boundaries

PostgreSQL is the source of truth for users, login attempts, sessions, replay state, documents, versions, workflow definitions/history, audit records, protected review metadata, Bates allocation, notifications, configuration, and backup jobs. PDF binaries and transformed versions are stored on the generated local filesystem target.

A successful database mutation must include its required audit record in the same transaction. Workflow transitions additionally include status history and owner notification in that transaction. Annotation creation includes mention notifications. Authentication lockout/login/logout state includes its security audit. Notification acknowledgement, Admin configuration, user creation, template changes, workflow-definition replacement, rollback, redaction metadata, and Bates metadata use the same rule.

Every audit requires an acting user. Automated scheduled backup and startup restore reconciliation use a protected system principal rather than nullable actor state. Startup creates that identity before reconciliation and backfills legacy scheduled rows that used `NULL`.

File-producing mutations use a compensating filesystem boundary: the file is generated first while the database transaction remains uncommitted, and the generated file is removed if version, document, sequence, audit, notification, or commit work fails.

## Operation coordination

`internal/app/operation_barrier.go` coordinates application operations at two layers:

- an in-process read/write maintenance gate covers every ordinary HTTP request;
- PostgreSQL session advisory locks coordinate application mutations across cooperating API processes that share the database.

Unsafe HTTP methods acquire the shared advisory lock for their complete handler. Manual backup and scheduled backup acquire the exclusive advisory lock across metadata collection, logical dump, filesystem tar, job and audit. Restore begins in global middleware, marks maintenance active, rejects new requests, drains active requests, and acquires the exclusive advisory lock before authentication and restore work continue.

The supported deployment has one API process and one local database/filesystem installation. The barrier therefore defines the recovery boundary for all supported application mutation paths. Direct external writes to PostgreSQL or `STORAGE_DIR` are outside the operating model.

## Package boundaries

`internal/core` owns deterministic domain policy such as roles, default workflow compatibility, object access, validation, mention parsing, and notification-cap calculations.

`internal/service` coordinates use cases that combine policy, persistence, and platform operations.

`internal/repository` and `internal/store` own persistence contracts and SQL-facing behavior. SQL fragments and schema knowledge do not belong in core policy.

`internal/platform` owns encryption, digesting, filesystem operations, strict PDF transforms, backup, staged restore, safe archive extraction, and password-file handling for PostgreSQL commands.

`internal/app` owns Echo routing, middleware, request/response mapping, transaction assembly, operation barriers, runtime assembly, configuration ownership, and narrow HTTP adapters.

## Roles and object access

The supported roles are Admin, Editor, and Reviewer. Admin manages identities and system configuration, Editor performs document operations, and Reviewer performs review operations. Admin does not automatically receive Editor document authority.

Role permission and object access are separate decisions. Object access depends on principal, document owner, document status, and requested operation. Backend policy enforces both decisions even when a route is role-gated.

The system scheduler principal is reserved for automated audit attribution. It is not returned by the Admin user collection and has a random unretained password verifier.

## Document workflow

The initial chain is:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Admin reads and replaces the ordered persisted chain through `/api/admin/workflow-statuses`. `Draft` remains the first mutable state and `Finalized` remains the last immutable state. A replacement must retain every status currently used by a document. Runtime transition and finalization resolve the next state from `workflow_status_definitions`; they do not use a hard-coded chain for request validation.

A transition locks the document and commits the document state, status history, audit record, and owner notification together. Finalized is terminal; every document mutation rejects a Finalized record.

## Redaction, Bates, and comparison

Redaction is a two-phase operation. A proposal stores protected coordinate-bound metadata with its audit in one transaction. Confirmation locks the staged proposal and document, produces a strict redacted PDF, verifies it, then commits the new version, document pointer/status, proposal state, history/notification when applicable, and audit together. Failed persistence removes the generated PDF.

Bates numbering locks the document and global sequence, reserves the complete page-number range, produces and verifies the visible-numbered PDF, then commits the range, Bates job, version, document pointer, and audit together. Failed work rolls back the sequence and removes the generated PDF.

Version comparison extracts structured text and location data to report added, removed, and modified blocks with page and bounding-box information.

## Authentication, sessions, freshness, and replay

Passwords are compared through bcrypt verifiers stored in protected form. Password inputs are validated against bcrypt's 72-byte limit. Failed attempts in the preceding 15 minutes are counted under a user row lock; the fifth attempt applies a 15-minute lock. Failed-attempt state and audit, successful reset/session creation and audit, and logout blacklist/session revocation and audit each commit atomically.

JWTs are locally signed. PostgreSQL session state remains required for inactivity expiration and immediate logout. Authenticated requests require a fresh `X-Request-Timestamp` and unique `X-Request-ID`; blacklist reads, replay writes, and session updates fail closed.

## Protected metadata and audit reads

AES-256-GCM columns hold sensitive source values. Deterministic lookup keys are used only for equality lookup. Audit writes store source IP and structured metadata in ciphertext columns plus a deterministic source-IP lookup. Startup backfills that lookup for existing rows. The Admin audit route queries typed protected rows, decrypts source IP and JSON metadata before response, and never treats blank compatibility columns as the source of truth.

Restore lifecycle journals are encrypted as local AES-256-GCM envelopes. PostgreSQL command passwords are supplied through short-lived mode-`0600` `PGPASSFILE` files and are absent from `pg_dump`/`pg_restore` arguments.

## Configuration ownership

`backup.local_volume` is deployment-owned and startup refreshes it from `BACKUP_DIR`; the generic Admin PATCH route rejects it. The generic route manages only the two pagination keys. Their rows are locked together and a proposed update must preserve `1 <= default <= max <= 100` before the transaction writes either value. Pagination clamps extremely large page numbers before multiplying the offset.

## Backup and recovery

Backup success requires a PostgreSQL custom dump, filesystem archive, metadata snapshot, job record, and audit record. The exclusive application mutation barrier spans the database dump and filesystem archive, so no supported application write can split those artifacts into different states. Database metadata and audit commit together; failed database persistence removes all generated backup files.

Restore validates both artifacts, enters code-enforced maintenance, safely extracts regular files and directories to staging, rejects path traversal and link/special entries, swaps the storage directory with a rollback copy, and invokes `pg_restore --single-transaction`. Database failure restores the previous filesystem directory.

The API records Requested followed by Completed or Failed only when the platform outcome is known. A retained Requested journal after process interruption becomes Interrupted with an unknown result and a system-principal audit. It remains until an Admin verifies the restored database/filesystem and submits Completed or Failed plus a verification note through `/api/admin/backup/restore/:id/resolve`. Automated WAL-based PITR orchestration is not claimed.

## Verification architecture

Go tests remain colocated with their packages. Stateful acceptance is under `tests/api/`; repository and generated-contract checks are under `tests/contracts/`.

`.github/workflows/ci.yml` is the sole GitHub workflow and performs static acceptance only. Admission precedes checkout, validates a manual target against the selected branch or the exact same-repository open PR head revision, collapses active duplicates, applies cooldown/failure latching to the canonical target/revision pair, paginates scoped history, and consumes exact one-time unlocks. The later job runs static gates only.

`ci/run_full_regression.sh`, Docker acceptance, API flows, browser interaction, databases, and deployments are separate manual or normal-lifecycle execution paths. A static reviewer must not run or wait for them. GitHub creates a run object before repository YAML executes, so repository admission is pre-checkout rather than platform-level pre-dispatch prevention.
