# Manual Completion Actions

This document lists the remaining work that could not be safely or honestly completed by automated repository edits alone. It is intended as a handoff checklist for the maintainer.

## Current status

The project is stronger than the initial skeleton, but it is still not a final compliance-grade delivery. The following areas remain blocking for final acceptance:

1. forensic PDF redaction
2. page-visible Bates numbering
3. real backup execution
4. text-level PDF comparison
5. broader API E2E coverage
6. handler/database integration tests

## 1. Forensic PDF redaction

### Current implementation

The current redaction flow stages regions, confirms redactions, creates a new document version, and writes audit records. However, the PDF transform still appends marker content rather than removing underlying PDF content.

### Required maintainer action

Replace the append-style transform with a real PDF content-removal implementation. The accepted implementation must remove the selected content from the resulting PDF so it cannot be recovered through text extraction or object-stream inspection.

### Acceptance check

Use a test PDF containing known sensitive text. After confirming a redaction, extract text from the resulting PDF and verify that the sensitive text is absent. Also verify that a new document version and audit row are created.

## 2. Page-visible Bates numbering

### Current implementation

Bates processing now creates a new document version and audit record. This is better than metadata-only behavior, but it does not draw visible Bates numbers onto each page.

### Required maintainer action

Add a PDF page-writing implementation that renders Bates text on every page using the configured prefix, suffix, zero padding, and start number. Add batch sequence allocation if cross-document Bates numbering is required.

### Acceptance check

Upload a multi-page PDF, apply Bates numbering, download the new version, and visually or programmatically verify that every page contains the expected label sequence.

## 3. Real backup execution

### Current implementation

The current backup path creates a local marker artifact and records a backup job. It is still not a real database dump or filesystem snapshot.

### Required maintainer action

Implement actual database backup execution in the runtime environment. The expected shape is:

- run a PostgreSQL logical dump inside the container or bundled runtime
- store the dump under the configured backup directory
- include stored PDF files or a separate filesystem archive/snapshot
- record success or failure in `backup_jobs`
- expose the resulting artifact path in the API response

### Acceptance check

Trigger the backup endpoint, confirm the job reaches a completed state, confirm a non-empty database dump artifact exists, and confirm the PDF storage artifact or snapshot exists.

## 4. Text-level PDF comparison

### Current implementation

The compare endpoint reads both version files and reports binary and metadata differences. It does not extract PDF text or return real page and bounding-box level changes.

### Required maintainer action

Add a local PDF text extraction and comparison layer. The API should return added, removed, and modified segments with page numbers and bounding boxes when available.

### Acceptance check

Create two known PDF versions with a controlled text change. Compare them and verify that the API returns the changed text segment with the correct page context.

## 5. API E2E coverage

### Current implementation

The API tests cover login bootstrap, selected RBAC, upload, document object read checks, admin read endpoints, audit list, and backup job list. Coverage is still below the required threshold.

### Required maintainer action

Add API tests for:

- logout and token revocation behavior
- request replay rejection
- stale timestamp rejection
- batch upload
- rollback
- full workflow chain
- invalid workflow transition
- finalize
- post-finalization mutation rejection
- redaction proposal, list, and confirm
- annotation create, list, disposition update, invalid type, and mention notification
- Bates new version creation
- compare valid and invalid versions
- audit filters
- notification read acknowledgement
- backup run and artifact validation

### Acceptance check

Run the root test entrypoint in the Docker acceptance environment. The output should contain real pass/fail assertions, not skipped suites that exit successfully.

## 6. Handler and database integration tests

### Current implementation

The project has helper-level Go unit tests for rules, workflow, PDF inspection, crypto, access control, and mention parsing. These tests do not prove database side effects.

### Required maintainer action

Add handler or API integration tests that assert persistent side effects:

- document and version rows are created
- audit rows are created for mutating actions
- notification rows are created for workflow and mentions
- backup job rows are created with correct status
- object-level denials are enforced
- finalized documents reject mutations

### Acceptance check

Use an isolated test database or Docker acceptance environment. Tests should fail if the expected database rows or file artifacts are missing.

## 7. Documentation status

`docs/requirement-check.md` should remain conservative. Do not mark the following as Complete until the acceptance checks above pass:

- forensic redaction
- visible Bates numbering
- real database and filesystem backup
- text-level PDF comparison
- full API endpoint coverage
- handler/database integration testing
