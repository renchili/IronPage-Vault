# Requirement Check

This document maps prompt requirements to the current strict acceptance implementation after compile and fallback-semantics fixes.

## Legend

| Status | Meaning |
|---|---|
| Complete | Implemented in code and documented |
| Partial | Implemented with an explicitly documented limitation |
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
| Swaggo/OpenAPI support | Complete | `docs/swagger/openapi.yaml` documents the public acceptance API paths |
| Backend test UI | Complete | `public/manual-test.html` |

## Authentication and Session Security

| Requirement | Status | Evidence |
|---|---|---|
| Local username/password auth | Complete | `POST /api/auth/login` |
| bcrypt password hashing | Complete | seed/user creation paths use bcrypt |
| Password complexity | Complete | user creation enforces minimum length and requires both a digit and a special character |
| Lock after 5 failed attempts | Complete | auth handler tracks failed attempts |
| JWT | Complete | login issues signed JWT |
| 8-hour inactivity timeout | Complete | server-side sessions use `last_seen_at` |
| JWT blacklist/logout | Complete | logout inserts into blacklist |
| Request timestamp check | Complete | auth middleware validates request timestamp |
| Replay guard | Complete | request IDs are tracked per token |
| AES-256 column-level encryption | Complete | AES-GCM encrypts annotation comments, redaction reasons, and redaction coordinate ciphertext mirrors |
| Contextual masking | Complete | sensitive password/session fields are hidden from JSON responses; encrypted content is stored as ciphertext |

## RBAC and Object Access

| Requirement | Status | Evidence |
|---|---|---|
| Admin role | Complete | admin routes under `/api/admin` |
| Editor role | Complete | upload, redaction, Bates, rollback, finalize routes |
| Reviewer role | Complete | annotation/review routes |
| Route-level RBAC | Complete | `requireRole` middleware |
| Admin not automatically Editor | Complete | Admin is not included in Editor-only routes |
| Object-level document read authorization | Complete | `access.go` scopes list/get/file/versions |
| Object-level mutation authorization | Complete | mutation handlers use document object checks and denial API coverage exists in `API_tests/test_acceptance_denials.sh` and `API_tests/test_static_review_reject_flows.sh` |

## Document Lifecycle

| Requirement | Status | Evidence |
|---|---|---|
| Single PDF upload | Complete | `POST /api/documents` |
| Batch import up to 250 | Complete | `batchUploadDocuments` persists each file through the shared upload path |
| Metadata in PostgreSQL | Complete | `documents`, `document_versions` |
| Local file pointer | Complete | `document_versions.file_path` |
| 200 MB limit | Complete | upload/inspection enforce max bytes |
| 500 page limit | Complete | local PDF inspection enforces max page count |
| 50 version ceiling | Complete | redaction/Bates check version ceiling |
| Version rollback | Complete | validates version, rejects Finalized, updates current version, writes audit |
| Document comparison | Complete | compare reads real version files, attempts `pdftotext -bbox`, and returns text blocks with page/bbox metadata |

## Workflow

| Requirement | Status | Evidence |
|---|---|---|
| Draft -> Under Review -> Redaction Pending -> Approved -> Finalized | Complete | `NextWorkflowStatus`, transition handler |
| Invalid transition rejection | Complete | transition handler enforces next status |
| Finalized immutable | Complete | mutation paths call `ensureMutable` or reject Finalized |
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
| Forensic burn-in / true content removal | Complete | service path requires raster burn-in with `pdftoppm` + Pillow + reportlab; missing dependencies or raster failure returns an error |
| Coordinate encryption | Complete | coordinate ciphertext mirrors are stored beside numeric coordinates for operational geometry |

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
| New document version | Complete | Bates route creates a new PDF version only after visible overlay processing succeeds |
| Actual page-visible Bates numbering | Complete | service path requires reportlab+pypdf visible page labels; missing dependencies or overlay failure returns an error |
| Batch sequence allocation | Complete | `bates_sequences` allocates a global sequence when no explicit start is supplied |

## Audit

| Requirement | Status | Evidence |
|---|---|---|
| Audit helper writes logs | Complete | `domain_events.go` writes audit rows |
| Main mutation audit | Complete | mutation routes write audit and API tests cover the main acceptance flows |
| Filterable query | Complete | `auditLogsFiltered` supports actor/document/action/request/source/date filters and route is Admin-only |
| Indefinite retention | Complete | no deletion policy exists |

## Notifications

| Requirement | Status | Evidence |
|---|---|---|
| In-app queue | Complete | `notifications` table |
| Workflow notification | Complete | workflow transition calls `notifyUser` |
| 500 unread ceiling | Complete | `createNotification` marks oldest unread when ceiling is reached |
| Per-user query | Complete | `/api/notifications` uses principal user ID |
| Read acknowledgement | Complete | `/api/notifications/:id/read` verifies the row belongs to the caller and API coverage asserts missing notifications return 404 |
| Annotation mention notification | Complete | annotation creation calls `notifyMentionedUsers` before returning |
| Admin editable templates | Complete | templates can be listed and updated through Admin-only notification template endpoints |

## Backup and Recovery

| Requirement | Status | Evidence |
|---|---|---|
| Backup metadata | Complete | `backup_jobs` table and endpoint |
| Local backup artifact output | Complete | backup writes metadata plus strict `pg_dump` and filesystem `tar` artifacts; missing artifacts return an error |
| Real pg_dump execution | Complete | API success requires `pg_dump_custom` mode |
| Filesystem snapshot | Complete | API success requires `tar` snapshot mode |
| Restore workflow | Complete | Admin restore route requires artifact paths and returns success only after `pg_restore` and `tar` succeed |
| PITR docs | Complete | `docs/pitr.md` |

## Testing and Acceptance

| Requirement | Status | Evidence |
|---|---|---|
| Unit tests | Complete | root `run_tests.sh` directly invokes `go test ./...` |
| API tests | Complete | API tests cover auth/RBAC/upload/admin/workflow/redaction/Bates/backup/compare/notification denial flows and validate strict backup/restore response semantics |
| No SKIP-as-success suites | Complete | no SKIP-as-success acceptance suites are required for the documented path |
| Docker acceptance path | Complete | `scripts/docker_acceptance.sh` exists |
| Sample PDF/CSV | Complete | `testdata/` fixtures exist |

## Current Blocking Gaps

None tracked in this document after strict fallback semantics and compile fixes.
