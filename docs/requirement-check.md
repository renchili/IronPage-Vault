# Requirement Check

This document maps required behavior to implementation and evidence paths. A status claim is valid only for the exact revision named by its generated test artifact.

## Verification model

| Layer | Purpose | Evidence boundary |
|---|---|---|
| Repository contracts | Detect source, layout, configuration, documentation, and workflow contradictions | Static evidence only |
| Local report | Record stages actually run by `run_tests.sh` | Claims only the rows in its generated `results.tsv` |
| Docker/API/browser acceptance | Exercise generated deployment and critical stateful flows | Runtime evidence for the exact execution |
| Complete regression | Run every defined stage sequentially and retain a successful artifact | Full evidence only when `summary.json` passes for the tested SHA |
| GitHub workflow | Serialize validation and enforce cooldown/failure latch | `.github/workflows/ci.yml` is the sole workflow |

The workflow handles pull requests, merge groups, pushes to `main`, and reviewed manual replays through one repository-and-target concurrency group. It stops at the first failure and performs no later artifact action after that failure.

## Requirement and evidence paths

| Area | Implementation path | Required evidence |
|---|---|---|
| Complete local configuration | `scripts/deploy.sh`, `docker-compose.yml`, `Dockerfile`, `internal/app/config.go`, `scripts/entrypoint.sh` | generated installation-specific ports, identity, paths, credentials, image args, and validation; no product fallback |
| Initial administrator | `internal/app/database.go` | empty generated normal-mode volume creates one Admin; restart without bootstrap values preserves it |
| Acceptance fixtures and UI | `internal/app/config.go`, `internal/app/server.go`, `public/index.html` | acceptance mode requires execution-scoped fixtures; normal mode does not serve `/ui/`; only one HTML surface exists |
| Rolling login lockout | `internal/app/auth.go`, `migrations/002_login_attempt_window.sql` | `tests/api/test_auth_lockout_docker.sh` proves window expiry, fifth-attempt lock, recovery, and state clearing |
| Authentication persistence | `internal/app/auth.go` | fault injection proves lockout, login reset, blacklist, replay, session, and logout errors fail closed |
| Sessions and replay | authentication middleware and persistence tables | timestamp, duplicate request ID, inactivity, logout, and revoked-token negative flows |
| Strict redaction | `internal/service/`, `internal/platform/` | extracted text no longer contains redacted content |
| Bates numbering | service/platform logic and Bates tables | visible label and cross-document sequence progression |
| Version comparison | comparison service and platform extraction | text changes with page and bounding-box data |
| Finalized immutability | workflow and mutator guards | every material mutation rejects a Finalized document |
| Protected metadata | crypto helpers, ciphertext columns, masking paths | storage contract and API exposure checks |
| Audit and notifications | audit and notification handlers/storage | mutating flow followed by filtered read and notification state verification |
| Backup and restore | backup/restore service and platform adapters | real PostgreSQL dump, filesystem archive, restore, and subsequent state verification |
| Canonical repository layout | `tests/api/`, `tests/contracts/`, `public/index.html` | no legacy test directories or duplicate served HTML |
| Truthful local report | `run_tests.sh`, `ci/run_tests_contract_check.sh` | report coverage exactly equals generated stage rows |
| CI execution safety | `.github/workflows/ci.yml`, `ci/ci_execution_guard.py`, `ci/run_full_regression.sh` | one shared lock, ten-minute cooldown, failed-SHA latch, sequential first-error stop, no post-error upload |
| Documentation truth | `README.md`, `docs/`, `ci/docs_consistency_check.sh` | paths, configuration, commands, and evidence statements match current source |
| Full regression | `ci/run_full_regression.sh` | `summary.json` reports `overall_status=passed`, every recorded stage is zero, and artifact identifies the tested SHA |

## Acceptance conditions

Static implementation repair does not create runtime evidence. Full acceptance for a revision still requires:

- generated normal-mode deployment and restart evidence;
- rolling-window and authentication fault-injection evidence;
- stateful RBAC, workflow, PDF, audit, notification, backup, and restore evidence;
- browser interaction evidence for any claim beyond rendering; and
- a complete-regression artifact tied to the exact revision.

A reviewer report is a review summary, never a test artifact.

## Generated API documentation

Route-level Swaggo annotations under `internal/app/swagger_*.go` are authoritative. Generated files under `docs/swagger/` must be produced by supported entrypoints before compilation or contract checking.

## Compatibility notes

Compatibility columns may remain where migration safety requires them, but protected plaintext must not be the source of truth. Sensitive source values belong in ciphertext columns; deterministic lookup values must be explicitly limited to lookup use.
