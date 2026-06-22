# Requirement Check

This document maps the requested acceptance items to the implementation and verification model.

## Verification model

The repository intentionally separates PR gates, merge-candidate regression, post-merge evidence, and local/manual verification.

| Layer | Purpose | What runs |
|---|---|---|
| Pull request CI | Prevent obviously unsafe changes before merge | change-impact analysis, gofmt, targeted `go vet`, targeted `go test`, generated Swagger contract, shell syntax, Docker build, CI-flow contract probes, local entrypoint contract probes |
| Merge queue regression | Verify the temporary merge result before it reaches `main` | reusable full regression on `merge_group` |
| Post-merge regression evidence | Retain evidence for product/runtime/regression-impacting changes already merged to `main` | reusable full regression with logs, JSON summary, Markdown summary, and retained artifacts |
| Local/manual replay | Developer reproduction path | `./run_tests.sh`, which generates Swagger artifacts before compiling/tests |

PR CI is impact-based. It does not use `run_tests.sh` as the pull-request pass/fail source. Runtime/API Docker acceptance is executed by the reusable full regression flow on merge candidates, post-merge evidence runs, or manual workflow dispatch.

## Current fixed static recheck blockers

- `run_tests.sh` now generates `docs/swagger` before `go test -mod=mod ./...`, so a fresh checkout does not depend on committed generated Swagger artifacts.
- `ci/run_tests_contract_check.sh` verifies the local test entrypoint can generate Swagger artifacts from a clean state.
- `ci/swagger_contract_check.sh` verifies every route-level `@Router` annotation appears in generated `docs/swagger/swagger.yaml` and checks key response contracts.
- PR CI runs the local entrypoint contract when `run_tests.sh`, Swagger generation, or Swagger artifacts change.
- PR CI runs the regression flow contract when full-regression workflow or runner logic changes.
- `run_tests.sh` and `API_tests/lib.sh` preserve token availability across API scripts.
- Mention notification test uses `Sticky note`, a supported annotation type.
- Bates apply response returns `start_number`, matching the multi-document sequence test.

## Final hardening evidence

| Area | Status | Evidence |
|---|---|---|
| Compile blocker | Complete | `internal/app/workflows.go` malformed compare block was removed |
| Redaction processing | Complete | Service path requires raster burn-in for successful redaction output |
| Bates processing | Complete | Service path requires successful visible page overlay for Bates output |
| Backup success semantics | Complete | Backup API success requires database dump and filesystem archive artifacts |
| Scheduled backup evidence | Complete | `ci/scheduled_backup_contract_check.sh` verifies scheduler startup, interval gating, strict artifact worker path, job row, and audit event |
| Restore success semantics | Complete | Restore API requires artifact paths and successful restore/archive extraction before success |
| Redaction coordinate storage | Complete | Request geometry is written to encrypted coordinate columns; legacy numeric columns are zero placeholders |
| Redaction API exposure | Complete | Redaction list response omits coordinate and reason fields |
| Sensitive metadata storage matrix | Complete | `docs/metadata-security.md` and `ci/metadata_storage_check.sh` cover redaction reason, redaction geometry, and annotation comment storage/exposure rules |
| Compare API test chain | Complete | Self-contained compare test creates a second version before comparing |
| Compare content accuracy | Complete | `internal/service/compare_test.go` asserts added, removed, and modified text blocks contain the expected changed text |
| API token orchestration | Complete | `run_tests.sh` and `API_tests/lib.sh` preserve token availability across scripts |
| Mention notification test | Complete | Test uses `Sticky note` |
| Bates sequence contract | Complete | Bates apply response includes `start_number` |
| PR CI | Complete | Impact-based static/build gates, generated Swagger contract, local entrypoint contract, and CI-flow contract probes |
| Full regression | Complete | Reusable full regression runs generated Swagger, generated Swagger route coverage, scheduled backup contract, metadata storage contract, full gofmt, `go vet ./...`, race tests, Docker build, and Docker acceptance |
| Static regression guards | Complete | `unit_tests/test_rules.sh` and `unit_tests/test_structure_rules.sh` guard reject-condition regressions |

## Swagger generated artifact policy

Generated Swagger artifacts under `docs/swagger` are treated as generated output, not the primary source of truth. Route-level Swaggo annotations under `internal/app/swagger_*.go` are the source of truth. Every compile/test entrypoint that imports `ironpage-vault/docs/swagger` must generate or stub the package before compiling.

See `docs/swagger-artifacts.md` for the operational policy.

## Known limitations / separate product-scope follow-ups

No product-scope recheck limitations are currently tracked in this document.

## Notes

The application keeps the original API shape and database schema compatibility where needed. For redaction geometry, existing numeric columns remain in the schema for compatibility, but the application writes zero placeholders and uses encrypted columns as the source of truth.
