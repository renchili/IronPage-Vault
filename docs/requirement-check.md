# Requirement Check

This document maps the prompt requirements to the current project implementation and documentation. It is written for acceptance review.

## Legend

| Status | Meaning |
|---|---|
| Complete | Implemented in code and documented |
| Partial | Implemented or documented, but needs more depth |
| Planned | Documented as required, not yet complete |
| Not applicable | Outside current requested scope |

## Stack and Deployment

| Requirement | Status | Evidence |
|---|---|---|
| Backend in Go | Complete | `go.mod`, `cmd/server/main.go`, `internal/app/` |
| Echo framework | Complete | Echo server and routes in `internal/app/server.go` |
| sqlx database access | Complete | `internal/app/database.go` and handlers use sqlx |
| PostgreSQL as metadata store | Complete | `migrations/001_schema.sql`, Docker image based on PostgreSQL |
| PDF binaries on local filesystem | Complete | upload handler stores files under `STORAGE_DIR` |
| No external runtime services | Complete | Docker Compose uses one service only |
| Single-container standalone deployment | Complete | `Dockerfile`, `docker-compose.yml`, `scripts/entrypoint.sh` |
| Swaggo support | Partial | Swaggo dependencies added and generation command documented; route wiring and generated docs can be added next |
| Informal test frontend | Partial | `public/manual-test.html` added as guidance page; interactive console can be expanded |

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
| Batch import up to 250 | Partial | endpoint validates the limit; persistence flow can be expanded from single-file path |
| 50 revision ceiling | Partial | redaction confirmation checks max versions; rollback and other version creation paths need full coverage |
| Version rollback | Planned | route exists; complete service logic still needs final implementation |
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
| Finalized immutable | Partial | enforced in redaction, annotation, transition, finalize; extend to every future mutation |
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
| Logical dump execution | Planned | metadata endpoint exists; actual dump command should be added carefully for local-only use |
| Filesystem snapshot | Planned | backup volume documented; snapshot implementation can be added next |
| PITR docs | Planned | required docs need expansion |

## Testing and Acceptance

| Requirement | Status | Evidence |
|---|---|---|
| README startup docs | Complete | `README.md` |
| `unit_tests/` | Complete | `unit_tests/test_rules.sh` |
| `API_tests/` | Partial | directory planned; script creation may need retry if connector blocks content |
| `run_tests.sh` | Planned | should call unit and API scripts |
| Real sample PDF | Complete | `testdata/pdfs/sample_contract.pdf` |
| CSV sample data | Complete | `testdata/csv/batch_import_manifest.csv` |
| Multi-role coverage | Partial | README matrix and API plan; API script should be expanded |

## Overall Completion Summary

The project now has the main backend skeleton, database schema, Docker startup files, core handlers, seed data, sample PDF, CSV manifest, README, and design/API/requirement documentation.

Remaining high-value work before strict final acceptance:

1. Complete `run_tests.sh` and API test scripts.
2. Wire Swagger UI route and generate `docs/swagger` output.
3. Replace prototype PDF transform with a stronger local PDF redaction implementation.
4. Add full AES-256 encryption integration for sensitive columns.
5. Expand object-level authorization and audit filters.
6. Complete rollback service logic and batch import persistence.
7. Expand informal test frontend into an interactive console if desired.
