# Requirement Check

This document maps required behavior to implementation and static proof paths. Existing execution artifacts are revision-specific optional context and must never be presented as execution of another revision.

## Verification model

| Layer | Purpose | Evidence boundary |
|---|---|---|
| Static reviewer acceptance | Inspect complete current source without executing project or CI work | Missing external execution does not alter the static verdict |
| GitHub static workflow | Detect source, layout, configuration, documentation, workflow, UI-rule, and naming contradictions | Static evidence only |
| Local report | Record stages actually run by `run_tests.sh` | Claims only rows in its generated `results.tsv` |
| Docker/API/browser acceptance | Exercise generated deployment and critical stateful flows | Optional runtime evidence for the exact execution |
| Complete regression | Run every defined stage sequentially and retain its artifact | Optional execution evidence for its tested revision |

The sole GitHub workflow derives automatic targets from the event. A manual target must equal the selected branch or identify the exact same-repository open PR whose branch and head SHA equal the selected workflow ref. Admission happens before checkout, collapses superseded active runs, paginates the workflow history, scopes cooldown and failure latching to the canonical target/revision, and permits one exact reviewed unlock.

Repository YAML cannot prevent GitHub from first creating a workflow-run object. Claims of true pre-dispatch blocking require separate platform-level evidence.

## Requirement and proof paths

| Area | Implementation path | Static acceptance proof |
|---|---|---|
| Complete local configuration | `scripts/deploy.sh`, Compose, Dockerfile, `internal/app/config.go`, `internal/app/database.go` | installation-specific identity, paths, ports and credentials; generated `BACKUP_DIR` is persisted after migration; schema contains no fixed machine backup path |
| Initial administrator | `internal/app/database.go` | empty-installation-only creation path and restart-preservation definitions |
| Host exposure | `scripts/deploy.sh` | IPv4 loopback validation, available-port selection logic, and explicit bind-race handling |
| Acceptance fixtures and UI | config/server, `public/index.html` | acceptance-mode fixture requirements, normal-mode UI exclusion, and one canonical surface |
| Buildable frontend generation rules | `skills/project-generation-workflow/SKILL.md`, `skills/full-project-acceptance-hard-gates/SKILL.md`, `docs/questions.md` | UI scope is conditional; production UI and implementation-guiding prototypes require exact component, icon, size, state, special-interaction, accessibility, platform-review, and traceability decisions; arbitrary YAML/JSON packages cannot substitute for the requested artifact or implementation |
| Rolling login lockout | `internal/app/auth.go`, migration 002 | rolling-window user lock plus failed-login audit in one transaction |
| Authentication persistence | `internal/app/auth.go` | successful reset/session/audit and logout blacklist/session/audit commit atomically; blacklist, replay and session errors fail closed |
| Persisted workflow management | `workflow_definitions.go`, `workflows.go`, `server.go` | Admin GET/PUT route, ordered validation, in-use-state protection, database-resolved successor, and transactional history/audit/notification |
| Finalized immutability | workflow and mutator guards | every material mutation path rejects a Finalized document |
| Strict redaction | `redactions.go`, service/platform adapters | proposal/audit transaction; strict file generation; checked version/document/proposal writes; failure cleanup; transactional history/notification/audit |
| Bates numbering | `bates_version.go`, `repository/bates.go` | page-range reservation, PDF generation, job/version/document/audit in one transaction, rollback and file cleanup |
| Version comparison | comparison service/platform extraction | structured text, page, bounding-box and modified-block implementation |
| Annotation side effects | `annotations.go`, `mentions.go`, `notifications.go` | annotation, audit, encrypted-username mention lookup, unread-cap handling and notifications share one transaction |
| Protected audit API | `domain_events.go`, `repository/audit.go`, `audit_filters.go`, migration 003 | ciphertext storage, deterministic source-IP lookup/backfill, typed protected query and response decryption |
| Admin configuration | `admin.go`, `template_update.go`, `workflow_definitions.go` | user/config/template/workflow updates include audit in the same transaction |
| Backup | `backup_file.go`, `backup_scheduler.go`, `backup_cleanup.go` | dump/tar/metadata creation plus job/audit transaction; all generated paths removed on persistence failure |
| Restore | `platform/backup_exec.go`, `restore.go`, `restore_lifecycle.go`, `server.go` | safe staged extraction, reversible storage swap, single-transaction PostgreSQL restore, encrypted durable lifecycle journal, preserved acting user, idempotent Requested/terminal audits, and fail-closed startup reconciliation |
| Runtime limits and uniform errors | config/core/API handlers | 200 MB, 500 pages, 250 files, 50 versions, configured pagination and uniform envelope definitions |
| Collection pagination | `pagination_config.go`, `admin.go`, `documents.go`, `redactions.go`, `annotations.go`, `audit_filters.go` | every collection query applies configured `page`/`page_size`, SQL `LIMIT`/`OFFSET`, default 25 and maximum 100; response includes `data`, `page`, and `page_size` |
| Canonical repository layout | `tests/api/`, `tests/contracts/`, `public/index.html` | no legacy test directories, duplicate UI or process-status documents |
| Path and contamination audit | `ci/source_inventory.py` | all tracked files, collisions, near duplicates, unsafe characters, mixed conventions and explicit exceptions |
| Truthful local report | `run_tests.sh`, `ci/run_tests_contract_check.sh` | report coverage equals stage rows and skipped stages cannot pass |
| CI admission safety | `.github/workflows/ci.yml` | canonical manual target validation, active-run collapse, pre-checkout admission, scoped history/cooldown/latch, exact unlock and static-only later job |
| Documentation truth | README/docs/contracts | current paths, transaction boundaries, API routes, evidence rules, UI-generation boundaries, and limitations agree with source |

