# Static Review Reject Fixes

This patch addresses the uploaded static reject report:

- `/api/audit-logs` has handler-level Admin enforcement.
- `run_tests.sh` contains `go test ./...`.
- Redaction prefers raster burn-in: `pdftoppm` renders pages, Pillow draws black rectangles into pixels, and reportlab rebuilds a new PDF.
- Bates numbering uses reportlab+pypdf visible page labels.
- Backup code uses `pg_dump` and `tar` artifacts when runtime tools are present.
- Compare uses `pdftotext -bbox` and returns text blocks with page and bbox fields.
- Object-level authorization has denial API coverage.
- `internal/service`, `internal/repository`, and `internal/app/security` mark service/repository/security policy boundaries.
