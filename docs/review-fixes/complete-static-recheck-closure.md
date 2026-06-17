# Complete Static Recheck Closure

This patch closes the remaining static recheck findings that were not covered by the earlier test-blocker patch.

## Added strict negative coverage

- `API_tests/test_strict_dependency_failures.sh` verifies service paths use strict entrypoints and that fallback entrypoints are not used by acceptance service paths.
- `internal/platform/pdf_strict_test.go` and `internal/platform/backup_strict_test.go` assert strict paths fail rather than succeed when required inputs or artifacts are absent.

## Added sequence and artifact coverage

- `API_tests/test_bates_sequence_multi_doc.sh` creates two documents and verifies Bates sequence allocation increases across documents.
- `API_tests/test_admin_ops.sh` now also verifies backup job output includes `full_backup`, and validates artifact paths.

## Added structural regression coverage

- `unit_tests/test_structure_rules.sh` guards against reintroducing the exact reject conditions from the report:
  - legacy restore 202 expectation,
  - external compare version environment dependency,
  - finalized test without Approved state,
  - missing PDF content tests,
  - missing backup artifact checks,
  - missing Bates sequence coverage,
  - missing strict dependency coverage.

## CI note

This branch includes follow-up fixes for Go compile and mention parsing failures found by CI.
