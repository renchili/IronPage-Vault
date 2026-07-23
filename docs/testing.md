# Testing Guide

IronPage Vault separates static repository acceptance, Go unit tests, repository contracts, stateful API/browser acceptance, local report generation, and complete regression evidence. Every execution result applies only to the revision that produced it.

## Test locations

```text
internal/**/*_test.go  colocated Go unit and package tests
tests/contracts/       repository, transaction-shape, structure, recovery-integrity, and generated-contract checks
tests/api/             HTTP, PostgreSQL, filesystem, PDF, bootstrap, auth, backup and browser flows
ci/                    static workflow contracts and manual full-regression helpers
```

`run_tests.sh` is the local report entrypoint. `ci/run_full_regression.sh` is the complete regression entrypoint. Neither is called by the GitHub static workflow.

## Static reviewer boundary

A static acceptance reviewer reads source and pre-existing evidence only. The reviewer must not run tests, scripts, generators, formatters, builds, containers, databases, browsers, deployments, or CI to fill evidence gaps. Missing execution does not alter the static verdict, and static inspection must not claim runtime execution.

## Swagger generation and contract boundary

Route-level annotations under `internal/app/swagger_*.go` are the API contract source of truth. `scripts/generate_swagger.sh` writes transient generated files under `docs/swagger/`:

```text
docs/swagger/docs.go
docs/swagger/swagger.json
docs/swagger/swagger.yaml
```

Supported execution entrypoints generate these files before compilation or generated-contract comparison. `run_tests.sh`, the Docker builder, and `ci/run_full_regression.sh` own that lifecycle. `.github/workflows/ci.yml` does not generate Swagger or compile the application; it checks route annotations statically through `tests/contracts/swagger_route_coverage.sh`.

Generated files are revision-specific artifacts, not hand-maintained repository documentation. A static reviewer must not run the generator to fill a missing artifact and must not treat an earlier generated file as proof for the current revision. `ci/swagger_contract_check.sh` may compare generated routes only when a supported execution entrypoint already generated the files.

## Local report entrypoint

```bash
bash run_tests.sh
```

Stateful stages additionally require an isolated acceptance service and:

