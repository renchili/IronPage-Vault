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

## 3. Backend Layering Plan

The backend must not keep growing as one large `internal/app` package. The target package split is:

```text
internal/app        HTTP API adapter: routes, middleware, binding, response mapping
internal/core       pure domain rules: roles, workflow, access policy, validation
internal/service    application use cases and transaction orchestration
internal/store      SQL repositories and persistence adapters
internal/platform   PDF, filesystem, digest, crypto, and backup adapters
```

### 3.1 Migration sequence

1. Move pure domain validation rules from `internal/app/rules.go` to `internal/core`.
2. Move workflow chain rules from `internal/app/workflows.go` to `internal/core`.
3. Move notification unread-cap policy from `internal/app/notifications.go` to `internal/core`.
4. Move text token and mention parsing policy from `internal/app/mentions.go` to `internal/core`.
5. Move object-level document access policy from `internal/app/access.go` to `internal/core`.
3. Move crypto and digest helpers from `internal/app` to `internal/platform`.
4. Keep temporary app wrapper functions only to avoid a large handler rewrite in the same PR.
5. Move handlers and services to call `internal/core` and `internal/platform` directly.
6. Delete app wrapper functions once no callers remain.
7. Move SQL query ownership to `internal/store`.
8. Move PDF helper implementation to `internal/platform`.
9. Move backup implementation to `internal/platform`.
10. Introduce service-level orchestration for upload, workflow, redaction, annotation, compare, Bates, and backup flows.

### 3.2 Current access-policy migration decision

Object access depends on principal role/user ID and document owner/status. It does not require Echo, SQL, or HTTP responses, so the real implementation belongs in `internal/core`.

The list-query SQL filter is different. A WHERE clause knows database columns and should not be put in `internal/core`. It remains an adapter concern and should later move to `internal/store`.

### 3.3 Current platform migration decision

Crypto, digest, and PDF helpers provide low-level infrastructure capabilities. They do not belong in `internal/app` because the API layer should not own AES-GCM encryption, SHA-256 hashing, PDF inspection, or filesystem PDF transforms.

The real implementations now belong in `internal/platform`. Temporary app wrappers preserve compatibility until callers are moved directly to platform or service-layer abstractions.

### 3.4 Current workflow migration decision

The workflow chain is a domain rule and belongs in `internal/core`. The API layer keeps a temporary wrapper until handlers or service-layer use cases call core directly.

### 3.5 Current notification migration decision

The unread-cap calculation is domain policy and belongs in `internal/core`. The SQL update and insert remain outside core because persistence behavior should later move to `internal/store`.

### 3.6 Current text token migration decision

Mention parsing is deterministic text policy and belongs in `internal/core`. User lookup and notification creation remain outside core because they are persistence side effects.

## 4. Core Domain Modules

### 4.1 Authentication and Session Management

- Local username and password login only.
- Store bcrypt-hashed credentials in PostgreSQL.
- Lock accounts for 15 minutes after 5 failed attempts within a rolling window.
- Issue locally signed JWT tokens.
- Enforce 8-hour inactivity timeout through server-side session tracking.
- Maintain a server-side JWT `jti` blacklist for logout and replay protection.
- Validate request timestamps and reject requests older than 60 seconds.

### 4.2 RBAC

Implement three roles only:

- Admin
- Editor
- Reviewer

RBAC must be enforced at the API capability boundary and again inside the service layer for sensitive operations.

### 4.3 Document Management

- Upload single PDF documents.
- Batch import up to 250 PDF files per operation.
- Store PDF binaries on the local filesystem.
- Store metadata and filesystem pointers in PostgreSQL.
- Reject files larger than 200 MB.
- Reject PDFs with more than 500 pages.
- Maintain a maximum of 50 revisions per document.
- Support version listing and rollback.
- Prevent all mutations once a document reaches Finalized status.

### 4.4 Workflow Management

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

### 4.5 Redaction Workflow

- Redaction uses a two-phase commit model.
- Proposed redactions are stored as coordinate-bound metadata.
- Redaction coordinates must be encrypted at rest.
- Only an authorized Editor may confirm burn-in.
- Burn-in confirmation must permanently strip the underlying content from the PDF binary.
- Every redaction proposal and confirmation must be audited.

### 4.6 Annotation Workflow

Reviewers may retrieve documents, create supported annotations, add comments up to 2,000 characters, set dispositions, and advance documents through permitted review workflow states.

Each annotation must store author, timestamp, page, coordinates, type, optional comment, and disposition.

### 4.7 Bates Numbering

- Support custom prefixes.
- Support custom suffixes.
- Support zero-padding up to 10 digits.
- Apply numbering sequentially across batch document sets.
- Record Bates jobs and sequence allocations.
- Audit all Bates operations.

### 4.8 Document Comparison

- Accept two version identifiers.
- Extract text locally.
- Return structured diff output containing added, removed, and modified segments with page numbers and bounding coordinates where available.

### 4.9 Audit Logging

- Record actor, document, action type, request ID, source IP, metadata, and timestamp.
- Support filtering by actor, document, action type, request ID, source IP, and time range.
