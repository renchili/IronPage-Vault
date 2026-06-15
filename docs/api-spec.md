# API Specification

This document describes the IronPage Vault REST API. The implementation is a Go Echo backend using PostgreSQL for metadata and local filesystem storage for PDF binaries.

## OpenAPI and Swaggo Support

The project includes Swaggo dependencies in `go.mod`:

- `github.com/swaggo/swag`
- `github.com/swaggo/echo-swagger`

Recommended generation command:

```bash
swag init -g cmd/server/main.go -o docs/swagger
```

Recommended Swagger UI route when enabled in the Echo server:

```text
GET /swagger/*
```

The Markdown API spec below is the acceptance reference. Swagger annotations should mirror this file.

## Common Requirements

Authenticated endpoints require:

```text
Authorization: Bearer <token>
X-Request-ID: unique request id
X-Request-Timestamp: RFC3339 timestamp within 60 seconds
```

All business errors use this envelope:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {},
    "request_id": "req_example",
    "timestamp": "2026-06-15T00:00:00Z"
  }
}
```

Pagination defaults to `page_size=25` and has a hard maximum of `100`.

## Seed Roles

| Role | Purpose |
|---|---|
| Admin | user management, config, workflow definitions, notification templates, backup metadata |
| Editor | document upload, batch import, rollback, redaction confirmation, Bates numbering, finalization |
| Reviewer | document retrieval, annotation creation, annotation disposition, review workflow movement |

## Auth APIs

### POST /api/auth/login

Local username and password login.

Request:

```json
{
  "username": "editor",
  "password": "Editor123!"
}
```

Success response:

```json
{
  "token": "jwt-token",
  "token_type": "Bearer",
  "expires_in_seconds": 28800,
  "user": {
    "id": "usr_xxx",
    "username": "editor",
    "display_name": "Default Editor",
    "role": "Editor"
  }
}
```

Important errors:

- `INVALID_LOGIN_REQUEST`
- `INVALID_CREDENTIALS`
- `ACCOUNT_LOCKED`
- `TOKEN_SIGN_ERROR`
- `SESSION_CREATE_ERROR`

### POST /api/auth/logout

Revokes the current JWT by storing its `jti` in the server-side blacklist.

Required role: any authenticated user.

### GET /api/auth/me

Returns the authenticated principal.

Required role: any authenticated user.

## Admin APIs

### POST /api/admin/users

Required role: Admin.

Creates a local user.

Request:

```json
{
  "username": "legal-editor-2",
  "display_name": "Legal Editor 2",
  "role": "Editor",
  "password": "Editor123!"
}
```

### GET /api/admin/users

Required role: Admin.

Lists local users with masked password hashes.

### GET /api/admin/config

Required role: Admin.

Lists configuration entries.

### PATCH /api/admin/config/:key

Required role: Admin.

Updates one configuration entry.

Request:

```json
{
  "value": "new-value"
}
```

### GET /api/admin/workflow-statuses

Required role: Admin.

Lists workflow status definitions.

### GET /api/admin/notification-templates

Required role: Admin.

Lists notification templates.

### POST /api/admin/backup/run

Required role: Admin.

Creates backup job metadata for a local logical dump or filesystem snapshot process.

### GET /api/admin/backup/jobs

Required role: Admin.

Lists backup jobs.

## Document APIs

### POST /api/documents

Required role: Editor.

Uploads a single PDF document as multipart form data.

Form fields:

| Field | Required | Description |
|---|---:|---|
| title | no | document title |
| file | yes | PDF file |

Validation:

- file must start with a valid PDF header
- file size must not exceed 200 MB
- page count must not exceed 500 pages

### POST /api/documents/batch

Required role: Editor.

Accepts up to 250 PDF files in one operation.

### GET /api/documents

Required role: authenticated user.

Lists documents with pagination.

### GET /api/documents/:id

Required role: authenticated user.

Returns document metadata.

### GET /api/documents/:id/file

Required role: authenticated user.

Downloads the current PDF version.

### GET /api/documents/:id/versions

Required role: authenticated user.

Lists versions for a document.

### POST /api/documents/:id/rollback

Required role: Editor.

Rolls back to a previous version within the 50-version ceiling.

### POST /api/documents/:id/workflow/transition

Required role: Editor or Reviewer, depending on transition.

Request:

```json
{
  "status": "Under Review"
}
```

Status chain:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

### POST /api/documents/:id/finalize

Required role: Editor.

Marks a document as Finalized. Finalized documents are immutable.

## Redaction APIs

### POST /api/documents/:id/redactions

Required role: Editor.

Stages a redaction region.

Request:

```json
{
  "page": 1,
  "x": 10,
  "y": 20,
  "width": 100,
  "height": 40,
  "reason": "privileged content"
}
```

### GET /api/documents/:id/redactions

Required role: authenticated user.

Lists redaction proposals for the document.

### POST /api/documents/:id/redactions/:redaction_id/confirm

Required role: Editor.

Confirms redaction burn-in and creates a new document version.

## Annotation APIs

### POST /api/documents/:id/annotations

Required role: Reviewer.

Request:

```json
{
  "type": "Sticky note",
  "page": 1,
  "x": 10,
  "y": 20,
  "width": 100,
  "height": 30,
  "comment": "Needs legal review",
  "disposition": "Needs Discussion"
}
```

Allowed annotation types:

- Sticky note
- Highlight
- Strikethrough
- Freeform text stamp

Allowed dispositions:

- Approved
- Rejected
- Needs Discussion

### GET /api/documents/:id/annotations

Required role: authenticated user.

Lists annotations.

### PATCH /api/annotations/:id/disposition

Required role: Reviewer.

Request:

```json
{
  "disposition": "Approved"
}
```

## Bates APIs

### POST /api/documents/:id/bates

Required role: Editor.

Request:

```json
{
  "prefix": "CASE-",
  "suffix": "-A",
  "zero_padding": 6,
  "start": 1
}
```

Rules:

- zero padding must be between 0 and 10
- Bates job creation must be audited

## Comparison API

### POST /api/documents/compare

Required role: authenticated user.

Request:

```json
{
  "left_version_id": "ver_left",
  "right_version_id": "ver_right"
}
```

Response includes added, removed, and modified segments with page and bounding box placeholders where exact PDF coordinates are not available.

## Audit APIs

### GET /api/audit-logs

Required role: authenticated user.

Supports pagination and is intended to support filters by user, document, action type, and date range.

## Notification APIs

### GET /api/notifications

Required role: authenticated user.

Returns the current user's in-app notifications.

### POST /api/notifications/:id/read

Required role: authenticated user.

Marks a notification as read.
