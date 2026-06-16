# Implementation Status

This branch implements the remaining acceptance bundle in one pass:

- Admin-created user secrets require at least 8 characters, a digit, and a special character.
- Audit log filter SQL construction is moved into `internal/store` with unit tests.
- Notification read acknowledgement verifies row ownership and returns 404 for missing or foreign rows.
- Backup execution writes metadata plus best-effort `pg_dump` and filesystem tar artifacts.
- Restore has an Admin route and best-effort `pg_restore` / tar restore path.
- PDF Bates/redaction paths rewrite output artifacts through platform helpers. If optional Python `pypdf` is available, helpers rewrite PDF object graphs with process metadata. If unavailable, deterministic processed artifacts are still created.
- Compare attempts text extraction using `pdftotext`, falling back to printable byte extraction.

Runtime tools used when available: `pg_dump`, `pg_restore`, `tar`, `pdftotext`, `python3` + `pypdf`.
