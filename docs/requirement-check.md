# Requirement Check

This document maps prompt requirements to the current implementation. It is intentionally honest: several compliance-grade PDF and test-coverage requirements remain incomplete.

## Legend

| Status | Meaning |
|---|---|
| Complete | Implemented in code and documented |
| Partial | Partly implemented, but not enough for final acceptance |
| Planned | Not implemented yet |
| Not applicable | Outside current requested scope |

## Stack and Deployment

| Requirement | Status | Evidence |
|---|---|---|
| Pure backend project | Complete | README and design docs state UI is only a backend test aid |
| Go backend | Complete | `cmd/server`, `internal/app` |
| Echo framework | Complete | `internal/app/server.go` |
| sqlx database access | Complete | database and handlers use sqlx |
| PostgreSQL metadata store | Complete | `migrations/001_schema.sql` |
| Local filesystem PDF storage | Complete | document versions store file paths |
| Single-container Docker deployment | Complete | `Dockerfile`, `docker-compose.yml`, `scripts/entrypoint.sh` |
| Docker builder, no local Go dependency | Complete | `Dockerfile`, `docs/docker-builder.md` |
| Swaggo/OpenAPI support | Partial | static YAML exists; generated Swaggo files were not produced because no commands were run |
| Backend test UI | Complete | `public/manual-test.html` |

## Authentication and Session Security

| Requirement | Status | Evidence |
|---|---|---|
| Local username/password auth | Complete | `POST /api/auth/login` |
| bcrypt password hashing | Complete | seed/user creation paths use bcrypt |
| Lock after 5 failed attempts | Complete | auth handler tracks failed attempts |
| JWT | Complete | login issues signed JWT |
| 8-hour inactivity timeout | Complete | server-side sessions use `last_seen_at` |
| JWT blacklist/logout | Complete | logout inserts into blacklist |
| Request timestamp check | Complete | auth middleware validates request timestamp |
| Replay guard | Complete | request IDs are tracked per token |
| AES-256 column-level encryption | Partial | AES-GCM helper exists and annotation comment/redaction reason are encrypted; coordinate fields are still numeric plaintext |
| Contextual masking | Partial | password hashes are hidden; broader contextual masking is incomplete |

## RBAC and Object Access

| Requirement | Status | Evidence |
|---|---|---|
| Admin role | Complete | admin routes under `/api/admin` |
| Editor role | Complete | upload, redaction, Bates, rollback, finalize routes |
| Reviewer role | Complete | annotation/review routes |
| Route-level RBAC | Complete | `requireRole` middleware |
| Admin not automatically Editor | Complete | Admin is not included in Editor-only routes |
| Object-level document read authorization | Complete | `access.go` scopes list/get/file/versions |
| Object-level mutation authorization | Partial | redaction, annotation, Bates, workflow, finalize, and compare use object checks; every mutation still needs API denial coverage |

## Document Lifecycle

| Requirement | Status | Evidence |
|---|---|---|
| Single PDF upload | Complete | `POST /api/documents` |
| Batch import up to 250 | Complete | `batchUploadDocuments` persists each file through the shared upload path |
| Metadata in PostgreSQL | Complete | `documents`, `document_versions` |
| Local file pointer | Complete | `document_versions.file_path` |
| 200 MB limit | Complete | upload/inspection enforce max bytes |
| 500 page limit | Complete | local PDF inspection enforces max page count |
| 50 version ceiling | Partial | redaction/Bates check version ceiling; all version-producing paths still need tests |
| Version rollback | Complete | validates version, rejects Finalized, updates current version, writes audit |
| Document comparison | Partial | compares real version files and binary metadata; response declares `comparison_kind=binary_metadata`, `text_diff_supported=false`, and `bbox_supported=false` |

## Workflow

| Requirement | Status | Evidence |
|---|---|---|
| Draft -> Under Review -> Redaction Pending -> Approved -> Finalized | Complete | `NextWorkflowStatus`, transition handler |
| Invalid transition rejection | Complete | transition handler enforces next status |
| Finalized immutable | Partial | major mutation paths reject Finalized; full route-by-route tests still needed |
| Workflow audit | Complete | transition writes audit |
| Workflow notification | Complete | transition uses notification helper |

## Redaction

