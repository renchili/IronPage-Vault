# Implementation Status

Implemented in this acceptance fix bundle:

- Admin-created user secrets require minimum length, a digit, and a special character.
- Audit log filter query construction lives in `internal/store` with unit coverage.
- Notification read acknowledgement verifies row ownership and can return 404.
- Backup writes metadata plus best-effort `pg_dump` and filesystem `tar` artifacts.
- Restore has an Admin route and uses `pg_restore` / `tar` when available.
- Bates processing uses `reportlab+pypdf` to draw visible page labels.
- Redaction processing uses `reportlab+pypdf` to draw filled black rectangles over target regions.
- Compare attempts `pdftotext` extraction and falls back to printable-byte extraction.
- Runtime Docker image installs the required PDF/text/backup tools.
- Root `run_tests.sh` directly runs `go test ./...`.

Known remaining limitation:

- Text compare is text-level and binary fallback, not true bbox-level semantic PDF diff.
- Legal-grade redaction should be validated with representative PDFs before production use.
