# Runtime Tools

The Docker image includes the runtime tools required by the full acceptance flows:

- `pg_dump` / `pg_restore` from the PostgreSQL base image for database backup and restore.
- `tar` for filesystem storage snapshot and restore.
- `pdftotext` from `poppler-utils` for best-effort text extraction in compare.
- `python3`, `pypdf`, and `reportlab` for PDF object rewriting plus visible Bates/redaction overlays.

PDF behavior:

- Bates processing draws visible labels near the bottom-right of each page when `python3+pypdf+reportlab` are available.
- Redaction processing draws filled black rectangles on the specified page regions when `python3+pypdf+reportlab` are available.
- If optional PDF drawing dependencies are missing in a custom runtime, the application returns a fallback mode and still writes a deterministic processed artifact with a manifest.

Backup behavior:

- Backup attempts `pg_dump --format=custom` and a `tar` snapshot of the storage directory.
- Restore attempts `pg_restore` and `tar` extraction when artifacts and tools are available.
