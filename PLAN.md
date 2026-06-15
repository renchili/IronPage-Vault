# IronPage Vault Implementation Plan

## 1. Project Goal

Build IronPage Vault as a fully offline, single-container legal PDF lifecycle management backend for air-gapped environments. The system must provide local authentication, strict RBAC, version-controlled PDF handling, forensic redaction workflows, Bates numbering, reviewer annotations, structured audit logging, in-app notifications, configurable workflow definitions, and local backup support.

## 2. Fixed Technology Stack

- Backend language: Go
- Backend framework: Echo
- Database access: sqlx
- Database: PostgreSQL
- Binary storage: local filesystem
- Frontend: none required
- Deployment: single Docker container on a standalone machine
- External dependencies at runtime: none

## 3. Core Domain Modules

### 3.1 Authentication and Session Management

- Local username and password login only.
- Store bcrypt-hashed credentials in PostgreSQL.
- Lock accounts for 15 minutes after 5 failed attempts within a rolling window.
- Issue locally signed JWT tokens.
- Enforce 8-hour inactivity timeout through server-side session tracking.
- Maintain a server-side JWT `jti` blacklist for logout and replay protection.
- Validate request timestamps and reject requests older than 60 seconds.

### 3.2 RBAC

Implement three roles only:

- Admin
- Editor
- Reviewer

RBAC must be enforced at the API capability boundary and again inside the service layer for sensitive operations.

### 3.3 Document Management

- Upload single PDF documents.
- Batch import up to 250 PDF files per operation.
- Store PDF binaries on the local filesystem.
- Store metadata and filesystem pointers in PostgreSQL.
- Reject files larger than 200 MB.
- Reject PDFs with more than 500 pages.
- Maintain a maximum of 50 revisions per document.
- Support version listing and rollback.
- Prevent all mutations once a document reaches Finalized status.

### 3.4 Workflow Management

Mandatory status chain:

1. Draft
2. Under Review
3. Redaction Pending
4. Approved
5. Finalized

Rules:

- No status skipping unless explicitly allowed by Admin-configured workflow rules.
- Finalized documents are permanently immutable.
- Workflow transition events must create audit records and notifications.

### 3.5 Redaction Workflow

- Redaction uses a two-phase commit model.
- Proposed redactions are stored as coordinate-bound metadata.
- Redaction coordinates must be encrypted at rest.
- Only an authorized Editor may confirm burn-in.
- Burn-in confirmation must permanently strip the underlying content from the PDF binary.
- Every redaction proposal and confirmation must be audited.

### 3.6 Annotation Workflow

Reviewers may:

- Retrieve documents.
- Create sticky notes.
- Create highlights.
- Create strikethroughs.
- Create freeform text stamps.
- Add optional comments up to 2,000 characters.
- Mark annotation disposition as Approved, Rejected, or Needs Discussion.
- Advance documents through the review workflow where permitted.

Each annotation must store:

- Author
- Timestamp
- Page number
- Bounding coordinates
- Annotation type
- Optional comment
- Disposition

### 3.7 Bates Numbering

- Support custom prefixes.
- Support custom suffixes.
- Support zero-padding up to 10 digits.
- Apply numbering sequentially across batch document sets.
- Record Bates jobs and sequence allocations.
- Audit all Bates operations.

### 3.8 Document Comparison

- Accept two version identifiers.
- Extract text locally.
- Return structured diff output containing:
  - Added text segments
  - Removed text segments
  - Modified text segments
  - Page numbers
  - Bounding coordinates where available

### 3.9 Audit Logging

Audit every mutating action, including:

- Document uploads
- Batch imports
- Version creation
- Version rollback
- Redaction proposals
- Redaction confirmations
- Workflow transitions
- Annotation creation
- Annotation disposition changes
- Bates numbering
- Document finalization
- User management
- Configuration changes
- Notification template changes
- Backup job actions

Audit records must include:

- Acting user
- Target document identifier where applicable
- Action type
- Timestamp
- Request ID
- Source IP where available
- Structured metadata

Audit logs must be retained indefinitely and support filtering by:

- User
- Document
- Action type
- Date range

### 3.10 Notification Queue

- Generate in-app notifications for workflow transitions.
- Generate in-app notifications for annotation mentions.
- Store notifications in PostgreSQL.
- Query notifications per user.
- Enforce a 500-record unread ceiling per user.
- Support explicit read acknowledgment.
- Notification templates must be Admin-configurable.

### 3.11 Configuration Center

Admins may manage:

- Workflow status definitions
- Notification templates
- System dictionaries
- Backup strategy configuration
- Naming conventions

Configuration changes must be audited.

### 3.12 Backup and Recovery

- Support scheduled PostgreSQL logical dumps.
- Support filesystem snapshots to a designated local volume.
- Store backup job metadata in PostgreSQL.
- Maintain point-in-time recovery documentation in the configuration center.
- No cloud or external network dependency is allowed.

## 4. Suggested Database Tables

- users
- roles
- user_roles
- login_attempts
- sessions
- jwt_blacklist
- request_replay_guard
- documents
- document_versions
- document_files
- document_status_history
- redaction_proposals
- redaction_confirmations
- annotations
- annotation_dispositions
- bates_jobs
- bates_sequences
- document_diffs
- audit_logs
- notifications
- notification_templates
- config_entries
- workflow_status_definitions
- backup_jobs

## 5. API Groups

