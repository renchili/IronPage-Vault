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

The later job defines static checks for workflow/shell/Python syntax, Go formatting, source inventory, documentation, repository/structure contracts, backup/metadata contracts, backup/restore integrity, and Swagger routes. The successful source inventory is retained only after every static gate succeeds.

GitHub creates a workflow-run object before YAML admission executes. The repository provides pre-checkout rejection and active-run collapse, not platform-level pre-dispatch prevention.

## Transaction and state-integrity definitions

Static contracts require:

- upload document/version/audit to share one transaction and remove an orphan directory on failure;
- workflow/finalization document/history/audit/notification to share one transaction;
- Admin workflow PUT to validate ordered definitions and runtime transitions to query persisted definitions;
- redaction proposal/confirmation and annotation/mention side effects to check every write;
- Bates page-range reservation, job, version, document pointer and audit to commit together;
- audit source IP/metadata to use ciphertext, deterministic source lookup/backfill, response decryption, and a non-empty acting user;
- backup artifact cleanup when job/audit persistence fails;
- every application mutation to participate in the shared advisory barrier;
- manual and scheduled backup to hold the exclusive barrier across PostgreSQL dump and filesystem tar;
- scheduled backup job/audit attribution to use the protected system principal;
- global restore admission to prevent concurrent restore authentication;
- authenticated Admin restore to activate route maintenance before handler work, reject new requests, drain active requests and prevent live mutation;
- staged restore path validation, filesystem rollback and PostgreSQL single-transaction mode;
- a Requested journal with no durable platform result to become Interrupted/unknown rather than Failed;
- Interrupted resolution to require an Admin, Completed or Failed conclusion, and a non-empty verification note;
- Admin config to reject deployment-owned/unknown keys and invalid pagination pairs before persistence;
- PostgreSQL child argv to exclude database passwords and use a mode-`0600` PGPASSFILE;
- page values to clamp before offset multiplication.

## Recovery and configuration test definitions

| Behavior | Definition |
|---|---|
| unsafe method and exclusive operation classification | `internal/app/operation_barrier_test.go` |
| exact restore request classification | `TestRestoreRequestClassificationIsExact` |
| maintenance denial and concurrent restore rejection | `TestMaintenanceRejectsOrdinaryAndConcurrentRestoreRequests` |
| Requested becomes Interrupted/unknown | `TestRequestedRestoreBecomesInterruptedNotFailed` |
| lifecycle journal encryption/plaintext rejection | `internal/app/restore_lifecycle_test.go` |
| pagination pair bounds and config ownership | `internal/app/config_management_test.go` |
| maximum page offset cannot overflow | `TestMaximumPageOffsetDoesNotOverflowInt` |
| password-free pg_dump/pg_restore argv | `TestPostgresCommandArgumentsExcludePassword` |
| PGPASSFILE mode and escaping | `TestPGPassFileUsesRestrictedModeAndEscaping` |
| ambient PostgreSQL credentials removed | `TestPostgresCommandEnvironmentRemovesAmbientPassword` |
| deployment-owned/unknown/invalid config HTTP errors | `tests/api/test_admin_ops.sh` |
| restore resolution missing-journal error | `tests/api/test_admin_ops.sh` |
| complete source-shape guard | `tests/contracts/backup_restore_integrity.sh` |
| complete Swaggo route inventory | `tests/contracts/swagger_route_coverage.sh` |

`internal/app/workflow_definitions_test.go` defines ordered definitions and ordered-chain validation. `internal/app/pii_storage_test.go` defines encrypted and legacy audit response opening. `tests/api/test_admin_ops.sh` defines Admin workflow replacement, non-Admin denial, opened audit fields, strict backup/restore, configuration ownership, range validation and reconciliation errors.

## Docker and API acceptance definitions

```bash
bash ci/docker_acceptance.sh
```

The orchestrator is defined to create independent generated runtime files without fixed port, database identity, path, credential, or image fallback. The API suite defines:

- generated bootstrap/restart behavior;
- rolling lockout, login/session/logout failure handling and audit;
- role and object-access positive/negative paths;
- persisted workflow management, transitions, history, notification and terminal immutability;
- strict redaction, transactional Bates numbering and structured comparison;
- audit filtering/decryption, annotations, mention notifications and read acknowledgement;
- backup artifacts, restore state, maintenance and recovery;
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
