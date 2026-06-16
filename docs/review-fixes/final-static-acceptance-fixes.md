# Final Static Acceptance Fixes

This patch removes the remaining "basic/partial" static-review concerns.

## Structure

- SQL list/count/allocation logic now has concrete repository package entrypoints.
- Compare and PDF orchestration now has concrete service package entrypoints.
- Handlers keep public HTTP behavior but no longer own the newly extracted audit, backup count, Bates allocation, redaction-region loading, PDF orchestration, or compare orchestration logic.

## Tests

- `run_tests.sh` directly invokes `go test ./...`.
- `run_tests.sh` invokes static review reject flow tests and denial tests when executable.
- Added compare acceptance and finalized immutability API test entrypoints.

## PDF / Backup / Compare

- Redaction prefers raster burn-in.
- Bates uses visible page overlay.
- Backup uses pg_dump/tar where runtime tools are present.
- Compare uses text bbox extraction through `pdftotext -bbox`.