### 5.1 Auth APIs

- POST /api/auth/login
- POST /api/auth/logout
- GET /api/auth/me

### 5.2 Admin APIs

- POST /api/admin/users
- GET /api/admin/users
- PATCH /api/admin/users/:id
- GET /api/admin/config
- PATCH /api/admin/config/:key
- GET /api/admin/workflow-statuses
- PATCH /api/admin/workflow-statuses
- GET /api/admin/notification-templates
- PATCH /api/admin/notification-templates/:id
- POST /api/admin/backup/run
- GET /api/admin/backup/jobs

### 5.3 Document APIs

- POST /api/documents
- POST /api/documents/batch
- GET /api/documents
- GET /api/documents/:id
- GET /api/documents/:id/file
- GET /api/documents/:id/versions
- POST /api/documents/:id/rollback
- POST /api/documents/:id/finalize
- POST /api/documents/:id/workflow/transition

### 5.4 Redaction APIs

- POST /api/documents/:id/redactions
- GET /api/documents/:id/redactions
- POST /api/documents/:id/redactions/:redaction_id/confirm

### 5.5 Annotation APIs

- POST /api/documents/:id/annotations
- GET /api/documents/:id/annotations
- PATCH /api/annotations/:id/disposition

### 5.6 Bates APIs

- POST /api/documents/:id/bates
- POST /api/documents/bates/batch

### 5.7 Comparison APIs

- POST /api/documents/compare

### 5.8 Audit and Notification APIs

- GET /api/audit-logs
- GET /api/notifications
- POST /api/notifications/:id/read

## 6. Global API Requirements

- All paginated responses default to 25 records.
- Maximum page size is 100 records.
- Every error response must use a uniform error envelope.
- Every mutating endpoint must generate an audit log.
- All APIs must reject unauthorized role access.
- All APIs must reject mutation attempts on Finalized documents.

## 7. Uniform Error Envelope

```json
{
  "error": {
    "code": "DOCUMENT_FINALIZED",
    "message": "Finalized documents are immutable",
    "details": {},
    "request_id": "req_example",
    "timestamp": "2026-06-15T00:00:00Z"
  }
}
```

## 8. Security Requirements

- No external authentication providers.
- No cloud service dependency.
- No remote PDF processing service.
- No network calls required after deployment.
- AES-256 column-level encryption for sensitive fields.
- Role-contextual response masking.
- bcrypt password hashing.
- JWT `jti` blacklist.
- Request timestamp validation.
- Server-side replay guard.
- Strict object-level authorization checks.

## 9. Docker Packaging Requirements

- The deliverable must run with a single container.
- PostgreSQL and the Go API must run inside that container.
- Runtime operation must not require internet access.
- All migrations and seed data must be included locally.
- Local storage directories must be mountable as Docker volumes.
- README must document startup, shutdown, backup, restore, and test procedures.

## 10. Acceptance Tests

The project is incomplete unless tests cover:

- Admin-only user management
- Admin-only configuration management
- Editor-only upload
- Editor-only batch import
- Reviewer-only annotation creation
- Reviewer cannot upload documents
- Editor cannot manage users
- Finalized documents reject all mutations
- Redaction proposal before confirmation
- Only Editor can confirm redaction burn-in
- 250-file batch import limit
- 50-version history ceiling
- 200 MB file-size enforcement
- 500-page PDF limit
- 2,000-character annotation comment limit
- Bates zero-padding up to 10 digits
- Pagination default and max page size
- Login failure lockout after five attempts
- JWT blacklist rejection
- Request timestamp rejection after 60 seconds
- Audit log creation for every mutating action
- Notification creation on workflow transition
- Unread notification ceiling of 500
- Backup job metadata creation

## 11. Implementation Milestones

### Milestone 1: Foundation

- Create Go module.
- Add Echo server.
- Add sqlx PostgreSQL connection.
- Add migrations.
- Add uniform response and error envelope.
- Add request ID middleware.
- Add health endpoint.

### Milestone 2: Auth and RBAC

- Implement users, roles, login attempts, sessions, JWT, blacklist, and replay guard.
- Add RBAC middleware.
- Add object-level authorization helpers.

### Milestone 3: Documents and Versions

- Implement PDF upload.
- Implement local file storage.
- Implement PDF validation.
- Implement document metadata.
- Implement version creation, listing, and rollback.

### Milestone 4: Workflow and Immutability

- Implement status transition engine.
- Enforce mandatory status chain.
- Enforce Finalized immutability.
- Add workflow audit logging and notifications.

### Milestone 5: Redaction and Bates

- Implement redaction proposal storage.
- Implement redaction burn-in confirmation.
- Implement Bates numbering jobs.
- Add audit records for all operations.

### Milestone 6: Annotations and Review

- Implement annotation creation.
- Implement disposition updates.
- Implement reviewer workflow advancement.

### Milestone 7: Audit, Notifications, and Config

- Implement audit filtering.
- Implement notification query and read acknowledgment.
- Implement Admin-managed configuration dictionaries, workflow definitions, and templates.

### Milestone 8: Backup and Recovery

- Implement backup job tracking.
- Implement local PostgreSQL logical dump invocation.
- Implement filesystem snapshot hooks.
- Add recovery documentation.

### Milestone 9: Docker and Final Validation

- Build single-container deployment.
- Add entrypoint process supervisor.
- Run migrations at startup.
- Add test runner.
- Validate fully offline execution.
