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
