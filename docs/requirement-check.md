# Requirement Check

This document maps required behavior to implementation and evidence paths. A status claim is valid only for the exact revision named by its retained artifact or verified source-tree equivalence record.

## Verification model

| Layer | Purpose | Evidence boundary |
|---|---|---|
| Static reviewer acceptance | Inspect source and pre-existing evidence without executing project or CI work | Missing runtime or interaction evidence is `NOT VERIFIED` |
| GitHub static workflow | Detect source, layout, configuration, documentation, workflow, and naming contradictions | Static evidence only |
| Local report | Record stages actually run by `run_tests.sh` | Claims only rows in its generated `results.tsv` |
| Docker/API/browser acceptance | Exercise generated deployment and critical stateful flows | Runtime evidence for the exact execution |
| Complete regression | Run every defined stage sequentially and retain a successful artifact | Full evidence only when `summary.json` passes for the tested revision |

The sole GitHub workflow uses one target namespace, target concurrency with active-run collapse, checkout-free admission, full history pagination, immediate denial rather than runner sleep, a failed-revision latch, and one-time reviewed unlocks tied to the exact failed run.

Repository YAML cannot prevent GitHub from first creating a workflow-run object. Claims of true pre-dispatch blocking require separate platform-level evidence.

## Requirement and evidence paths

| Area | Implementation path | Required evidence |
|---|---|---|
| Complete local configuration | `scripts/deploy.sh`, `docker-compose.yml`, `Dockerfile`, `internal/app/config.go`, `scripts/entrypoint.sh` | installation-specific ports, identity, paths, credentials, image args, and validation; no product fallback |
| Initial administrator | `internal/app/database.go` | empty generated normal-mode volume creates one Admin; restart without bootstrap values preserves it |
| Host exposure | `scripts/deploy.sh` | IPv4 loopback selection and an initially unoccupied host-port probe; Docker bind remains final runtime evidence |
| Acceptance fixtures and UI | `internal/app/config.go`, `internal/app/server.go`, `public/index.html` | acceptance mode requires execution-scoped fixtures; normal mode does not serve `/ui/`; only one HTML surface exists |
| Rolling login lockout | `internal/app/auth.go`, `migrations/002_login_attempt_window.sql` | `tests/api/test_auth_lockout_docker.sh` defines window expiry, fifth-attempt lock, recovery, and state clearing |
| Authentication persistence | `internal/app/auth.go` | stateful fault evidence for lockout, login reset, blacklist, replay, session, and logout failures |
| Sessions and replay | authentication middleware and persistence tables | timestamp, duplicate request ID, inactivity, logout, and revoked-token negative flows |
| Strict redaction | `internal/service/`, `internal/platform/` | extracted text no longer contains redacted content |
| Bates numbering | service/platform logic and Bates tables | visible label and cross-document sequence progression |
| Version comparison | comparison service and platform extraction | text changes with page and bounding-box data |
| Finalized immutability | workflow and mutator guards | every material mutation rejects a Finalized document |
| Protected metadata | crypto helpers, ciphertext columns, masking paths | storage contract and API exposure evidence |
| Audit and notifications | audit and notification handlers/storage | mutating flow followed by filtered read and notification state verification |
| Backup and restore | backup/restore service and platform adapters | PostgreSQL dump, filesystem archive, restore, and subsequent state verification |
| Canonical repository layout | `tests/api/`, `tests/contracts/`, `public/index.html` | no legacy test directories or duplicate served HTML |
| Path and contamination audit | `ci/source_inventory.py` | retained manifest covering all tracked files, case collisions, near duplicates, non-ASCII, whitespace, controls, mixed conventions, and explicit exceptions |
| Truthful local report | `run_tests.sh`, `ci/run_tests_contract_check.sh` | report coverage exactly equals generated stage rows |
| CI admission safety | `.github/workflows/ci.yml` | shared target key, active-run collapse, checkout-free admission, no runner sleep, complete pagination, failure latch, exact one-time unlock, no post-failure artifact step |
| Documentation truth | `README.md`, `docs/`, `ci/docs_consistency_check.sh` | paths, configuration, commands, and evidence statements match current source |
| Static inventory retention | `.github/workflows/ci.yml`, `ci/source_inventory.py` | successful static artifact tied to the exact revision |
| Full regression | `ci/run_full_regression.sh` | `summary.json` reports `overall_status=passed`, every recorded stage is zero, and artifact identifies the tested revision |

## Acceptance conditions

Static implementation inspection does not create runtime evidence. Full runtime acceptance for a revision still requires pre-existing:

- generated normal-mode deployment and restart evidence;
- rolling-window and authentication fault-injection evidence;
- stateful RBAC, workflow, PDF, audit, notification, backup, and restore evidence;
- browser interaction evidence for any claim beyond rendering; and
- a complete-regression artifact tied to the exact revision.

A static reviewer must not run project code, scripts, tests, builds, containers, databases, browsers, deployments, or CI to obtain those artifacts.

A reviewer report is a review summary, never a test artifact.

## Generated API documentation

Route-level Swaggo annotations under `internal/app/swagger_*.go` are authoritative. Generated files under `docs/swagger/` are created by supported execution entrypoints before compilation or generated-contract checking. Their presence or absence during static review does not authorize reviewer generation.

## Compatibility notes

Compatibility columns may remain where migration safety requires them, but protected plaintext must not be the source of truth. Sensitive source values belong in ciphertext columns; deterministic lookup values must be explicitly limited to lookup use.
