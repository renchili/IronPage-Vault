# Requirement Check

This document maps the prompt requirements to the current project implementation and documentation. It is written for acceptance review.

## Legend

| Status | Meaning |
|---|---|
| Complete | Implemented in code and documented |
| Partial | Implemented or documented, but needs more depth |
| Planned | Documented as required, not yet complete |
| Manual Patch | Requires a user-side edit because automated writes were blocked |
| Not applicable | Outside current requested scope |

## Stack and Deployment

| Requirement | Status | Evidence |
|---|---|---|
| Pure backend project | Complete | README and design docs state UI is only a backend test aid |
| Backend in Go | Complete | `go.mod`, `cmd/server/main.go`, `internal/app/` |
| Echo framework | Complete | Echo server and routes in `internal/app/server.go` |
| sqlx database access | Complete | `internal/app/database.go` and handlers use sqlx |
| PostgreSQL as metadata store | Complete | `migrations/001_schema.sql`, Docker image based on PostgreSQL |
| PDF binaries on local filesystem | Complete | upload handler stores files under `STORAGE_DIR` |
| No external runtime services | Complete | Docker Compose uses one service only |
| Single-container standalone deployment | Complete | `Dockerfile`, `docker-compose.yml`, `scripts/entrypoint.sh` |
| Docker builder instead of local build | Complete | Dockerfile uses a Go builder stage and no local build artifacts are required |
| Swaggo/OpenAPI support | Partial | Swaggo dependencies, `docs/swagger/swagger.yaml`, and `public/swagger.yaml` exist; generated Swaggo files are not produced because no commands were run |
| Backend test UI | Complete | `public/manual-test.html` is an API test tool, not a frontend deliverable |

## Authentication and Session Security

| Requirement | Status | Evidence |
|---|---|---|
| Local username/password auth | Complete | `POST /api/auth/login` |
| bcrypt password hashing | Complete | seeded users and user creation use bcrypt |
| Account lock after 5 failed attempts | Complete | login handler updates failed attempts and lock window |
| JWT tokens | Complete | login issues locally signed JWT |
| 8-hour inactivity timeout | Complete | server-side sessions with `last_seen_at` |
| JWT jti blacklist | Complete | logout inserts into `jwt_blacklist` |
| Request timestamp validation | Complete | auth middleware checks `X-Request-Timestamp` |
| Replay guard | Complete | auth middleware records `X-Request-ID` per jti |
| AES-256 column-level encryption | Planned | schema and docs reserve sensitive-field handling; full encryption helper integration remains future work |
| Contextual masking | Partial | password hash is hidden from JSON model; full role-contextual masking needs more endpoints |

## RBAC

| Requirement | Status | Evidence |
|---|---|---|
| Admin role | Complete | Admin routes under `/api/admin` |
| Editor role | Complete | document upload, redaction, Bates, finalize routes |
| Reviewer role | Complete | annotation routes |
| Strict API boundary | Complete | `requireRole` middleware |
| Admin not automatically Editor | Complete | Admin is not included in Editor-only routes |
| Object-level authorization | Partial | document existence and finalized checks exist; ownership/team scoping can be expanded |

## Document Lifecycle

| Requirement | Status | Evidence |
|---|---|---|
| PDF upload | Complete | `POST /api/documents` |
| Local PDF inspection | Complete | `%PDF-` header and page marker inspection |
| 200 MB size limit | Complete | upload and inspection enforce configured max bytes |
| 500 page limit | Complete | local page count check |
| Metadata in PostgreSQL | Complete | `documents`, `document_versions` |
| Local filesystem pointer | Complete | `document_versions.file_path` |
| Batch import up to 250 | Partial | endpoint validates the limit; full persistence still needs expansion |
| 50 revision ceiling | Partial | redaction confirmation checks max versions; all version paths need full test coverage |
| Version rollback | Complete | rollback validates version, rejects Finalized, updates current version, and writes audit log |
| Document comparison | Partial | endpoint returns structured shape; deeper PDF text extraction can be improved |

## Workflow

| Requirement | Status | Evidence |
|---|---|---|
| Draft status | Complete | schema and constants |
| Under Review status | Complete | schema and transition chain |
| Redaction Pending status | Complete | schema and transition chain |
| Approved status | Complete | schema and transition chain |
| Finalized status | Complete | schema and transition chain |
| Mandatory status chain | Complete | `nextWorkflowStatus` |
| Finalized immutable | Complete | enforced in redaction, annotation, Bates, rollback, transition, and finalize paths |
| Workflow audit record | Complete | transition handler writes audit log |
| Workflow notification | Complete | transition handler creates notification |

## Redaction

