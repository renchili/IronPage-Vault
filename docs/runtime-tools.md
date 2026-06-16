# Runtime Tools

The Docker image includes the runtime tools required by the full acceptance flows:

- `pg_dump` / `pg_restore` from the PostgreSQL base image for database backup and restore.
- `tar` for filesystem storage snapshot and restore.
- `pdftotext` from `poppler-utils` for best-effort text extraction in compare.
- `python3`, `pypdf`, and `reportlab` for PDF object rewriting plus visible Bates/redaction overlays.