```text
BASE_URL
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

When absent, affected stages are `SKIP`, the report is `INCOMPLETE`, and exit status is `2`. The report describes only rows in `results.tsv`; a probe cannot claim unexecuted RBAC, PDF, backup, browser, Docker, or full-regression coverage.

## Static GitHub acceptance

`.github/workflows/ci.yml` is the sole GitHub workflow and performs static acceptance only.

Admission behavior:

- automatic targets are derived from pull-request, merge-group, or `main` push context;
- manual `target` must equal the selected branch, or identify the exact same-repository open PR whose branch and head SHA equal the selected ref;
- target concurrency uses `cancel-in-progress: true`;
- admission occurs before checkout and repository-controlled code;
- denied admission is cancelled rather than sleeping;
- history pagination is limited to the current workflow;
- cooldown and failure latching apply to the canonical target and exact revision;
- a different corrective revision is admitted immediately;
- ordinary reruns are denied;
- one-time unlock requires the canonical target, exact failed run ID, reviewed reason, same revision, and unused marker.

The later job defines static checks for workflow/shell/Python syntax, Go formatting, source inventory, documentation, repository/structure contracts, scheduled backup, metadata storage, backup/restore integrity, and Swagger routes. The successful source inventory is retained only after every static gate succeeds.

GitHub creates a workflow-run object before YAML admission executes. The repository provides pre-checkout rejection and active-run collapse, not platform-level pre-dispatch prevention.

## Transaction and state-integrity definitions

Static contracts require:

- upload to commit document, version, `document_files`, and audit in one transaction and remove an orphan directory on failure;
- redaction confirmation to commit version/file entity, proposal state, `redaction_confirmations`, document state, history/notification, and audit together;
- Bates page-number range, job, version/file entity, document pointer and audit to commit together;
- comparison to commit protected `document_diffs` and `DOCUMENT_DIFF_CREATE` audit together;
- workflow/finalization document/history/audit/notification to share one transaction;
- Admin workflow PUT to validate ordered definitions and runtime transitions to query persisted definitions;
- annotation/mention side effects to check every write;
- audit source IP/metadata to use ciphertext, deterministic source lookup/backfill, response decryption, and a non-empty acting user;
- backup artifact cleanup when job/audit persistence fails;
- every application mutation to participate in the shared advisory barrier;
- manual and scheduled backup to hold the exclusive barrier across PostgreSQL dump and filesystem tar;
- scheduled backup job/audit attribution to use the protected system principal;
- scheduled backup settings to be Admin-managed, PostgreSQL-persisted, range-validated, reloaded at startup and every minute, and independent of `BACKUP_INTERVAL`;
- global restore admission to prevent concurrent restore authentication;
- authenticated Admin restore to activate route maintenance before handler work, reject new requests, drain active requests and prevent live mutation;
- staged restore path validation, filesystem rollback and PostgreSQL single-transaction mode;
- a Requested journal with no durable platform result to become Interrupted/unknown rather than Failed;
- Interrupted resolution to require an Admin, Completed or Failed conclusion, and a non-empty verification note;
- Admin config to reject deployment-owned/unknown keys and invalid pagination or backup schedule values before persistence;
- PostgreSQL child argv to exclude database passwords and use a mode-`0600` PGPASSFILE;
- page values to clamp before offset multiplication.

## Required boundary and negative definitions

| Requirement | Definition |
|---|---|
| 59-second freshness accepted | `tests/api/test_request_guard_edges.sh` |
| 61-second old/future timestamps rejected | `tests/api/test_request_guard_edges.sh` |
| same JWT/JTI request ID replay rejected | `tests/api/test_request_guard_edges.sh` |
| different JWT/JTI replay scope | `tests/api/test_request_guard_edges.sh` |
| 0/1/499/500/501 PDF pages | `TestInspectPDFPageBoundaries` |
| `/Pages` root and compressed token false positives | `TestInspectPDFIgnoresPagesRootAndCompressedStreamTokens` |
| page object in compressed object stream | `TestInspectPDFReadsPageFromCompressedObjectStream` |
| malformed PDF rejection | `TestInspectPDFRejectsMalformedPDF` |
| 49 → 50 allowed | `TestVersionLimitAllowsFortyNineToFifty` |
| 50 → 51 rejected | `TestVersionLimitRejectsFiftyToFiftyOne` |
| rollback target within ceiling | `TestRollbackVersionMustRemainWithinCeiling` |
| Admin backup Boolean/interval validation | `TestBackupScheduleConfigurationValidation`, `tests/api/test_admin_ops.sh` |
| scheduler restart/reload source contract | `ci/scheduled_backup_contract_check.sh` |
| complete Finalized mutation matrix | `tests/api/test_finalized_immutability.sh` |

The Finalized definition stages a redaction and annotation before finalization, then verifies rejection of rollback, redaction proposal, redaction confirmation, annotation creation, annotation disposition, Bates, workflow transition, and repeated finalization. It snapshots and rechecks version, redaction, annotation, audit, and notification counts. The backend has no replacement-upload or metadata-mutation route, so those categories are statically verified as absent rather than represented by invented endpoints.

Version-limit source ordering is also guarded: redaction must call `nextDocumentVersion` before `ApplyRedactionBurnIn`; Bates must call it before `AllocateBatesRange`. This establishes that revision 51 cannot create an orphan output or partially reserve the global sequence.

## Recovery and configuration test definitions

| Behavior | Definition |
|---|---|
| unsafe method and exclusive operation classification | `internal/app/operation_barrier_test.go` |
| exact restore request classification | `TestRestoreRequestClassificationIsExact` |
| maintenance denial and concurrent restore rejection | `TestMaintenanceRejectsOrdinaryAndConcurrentRestoreRequests` |
| Requested becomes Interrupted/unknown | `TestRequestedRestoreBecomesInterruptedNotFailed` |
| lifecycle journal encryption/plaintext rejection | `internal/app/restore_lifecycle_test.go` |
| pagination pair bounds and config ownership | `internal/app/config_management_test.go` |
| backup schedule validation | `internal/app/backup_scheduler_test.go`, `internal/app/backup_interval_test.go` |
| maximum page offset cannot overflow | `TestMaximumPageOffsetDoesNotOverflowInt` |
| password-free pg_dump/pg_restore argv | `TestPostgresCommandArgumentsExcludePassword` |
| PGPASSFILE mode and escaping | `TestPGPassFileUsesRestrictedModeAndEscaping` |
| ambient PostgreSQL credentials removed | `TestPostgresCommandEnvironmentRemovesAmbientPassword` |
| deployment-owned/unknown/invalid config HTTP errors | `tests/api/test_admin_ops.sh` |
| restore resolution missing-journal error | `tests/api/test_admin_ops.sh` |
| complete source-shape guard | `tests/contracts/backup_restore_integrity.sh` |
| complete Swaggo route inventory | `tests/contracts/swagger_route_coverage.sh` |

`internal/app/workflow_definitions_test.go` defines ordered definitions and ordered-chain validation. `internal/app/pii_storage_test.go` defines encrypted and legacy audit response opening. `tests/api/test_admin_ops.sh` defines Admin workflow replacement, backup schedule management, non-Admin denial, opened audit fields, strict backup/restore, configuration ownership, range validation and reconciliation errors.

## Docker and API acceptance definitions

```bash
bash ci/docker_acceptance.sh
```

The orchestrator is defined to create independent generated runtime files without fixed port, database identity, path, credential, or image fallback. The API suite defines:

- generated bootstrap/restart behavior;
- rolling lockout, login/session/logout failure handling and audit;
- real timestamp freshness and replay rejection through authentication middleware;
- role and object-access positive/negative paths;
- persisted workflow management, transitions, history, notification and complete terminal immutability;
- strict redaction, transactional Bates numbering, persisted required entities, and structured comparison;
- 500-page and 50-revision boundary definitions;
- audit filtering/decryption, annotations, mention notifications and read acknowledgement;
- backup artifacts, persisted schedule, restore state, maintenance and recovery;
- Admin configuration ownership and pagination integrity;
- uniform errors and pagination.

These are test definitions. Only an existing execution artifact proves what ran for its exact revision.

## Browser acceptance definitions

`public/index.html` is served at `/ui/` only in acceptance mode. Screenshot definitions cover rendering; the CDP interaction definition covers input validation, focus, incorrect/successful login, network failure, recovery and retry. A source file or static screenshot is not interaction execution evidence.

## Complete regression

```bash
bash ci/run_full_regression.sh artifacts/regression
```

The runner is defined as sequential and fail-fast, with stage results, truthful summary, source inventory, logs and UI evidence directories under `artifacts/regression/`. A static reviewer does not execute it or require its artifact.

## Evidence boundary

A reviewer report is a static summary. When optional existing execution evidence is cited, record tested SHA, run/job, generated status, artifact identity/digest, tree differences and omitted checks.
