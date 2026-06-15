# Testing Guide

IronPage Vault keeps testing material in two directories:

```text
unit_tests/
API_tests/
```

The root script `run_tests.sh` is the common acceptance entrypoint.

## Unit checks

Unit checks cover project structure and local fixtures:

- required metadata files
- migration files
- Docker files
- sample PDF files
- CSV manifests
- role definitions
- workflow definitions
- documentation files

## API checks

API checks cover the running service:

- health endpoint
- login endpoint
- protected endpoint behavior
- Admin-only endpoints
- Editor-only endpoints
- Reviewer-only endpoints
- document upload
- annotation flow
- workflow flow
- audit log query
- notification query
- backup job metadata

## Manual UI checks

The manual UI is served from:

```text
/ui/manual-test.html
```

It is a testing helper only. It does not replace API tests.

## Acceptance flow

1. Start the service with Docker Compose.
2. Verify `/healthz` returns OK.
3. Login as Admin, Editor, and Reviewer.
4. Upload a PDF as Editor.
5. Verify Reviewer cannot upload.
6. Create annotation as Reviewer.
7. Verify Editor cannot manage users.
8. Stage redaction as Editor.
9. Confirm redaction as Editor.
10. Apply Bates numbering as Editor.
11. Move workflow through the required chain.
12. Finalize the document.
13. Verify post-finalization mutation rejection.
14. Query audit logs.
15. Query notifications.
16. Create backup job metadata as Admin.

## Test data

Fixtures are local:

```text
testdata/pdfs/sample_contract.pdf
testdata/csv/batch_import_manifest.csv
```

No internet download should be required for acceptance.