| Requirement | Status | Evidence |
|---|---|---|
| Redaction proposal metadata | Complete | redaction table and proposal endpoint |
| Redaction reason encryption | Complete | reason is encrypted before storage |
| Editor confirmation | Complete | confirm route is Editor-only and object-scoped |
| New version after confirmation | Complete | confirm creates a document version |
| Audit records | Complete | proposal and confirmation write audit |
| Forensic burn-in / true content removal | Planned | current PDF marker helper only appends marker bytes and is not compliant permanent PDF content removal |
| Coordinate encryption | Planned | x/y/width/height remain numeric plaintext |

## Annotation

| Requirement | Status | Evidence |
|---|---|---|
| Reviewer-only creation | Complete | route role restriction plus non-Draft object scope |
| Allowed type validation | Complete | create handler calls `IsValidAnnotationType` |
| Comment length limit | Complete | 2000 character limit enforced |
| Disposition validation | Complete | create/update validate disposition |
| Comment encryption | Complete | annotation comment is encrypted before storage |
| Disposition update | Complete | `PATCH /api/annotations/:id/disposition` |
| Mention notification | Complete | `mentions.go` parses `@username` and creates local notifications through `notifyMentionedUsers` |

## Bates Numbering

| Requirement | Status | Evidence |
|---|---|---|
| Prefix/suffix/padding/start validation | Complete | Bates handler validates and normalizes inputs |
| Persistent job record | Complete | `bates_jobs` row is inserted |
| New document version | Partial | Bates route creates a new PDF version, but current processing only appends marker bytes and does not draw visible page numbering |
| Actual page-visible Bates numbering | Planned | no PDF page drawing engine is integrated |
| Batch sequence allocation | Planned | no cross-document sequence allocator exists |

## Audit

| Requirement | Status | Evidence |
|---|---|---|
| Audit helper writes logs | Complete | `domain_events.go` writes audit rows |
| Main mutation audit | Partial | many main flows write audit; coverage should be enforced route-by-route |
| Filterable query | Complete | `auditLogsFiltered` supports actor/document/action/request/source/date filters and route is Admin-only |
| Indefinite retention | Complete | no deletion policy exists |

## Notifications

| Requirement | Status | Evidence |
|---|---|---|
| In-app queue | Complete | `notifications` table |
| Workflow notification | Complete | workflow transition calls `notifyUser` |
| 500 unread ceiling | Complete | `createNotification` marks oldest unread when ceiling is reached |
| Per-user query | Complete | `/api/notifications` uses principal user ID |
| Read acknowledgement | Partial | `/api/notifications/:id/read` updates by user ID, but missing-row behavior still needs stronger API coverage |
| Annotation mention notification | Complete | annotation creation calls `notifyMentionedUsers` before returning |
| Admin editable templates | Partial | templates can be listed; edit endpoint is incomplete |

## Backup and Recovery

| Requirement | Status | Evidence |
|---|---|---|
| Backup metadata | Complete | `backup_jobs` table and endpoint |
| Local backup artifact output | Partial | current run endpoint writes a JSON metadata snapshot and should report `restore_supported=false`; it is not restore-capable |
| Real pg_dump execution | Planned | direct `pg_dump` implementation was not added; external command execution was blocked in this editing environment |
| Filesystem snapshot | Planned | not implemented |
| Restore workflow | Planned | not implemented |
| PITR docs | Complete | `docs/pitr.md` |

## Testing and Acceptance

| Requirement | Status | Evidence |
|---|---|---|
| Unit tests | Partial | root `run_tests.sh` now includes `go test ./...`, but handler/database integration tests are still limited |
| API tests | Partial | login/RBAC/upload/admin-read tests exist; workflow/redaction/Bates/compare/backup coverage remains insufficient |
| No SKIP-as-success suites | Partial | known SKIP suites were removed; coverage is still not near 90% |
| Docker acceptance path | Complete | `scripts/docker_acceptance.sh` exists |
| Sample PDF/CSV | Complete | `testdata/` fixtures exist |

## Current Blocking Gaps

1. True forensic PDF redaction burn-in is not implemented.
2. Page-visible Bates numbering is not implemented.
3. Real `pg_dump`, filesystem snapshot backup, and restore workflow are not implemented.
4. Compare API does not perform text-level PDF diff with real page/bbox extraction.
5. API endpoint coverage remains below the requested threshold.
6. Handler/database integration tests are still limited.
