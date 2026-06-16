# Strict Acceptance Semantics

This patch removes fallback-as-success behavior called out by static review.

- Redaction acceptance path uses strict raster burn-in only. Missing `pdftoppm`, Pillow, or reportlab returns an error instead of overlay/manifest success.
- Bates acceptance path requires successful reportlab+pypdf visible overlay. Manifest fallback is not used by the service path.
- Backup acceptance path requires both `pg_dump_custom` and `tar` artifacts. Missing or failed artifacts return error.
- Restore requires readable database dump and filesystem tar paths and returns success only after both `pg_restore` and `tar` complete.
- API tests validate response body fields for backup mode/status and reject empty restore bodies.
- Compare test creates its own second version and no longer depends on external `LEFT_VERSION_ID` / `RIGHT_VERSION_ID`.
