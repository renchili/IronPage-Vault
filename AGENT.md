# AGENT Rules for IronPage Vault

## 1. Non-Negotiable Requirements

The agent must implement IronPage Vault exactly as an air-gapped legal PDF lifecycle management backend. The system must use:

- Go as the backend language.
- Echo as the REST API framework.
- sqlx for database access.
- PostgreSQL as the only metadata, audit, configuration, session, and notification database.
- Local filesystem storage for PDF binary assets.
- A single Docker container for standalone offline deployment.

The agent must not introduce:

- External authentication providers.
- Cloud services.
- Remote PDF processing APIs.
- External notification delivery services.
- Non-PostgreSQL metadata stores.
- A required frontend.
- Runtime internet dependency.

## 2. Security Rules

The agent must enforce security in middleware and service layers.

Required controls:

- bcrypt password hashing.
- Account lock for 15 minutes after 5 failed attempts in a rolling window.
- Locally signed JWT tokens.
- 8-hour inactivity timeout backed by server-side session state.
- JWT `jti` blacklist.
- Request timestamp validation rejecting requests older than 60 seconds.
- Replay guard for request identifiers or nonces.
- AES-256 column-level encryption for sensitive fields.
- Role-contextual response masking.
- Object-level authorization checks.

The agent must never rely only on frontend checks, handler checks, or route grouping for authorization.

## 3. RBAC Rules

The only supported roles are:

- Admin
- Editor
- Reviewer

Admin may:

- Manage user accounts.
- Manage system configuration dictionaries.
- Manage workflow status definitions.
- Manage notification templates.
- Manage backup configuration.

Editor may:

- Upload documents.
- Batch import up to 250 files.
- Roll back versions within the 50-revision ceiling.
- Confirm redaction burn-in.
- Apply Bates numbering.
- Finalize documents.
- Manipulate document versions.

Reviewer may:

- Retrieve documents.
- Add annotations.
- Set annotation dispositions.
- Advance documents through the review workflow where permitted.

The agent must not accidentally grant Admin full document editing authority unless the project explicitly defines that behavior in code, tests, and documentation.

## 4. Document Rules

The agent must enforce:

- Maximum 200 MB per PDF.
- Maximum 500 pages per PDF.
- Maximum 250 files per batch import.
- Maximum 50 revisions per document.
- Local file inspection at ingestion time.
- Filesystem pointers stored in PostgreSQL.
- Immutable Finalized documents.

Once a document is Finalized, the agent must reject:

- Upload replacement.
- Version rollback.
- Redaction proposal.
- Redaction confirmation.
- Annotation creation.
- Annotation disposition change.
- Bates numbering.
- Workflow transition.
- Any metadata mutation.

## 5. Workflow Rules

The mandatory document status chain is:

1. Draft
2. Under Review
3. Redaction Pending
4. Approved
5. Finalized

The agent must enforce this chain unless Admin-configurable workflow definitions explicitly allow a different transition.

Every workflow transition must:

- Validate caller permission.
- Validate current status.
- Reject mutation if Finalized.
- Write an audit log.
- Create notifications where applicable.

## 6. Redaction Rules

Redaction must use two-phase commit:

1. Proposed redaction regions are staged as encrypted coordinate-bound metadata.
2. An authorized Editor confirms burn-in.

The agent must ensure that burn-in permanently strips the underlying PDF content region from the binary. Visual overlays alone are not acceptable as completed redaction.

Every redaction proposal and confirmation must produce an audit log.

## 7. Annotation Rules

Annotations must include:

- Author attribution.
- Timestamp.
- Page number.
- Bounding coordinates.
- Annotation type.
- Optional comment capped at 2,000 characters.
- Disposition.

Allowed annotation types:

- Sticky note
- Highlight
- Strikethrough
- Freeform text stamp

Allowed dispositions:

- Approved
- Rejected
- Needs Discussion

Reviewer permissions must be strictly enforced.

## 8. Bates Numbering Rules

The agent must support:

- Custom prefixes.
- Custom suffixes.
- Zero-padding up to 10 digits.
- Sequential numbering across batch document sets.
- Persistent Bates job records.
- Auditable sequence allocation.

## 9. Audit Rules

Every mutating action must create an audit log record. Audit logs must be retained indefinitely.

Audit logs must support filtering by:

- User
- Document
- Action type
- Date range

Audit logs must include:

- Acting user
- Target document identifier where applicable
- Action type
- Timestamp
- Request ID
- Structured metadata

## 10. Notification Rules

The agent must implement an internal in-app notification queue.

Notifications must be generated for:

- Workflow status transitions.
- Annotation mentions.

The agent must enforce:

- Per-user queries.
- Explicit read acknowledgment.
- 500 unread notification ceiling per user.
- Admin-editable notification templates.

## 11. API Rules

All API responses must be consistent.

Pagination:

- Default page size: 25
- Maximum page size: 100

All error responses must use this shape:

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

The agent must not return ad hoc error strings.

## 12. Database Rules

The agent must use migrations for all schema changes.

The database must include tables for:

- Users and roles
- Login attempts
- Sessions
- JWT blacklist
- Replay guard
- Documents
- Document versions
- Document files
- Workflow history
- Redaction proposals
- Redaction confirmations
- Annotations
- Bates jobs
- Bates sequences
- Document diffs
- Audit logs
- Notifications
- Notification templates
- Configuration entries
- Workflow status definitions
- Backup jobs

The agent must use transactions for multi-step mutations.

## 13. Backup Rules

The agent must support:

- Scheduled PostgreSQL logical dumps.
- Filesystem snapshots to a local volume.
- Backup job metadata.
- Point-in-time recovery documentation.

No backup process may require cloud storage or external network access.

## 14. Testing Rules

The agent must provide tests for:

- RBAC denial and approval paths.
- Finalized immutability.
- Redaction two-phase commit.
- Version limit enforcement.
- Batch import limit enforcement.
- PDF size and page validation.
- Login lockout.
- JWT blacklist.
- Request timestamp expiry.
- Audit logging.
- Notification creation and read acknowledgment.
- Pagination limits.
- Admin-only configuration.

A project without these tests is incomplete.

## 15. Documentation Rules

The agent must maintain:

- README.md
- API documentation
- RBAC documentation
- Security documentation
- Backup and restore documentation
- Point-in-time recovery documentation
- Test instructions
- Offline deployment instructions

Documentation must match the actual implementation.
