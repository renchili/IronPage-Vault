# Requirement Check

This document maps the requested acceptance items to the implementation after the latest recheck orchestration fixes.

## Current Blocking Gaps

None tracked after fixing the latest static recheck blockers:

- `run_tests.sh` reloads token files created by `test_api_flow.sh` between API scripts.
- `API_tests/lib.sh` auto-loads saved token files when token env vars are missing.
- Mention notification test uses `Sticky note`, a supported annotation type.
- Bates apply response returns `start_number`, matching the multi-document sequence test.
- CI runs Go tests, static rules, Docker build, and Docker acceptance.

## Final hardening evidence

| Area | Status | Evidence |
|---|---|---|
| Compile blocker | Complete | `internal/app/workflows.go` malformed compare block was removed |
| Redaction processing | Complete | Service path requires raster burn-in for successful redaction output |
| Bates processing | Complete | Service path requires successful visible page overlay for Bates output |
| Backup success semantics | Complete | Backup API success requires database dump and filesystem archive artifacts |
| Restore success semantics | Complete | Restore API requires artifact paths and successful restore/archive extraction before success |
| Redaction coordinate storage | Complete | Request geometry is written to encrypted coordinate columns; legacy numeric columns are zero placeholders |
| Redaction API exposure | Complete | Redaction list response omits coordinate and reason fields |
| Compare API test chain | Complete | Self-contained compare test creates a second version before comparing |
| API token orchestration | Complete | `run_tests.sh` and `API_tests/lib.sh` preserve token availability across scripts |
| Mention notification test | Complete | Test uses `Sticky note` |
| Bates sequence contract | Complete | Bates apply response includes `start_number` |
| CI | Complete | Workflow runs Go tests, static rules, Docker build, and Docker acceptance |
| Static regression guards | Complete | `unit_tests/test_rules.sh` and `unit_tests/test_structure_rules.sh` guard reject-condition regressions |

## Notes

The application keeps the original API shape and database schema compatibility where needed. For redaction geometry, existing numeric columns remain in the schema for compatibility, but the application writes zero placeholders and uses encrypted columns as the source of truth.
