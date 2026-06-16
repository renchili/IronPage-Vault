# Requirement Check

This document maps the requested acceptance items to the current implementation after the final hardening pass.

## Current Blocking Gaps

None tracked in this document after the final hardening pass.

## Final hardening evidence

| Area | Status | Evidence |
|---|---|---|
| Compile blocker | Complete | `internal/app/workflows.go` malformed compare block was removed before this pass |
| Redaction processing | Complete | Service path requires raster burn-in for successful redaction output |
| Bates processing | Complete | Service path requires successful visible page overlay for Bates output |
| Backup success semantics | Complete | Backup API success requires database dump and filesystem archive artifacts |
| Restore success semantics | Complete | Restore API requires artifact paths and successful restore/archive extraction before success |
| Redaction coordinate storage | Complete | Request geometry is written to encrypted coordinate columns; legacy numeric columns are zero placeholders |
| Redaction API exposure | Complete | Redaction list response omits coordinate and reason fields |
| Compare API test chain | Complete | Self-contained compare test creates a second version before comparing |
| API response assertions | Complete | Static review tests validate backup response fields and empty restore rejection |
| CI | Complete | `.github/workflows/ci.yml` runs `go test ./...`, static rules, and Docker build |
| Static regression guards | Complete | `unit_tests/test_rules.sh` checks strict processing entrypoints, coordinate storage rules, CI, restore strictness, and self-contained compare coverage |

## Notes

The application keeps the original API shape and database schema compatibility where needed. For redaction geometry, existing numeric columns remain in the schema for compatibility, but the application writes zero placeholders and uses encrypted columns as the source of truth.

## Latest static recheck closure

The latest static recheck reject items are addressed by `API_tests/test_admin_ops.sh`, `API_tests/test_compare_acceptance.sh`, `API_tests/test_finalized_immutability.sh`, `API_tests/test_pdf_content_acceptance.sh`, and `API_tests/test_notification_mention_side_effect.sh`.

- Strict restore empty-body expectation is aligned to HTTP 400.
- Compare acceptance no longer depends on external version ID environment variables.
- Finalized immutability test now creates its own document and walks the full workflow chain before finalization.
- Redaction content validation checks that target text is not extractable from the output PDF.
- Bates content validation checks that the expected label is extractable from the output PDF.
- Backup validation checks strict modes and artifact file existence.
- Audit filter validation checks returned rows match the requested action.
- Mention validation checks annotation mention notification side effects.

## Complete static recheck closure

The remaining static recheck gaps are covered by the following additional checks:

- Strict dependency failure behavior is guarded by `API_tests/test_strict_dependency_failures.sh` and platform strict unit tests.
- Bates sequence allocation across multiple documents is covered by `API_tests/test_bates_sequence_multi_doc.sh`.
- Backup job and artifact evidence is strengthened in `API_tests/test_admin_ops.sh`.
- Reject-condition regressions are guarded by `unit_tests/test_structure_rules.sh`.
- `run_tests.sh` invokes the new structure, strict dependency, and Bates sequence coverage.
