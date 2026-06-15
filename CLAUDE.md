# CLAUDE Rules for IronPage Vault

## Mission

You are implementing IronPage Vault, an air-gapped legal PDF lifecycle management backend for law firms, paralegals, compliance teams, and regulated organizations. The project must be secure, offline, auditable, role-restricted, and deterministic.

## Required Stack

Use only the required stack unless the user explicitly changes the requirement:

- Backend: Go
- HTTP framework: Echo
- Database access: sqlx
- Database: PostgreSQL
- PDF binary storage: local filesystem
- Deployment: single Docker container
- Frontend: none

Do not add cloud services, external SaaS integrations, remote PDF APIs, external identity providers, or runtime network dependencies.

## Implementation Priorities

Prioritize correctness in this order:

1. Security and RBAC
2. Finalized document immutability
3. Audit completeness
4. PDF validation and local storage integrity
5. Redaction two-phase commit correctness
6. Version control and rollback correctness
7. Workflow transition correctness
8. Notifications and configuration
9. Backup and recovery
10. Developer ergonomics

## Architecture Rules

Use a layered structure:

- `cmd/server` for startup
- `internal/http` or `internal/handlers` for Echo handlers
- `internal/service` or domain packages for business rules
- `internal/repository` for sqlx persistence
- `internal/middleware` for auth, RBAC, request IDs, timestamps, and replay checks
- `internal/storage` for local PDF file handling
- `internal/pdf` for PDF inspection, redaction, Bates numbering, and comparison
- `internal/audit` for audit logging
- `internal/crypto` for encryption helpers
- `migrations` for schema
- `docs` for documentation

Business rules must live in services, not only in handlers.

## RBAC Contract

The only roles are:

- Admin
- Editor
- Reviewer

Admin capabilities:

- User management
- Configuration dictionaries
- Workflow status definitions
- Notification templates
- Backup configuration

Editor capabilities:

- Upload PDF
- Batch import PDFs up to 250 files
- Version rollback up to the 50-revision history ceiling
- Redaction confirmation
- Bates numbering
- Document finalization
- Document manipulation

Reviewer capabilities:

- Retrieve documents
- Add annotations
- Update annotation dispositions
- Advance review workflow where allowed

Never assume Admin automatically has Editor permissions unless this is explicitly coded, documented, and tested.

## Security Contract

Implement all of the following:

- bcrypt password verification
- 15-minute account lock after five failed attempts in a rolling window
- Locally signed JWT access tokens
- Server-side session tracking for 8-hour inactivity expiration
- JWT `jti` blacklist
- Request timestamp validation with 60-second expiry
- Replay protection using request ID or nonce tracking
- AES-256 column-level encryption for sensitive fields
- Contextual response masking by caller role
- Object-level authorization checks

Do not store plaintext passwords. Do not leak encrypted or unmasked PII in API responses.

## Document Contract

Enforce:

- PDF only
- Maximum 200 MB per document
- Maximum 500 pages per document
- Maximum 250 files per batch import
- Maximum 50 versions per document
- Local filesystem storage for binaries
- PostgreSQL pointers for file locations
- Immutable Finalized state

Finalized means permanently immutable. Reject all mutation attempts after Finalized.

## Workflow Contract

The required status chain is:

1. Draft
2. Under Review
3. Redaction Pending
4. Approved
5. Finalized

Each transition must:

- Validate role permission
- Validate allowed transition
- Reject Finalized mutation
- Write audit log
- Create notification where applicable

## Redaction Contract

Redaction is not complete until burn-in confirmation.

Phase 1:

- Store proposed redaction region metadata.
- Store page and coordinates.
- Encrypt sensitive coordinates at rest.

Phase 2:

- Authorized Editor confirms burn-in.
- Permanently strip the underlying content from the PDF binary.
- Create a new document version.
- Write audit log.

Do not treat a black rectangle overlay as final redaction.

## Annotation Contract

Reviewer annotations must support:

- Sticky notes
- Highlights
- Strikethroughs
- Freeform text stamps

Each annotation must include:

- Author
- Timestamp
- Page number
- Bounding coordinates
- Optional comment up to 2,000 characters
- Disposition

Valid dispositions:

- Approved
- Rejected
- Needs Discussion

## Bates Contract

Bates numbering must support:

- Prefix
- Suffix
- Zero-padding up to 10 digits
- Sequential numbering across batch sets
- Persistent job tracking
- Audit logging

## Audit Contract

Every mutating operation must write an audit log in the same transaction when possible.

Audit logs must record:

- Acting user
- Target document identifier where applicable
- Action type
- Timestamp
- Request ID
- Structured metadata

Audit logs are retained indefinitely.

## Notification Contract

Use PostgreSQL-backed in-app notifications only.

Create notifications for:

- Workflow transitions
- Annotation mentions

Enforce:

- Per-user notification queries
- Explicit read acknowledgment
- 500 unread notification ceiling
- Admin-editable templates

## API Response Contract

Pagination:

- Default: 25
- Hard max: 100

All errors must use the uniform envelope:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {},
    "request_id": "req_id",
    "timestamp": "2026-06-15T00:00:00Z"
  }
}
```

Never return inconsistent error shapes.

## Database Contract

Use PostgreSQL migrations. Do not modify schema manually without migrations.

Use transactions for:

- Upload plus version creation
- Rollback
- Redaction confirmation
- Workflow transition
- Bates numbering
- Annotation mutation
- User management
- Configuration changes

## Docker Contract

The final service must run as one container on a standalone machine with no external network dependency.

The container must include:

- PostgreSQL
- Go API binary
- Migrations
- Seed data
- Local storage directories
- Entrypoint startup script

Document volume mounts for:

- PostgreSQL data
- PDF storage
- Backups

## Testing Contract

Before considering the implementation complete, add tests for:

- Auth login success and failure
- Account lockout
- JWT blacklist
- Timestamp expiry
- Replay rejection
- Admin RBAC
- Editor RBAC
- Reviewer RBAC
- Object-level authorization
- PDF upload validation
- Batch import limit
- Version history ceiling
- Finalized immutability
- Redaction proposal and burn-in
- Bates numbering
- Annotation creation and disposition
- Workflow transition
- Audit log creation
- Notification creation
- Pagination bounds
- Backup job creation

## Documentation Contract

Keep documentation current with the code.

Required docs:

- README.md
- API.md
- RBAC.md
- SECURITY.md
- BACKUP_RECOVERY.md
- PITR.md
- TESTING.md
- DEPLOYMENT_OFFLINE.md

If implementation and documentation disagree, update the documentation or fix the implementation before finalizing.
