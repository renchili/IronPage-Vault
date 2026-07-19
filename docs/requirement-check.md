# Requirement Check

This document maps required behavior to implementation and static proof paths. Existing execution artifacts are revision-specific optional context and must never be presented as execution of another revision.

## Verification model

| Layer | Purpose | Evidence boundary |
|---|---|---|
| Static reviewer acceptance | Inspect complete current source without executing project or CI work | Missing external execution does not alter the static verdict |
| GitHub static workflow | Detect source, layout, configuration, documentation, workflow, and naming contradictions | Static evidence only |
| Local report | Record stages actually run by `run_tests.sh` | Claims only rows in its generated `results.tsv` |
| Docker/API/browser acceptance | Exercise generated deployment and critical stateful flows | Optional runtime evidence for the exact execution |
| Complete regression | Run every defined stage sequentially and retain its artifact | Optional execution evidence for its tested revision |

The sole GitHub workflow uses one target namespace, target concurrency with active-run collapse, checkout-free admission, full history pagination, immediate denial rather than runner sleep, an exact-revision cooldown, a failed target/revision latch, immediate admission of a different corrective revision, and one-time reviewed unlocks tied to the exact failed run.

Repository YAML cannot prevent GitHub from first creating a workflow-run object. Claims of true pre-dispatch blocking require separate platform-level evidence.

## Requirement and proof paths

| Area | Implementation path | Static acceptance proof |
|---|---|---|
| Complete local configuration | `scripts/deploy.sh`, `docker-compose.yml`, `Dockerfile`, `internal/app/config.go`, `scripts/entrypoint.sh` | installation-specific ports, identity, paths, credentials, image args, validation, and no product fallback |
| Initial administrator | `internal/app/database.go` | empty-installation-only creation path and restart-preservation definitions |
| Host exposure | `scripts/deploy.sh` | IPv4 loopback validation, available-port selection logic, and explicit bind-race handling |
| Acceptance fixtures and UI | `internal/app/config.go`, `internal/app/server.go`, `public/index.html` | acceptance-mode fixture requirements, normal-mode UI exclusion, and one canonical surface |
| Rolling login lockout | `internal/app/auth.go`, `migrations/002_login_attempt_window.sql` | rolling-window transaction logic plus positive, negative, expiry, and clearing test definitions |
| Authentication persistence | `internal/app/auth.go` | fail-closed paths for lockout, login reset, blacklist, replay, session, and logout errors |
| Sessions and replay | authentication middleware and persistence tables | timestamp, duplicate request ID, inactivity, logout, and revoked-token negative paths |
| Strict redaction | `internal/service/`, `internal/platform/` | transform implementation and test definitions requiring removed content |
| Bates numbering | service/platform logic and Bates tables | visible-label and sequence-progression implementation/tests |
| Version comparison | comparison service and platform extraction | structured text, page, and bounding-box implementation/tests |
| Finalized immutability | workflow and mutator guards | every material mutation path rejects a Finalized document |
| Protected metadata | crypto helpers, ciphertext columns, masking paths | end-to-end protected storage and API exposure boundaries |
| Audit and notifications | audit and notification handlers/storage | mutating-flow, filtered-read, and notification-state definitions |
| Backup and restore | backup/restore service and platform adapters | dump, filesystem archive, restore mapping, failure handling, and state-verification definitions |
| Canonical repository layout | `tests/api/`, `tests/contracts/`, `public/index.html` | no legacy test directories or duplicate served HTML |
| Path and contamination audit | `ci/source_inventory.py` | all tracked files, case collisions, near duplicates, non-ASCII, whitespace, controls, mixed conventions, and explicit exceptions |
| Truthful local report | `run_tests.sh`, `ci/run_tests_contract_check.sh` | report coverage equals generated stage rows and skipped stages cannot pass |
| CI admission safety | `.github/workflows/ci.yml` | shared target key, active-run collapse, pre-checkout admission, no runner sleep, complete pagination, same-revision cooldown, failed-revision latch, new-revision admission, exact one-time unlock, and no post-failure artifact step |
| Documentation truth | `README.md`, `docs/`, `ci/docs_consistency_check.sh` | paths, configuration, commands, and static claims match current source |
| Static inventory retention | `.github/workflows/ci.yml`, `ci/source_inventory.py` | successful static artifact tied to the exact revision |
| Full-regression definition | `ci/run_full_regression.sh` | complete stage list, fail-fast propagation, truthful summary, revision fields, and artifact structure |

## Acceptance conditions

Static acceptance completes from current source, test definitions, schemas, migrations, configuration, workflows, deployment definitions, documentation, comments, and repository structure. It does not wait for or require:

- deployment or restart execution;
- authentication fault injection;
- stateful RBAC, workflow, PDF, audit, notification, backup, or restore execution;
- browser interaction execution;
- CI completion; or
- a complete-regression artifact.

Existing external artifacts may be described as optional read-only context for the exact revision and scope they executed. Their absence does not change the static verdict, and a reviewer must not run project code, scripts, tests, builds, containers, databases, browsers, deployments, or CI to create them.

A reviewer report is a static review summary, never a runtime artifact.

## Generated API documentation

Route-level Swaggo annotations under `internal/app/swagger_*.go` are authoritative. Generated files under `docs/swagger/` are created by supported execution entrypoints before compilation or generated-contract checking. Their presence or absence during static review does not authorize reviewer generation.

## Compatibility notes

Compatibility columns may remain where migration safety requires them, but protected plaintext must not be the source of truth. Sensitive source values belong in ciphertext columns; deterministic lookup values must be explicitly limited to lookup use.