## Collection pagination route mapping

| Collection route | Handler source | Contract definition |
|---|---|---|
| `GET /api/admin/users` | `internal/app/admin.go` | `tests/api/test_api_contracts.sh`, `swagger_auth_admin.go` |
| `GET /api/admin/config` | `internal/app/admin.go` | `tests/api/test_api_contracts.sh`, `swagger_auth_admin.go` |
| `GET /api/admin/workflow-statuses` | `internal/app/admin.go` | `tests/api/test_api_contracts.sh`, `swagger_admin_backup.go` |
| `GET /api/admin/notification-templates` | `internal/app/admin.go` | `tests/api/test_api_contracts.sh`, `swagger_admin_backup.go` |
| `GET /api/admin/backup/jobs` | `internal/app/admin.go` | `tests/api/test_api_contracts.sh`, `swagger_admin_backup.go` |
| `GET /api/documents` | `internal/app/documents.go` | `tests/api/test_api_contracts.sh`, `swagger_documents.go` |
| `GET /api/documents/:id/versions` | `internal/app/documents.go` | `tests/api/test_api_contracts.sh`, `swagger_documents.go` |
| `GET /api/documents/:id/redactions` | `internal/app/redactions.go` | `tests/api/test_api_contracts.sh`, `swagger_workflow_review.go` |
| `GET /api/documents/:id/annotations` | `internal/app/annotations.go` | `tests/api/test_api_contracts.sh`, `swagger_annotations_misc.go` |
| `GET /api/audit-logs` | `internal/app/audit_filters.go` | `tests/api/test_api_contracts.sh`, `swagger_annotations_misc.go` |
| `GET /api/notifications` | `internal/app/admin.go` | `tests/api/test_api_contracts.sh`, `swagger_annotations_misc.go` |

Each route has default, minimum-normalization, explicit boundary, and maximum-clamping definitions in `tests/api/test_api_contracts.sh`. The static repository contract checks that every listed route, handler, and Swaggo pagination annotation remains present.

## Static acceptance conditions

Static acceptance completes from current source, test definitions, schemas, migrations, configuration, workflows, deployment definitions, UI assets when applicable, documentation, comments, and repository structure. It does not wait for or require deployment, database, browser, CI, or regression execution. A reviewer must not run those systems to fill evidence gaps.

Existing external artifacts may be described only as optional read-only context for their exact revision and scope. A reviewer report is a static review summary, never a runtime artifact.

## Generated API documentation

Route-level Swaggo annotations under `internal/app/swagger_*.go` are authoritative. They include both GET and PUT workflow-definition routes and pagination parameters on every collection route. Generated files under `docs/swagger/` are produced only by supported execution entrypoints; static review does not authorize generation.

## Compatibility notes

Compatibility columns may remain for migration safety, but protected plaintext is not the source of truth. Audit source-IP lookup backfill is performed from ciphertext or legacy plaintext after migration. Deterministic values are lookup-only and are not returned as user data.
