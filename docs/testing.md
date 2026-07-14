# Testing Guide

IronPage Vault keeps testing material in these main locations:

```text
unit_tests/
API_tests/
```

The root script `run_tests.sh` is the common local acceptance entrypoint.

## Unit and static checks

Unit and static checks cover repository structure, implementation contracts, and local fixtures, including:

- required metadata files.
- migration files.
- Docker files.
- sample PDF files.
- CSV manifests.
- role definitions.
- workflow definitions.
- documentation files.
- storage and protected-metadata contracts.

## API and runtime checks

API checks cover the running service. The acceptance suite includes real HTTP/API flows for:

- health endpoint.
- login endpoint.
- protected endpoint behavior.
- Admin-only endpoints.
- Editor-only endpoints.
- Reviewer-only endpoints.
- document upload.
- annotation flow.
- workflow flow.
- finalized immutability.
- redaction confirmation.
- PDF content-removal acceptance.
- Bates sequence acceptance.
- document comparison acceptance.
- audit log query.
- notification query and mention side effects.
- strict dependency failure behavior.
- backup and restore behavior.

## Manual backend UI checks

The manual backend testing UI is served from:

```text
/ui/
```

The source page is the backend test aid under `public/`. It is not a production frontend, not a formal fullstack deliverable, and not a replacement for API acceptance tests.

The screenshot acceptance script verifies:

- `/healthz` returns 200.
- `/ui/` returns 200.
- the static page contract is present.
- a headless browser can render and capture a screenshot.
- screenshot evidence and a manifest are written under the local acceptance artifacts directory.

The screenshot acceptance script does not prove:

- login button behavior.
- success and failure interaction flows.
- retry and recovery behavior.
- keyboard focus order.
- accessibility announcements.
- resulting API state after UI interaction.

Those require a separate browser interaction test.

## Acceptance flow

A representative full acceptance flow should:

1. Start the service with Docker Compose.
2. Verify `/healthz` returns OK.
3. Authenticate as the supported local roles through the API.
4. Upload a PDF as Editor.
5. Verify Reviewer cannot upload.
6. Create annotation as Reviewer.
7. Verify Editor cannot manage users.
8. Stage redaction as Editor.
9. Confirm redaction as Editor.
10. Verify target redacted content is not extractable from the confirmed output.
11. Apply Bates numbering as Editor.
12. Verify visible or extractable Bates evidence and cross-document sequence behavior where applicable.
13. Move workflow through the required chain.
14. Finalize the document.
15. Verify post-finalization mutation rejection.
16. Query audit logs.
17. Query notifications.
18. Create strict backup artifacts as Admin.
19. Verify restore-capable backup metadata and artifact paths.
20. Capture backend test UI screenshot evidence.

## Test data

Fixtures are local:

```text
testdata/pdfs/sample_contract.pdf
testdata/csv/batch_import_manifest.csv
```

No internet download should be required for acceptance.

## Evidence boundary

A historical full-regression run is useful evidence only for the exact SHA it ran against. Documentation must not claim a fresh current-HEAD full regression unless the run ID and current HEAD SHA are recorded.

The uploaded re-audit report records that current `main` had not received an exact-current-HEAD full regression at the time of the report. That remains a separate evidence gap until a new run is executed and linked.