| Requirement | Status | Evidence |
|---|---|---|
| Staged coordinate metadata | Complete | `redaction_proposals` table and endpoint |
| Editor confirmation | Complete | confirm route is Editor-only |
| New version after confirmation | Complete | confirm creates a document version |
| Audit for proposal | Complete | proposal endpoint writes audit log |
| Audit for confirmation | Complete | confirmation endpoint writes audit log |
| Permanent PDF content removal | Partial | prototype creates transformed version; production-grade content stripping engine can replace transform function |
| Coordinate encryption | Planned | schema stores coordinates; AES integration remains next step |
| Missing document handling | Complete | mutable-state helper now separates not-found from Finalized |

## Annotation

| Requirement | Status | Evidence |
|---|---|---|
| Sticky note | Complete | accepted annotation type |
| Highlight | Complete | accepted annotation type |
| Strikethrough | Complete | accepted annotation type |
| Freeform text stamp | Complete | accepted annotation type |
| Author attribution | Complete | `author_user_id` |
| Timestamp | Complete | `created_at` |
| Coordinates | Complete | x/y/width/height fields |
| Comment cap 2000 | Complete | handler validates comment length |
| Disposition update | Complete | `PATCH /api/annotations/:id/disposition` |
| Reviewer-only creation | Complete | route role restriction |

## Bates Numbering

| Requirement | Status | Evidence |
|---|---|---|
| Prefix | Complete | `bates_jobs.prefix` |
| Suffix | Complete | `bates_jobs.suffix` |
| Zero padding up to 10 | Complete | handler validates 0 to 10 |
| Start number | Complete | request and schema support start number |
| Persistent job record | Complete | `bates_jobs` table |
| Audit | Complete | Bates endpoint writes audit log |
| Sequential batch across sets | Planned | current job stores sequence settings; batch sequence allocation table can be added next |

## Audit

| Requirement | Status | Evidence |
|---|---|---|
| Mutating action audit | Partial | implemented for main flows; must be kept mandatory for future flows |
| Actor | Complete | `actor_user_id` |
| Document ID | Complete | nullable `document_id` |
| Action type | Complete | `action_type` |
| Timestamp | Complete | `created_at` |
| Request ID | Complete | `request_id` |
| Filterable query | Partial | list endpoint exists; filters by user/document/action/date should be added |
| Indefinite retention | Complete | no deletion policy |

## Notifications

| Requirement | Status | Evidence |
|---|---|---|
| In-app queue | Complete | `notifications` table |
| Workflow notification | Complete | transition handler creates row |
| Annotation mention notification | Planned | annotation data model supports future mentions; parser is not complete |
| Per-user query | Complete | `/api/notifications` uses principal user ID |
| Read acknowledgment | Complete | `/api/notifications/:id/read` |
| 500 unread ceiling | Planned | table exists; enforcement can be added before insert paths |
| Admin-editable templates | Partial | templates exist; edit endpoint can be expanded |

## Configuration and Backup

| Requirement | Status | Evidence |
|---|---|---|
| Config entries | Complete | `config_entries` table and admin endpoints |
| Workflow status definitions | Complete | `workflow_status_definitions` table and endpoint |
| Notification templates | Complete | `notification_templates` table and endpoint |
| Backup job metadata | Complete | `backup_jobs` table and admin endpoint |
| Logical dump execution | Planned | metadata endpoint exists; actual local dump command can be added later |
| Filesystem snapshot | Planned | backup volume documented; snapshot implementation can be added next |
| PITR docs | Complete | `docs/pitr.md` |
| Backup/recovery docs | Complete | `docs/backup-recovery.md` |

## Testing and Acceptance

| Requirement | Status | Evidence |
|---|---|---|
| README project overview | Complete | `README.md` focuses on project and implementation |
| Usage document | Complete | `docs/usage.md` contains commands and operation flow |
| `unit_tests/` | Complete | `unit_tests/test_rules.sh` |
| `API_tests/` | Partial | helper, RBAC, upload scripts exist; full E2E login bootstrap is a manual patch |
| `run_tests.sh` | Complete | root script calls unit and API suites |
| Real sample PDF | Complete | `testdata/pdfs/sample_contract.pdf` |
| CSV sample data | Complete | `testdata/csv/batch_import_manifest.csv` |
| Multi-role coverage | Partial | implemented in scripts where tokens are supplied; hardcoded seeded token bootstrap left for user patch |
| Test UI | Complete | `public/manual-test.html` directly calls backend APIs |

## Overall Completion Summary

The project now has a pure-backend API structure, database schema, Docker builder setup, core handlers, seed data, sample PDF, CSV manifest, test UI, README, and design/API/requirement documentation.

Remaining high-value work before strict final acceptance:

1. Apply the manual seeded login bootstrap patch in `API_tests/test_api_flow.sh`.
2. Expand API tests for annotation, redaction, Bates, workflow, finalization, audit, notification, and backup assertions.
3. Replace prototype PDF transform with a stronger local PDF redaction implementation.
4. Add AES-256 encryption integration for sensitive columns.
5. Expand object-level authorization and audit filters.
6. Expand batch import from limit validation to full persisted batch import.
7. Generate official Swaggo files with Docker builder during acceptance if desired.
