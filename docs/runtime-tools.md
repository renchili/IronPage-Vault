# Runtime Tools

IronPage Vault uses external command-line tools for the full acceptance flows added in the implementation bundle.

The Docker image installs:

- `pg_dump` and `pg_restore` from the PostgreSQL base image for database backup and restore.
- `tar` for filesystem storage snapshots and restore.
- `pdftotext` from `poppler-utils` for best-effort PDF text extraction during compare.
- `python3` and `pypdf` for best-effort PDF object-graph rewrite paths used by Bates/redaction processing.

If any optional runtime tool is unavailable in a custom deployment, the application records a fallback mode in the API response or generated manifest instead of silently claiming the stronger processing mode.
