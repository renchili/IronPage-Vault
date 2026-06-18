# API Guide

IronPage Vault is a Go/Echo backend using PostgreSQL for metadata and local filesystem storage for PDF binaries.

## API contract source of truth

Route-level Swaggo annotations in Go source are the authoritative API contract. Generate the OpenAPI artifact with:

```bash
bash scripts/generate_swagger.sh
```

This produces `docs/swagger/swagger.yaml` and `docs/swagger/swagger.json`. Swagger UI is served at:

```text
GET /swagger/index.html
```

Do not add or maintain parallel handwritten OpenAPI files. This Markdown document is an operational guide; generated Swagger is the machine-readable contract.

## Common requirements

Authenticated endpoints require:

```text
Authorization: Bearer <token>
X-Request-ID: unique request id
X-Request-Timestamp: RFC3339 timestamp within 60 seconds
```

Business errors use this envelope:

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

Collection endpoints use `page` and `page_size`; the intended default is `25` and the maximum is `100`.

## Roles

| Role | Responsibilities |
|---|---|
| Admin | user management, configuration, workflow definitions, notification templates, audit log access, backup and restore |
| Editor | document upload, batch import, rollback, redaction confirmation, Bates numbering, finalization |
| Reviewer | document review, annotation creation and disposition, permitted workflow movement |

Object-level document access is evaluated in addition to route-level role checks.

## Route inventory

### Health and identity

| Method | Route | Access |
|---|---|---|
| GET | `/healthz` | public |
| POST | `/api/auth/login` | public |
| POST | `/api/auth/logout` | authenticated |
| GET | `/api/auth/me` | authenticated |

### Administration

| Method | Route | Access |
|---|---|---|
| POST | `/api/admin/users` | Admin |
| GET | `/api/admin/users` | Admin |
| GET | `/api/admin/config` | Admin |
| PATCH | `/api/admin/config/:key` | Admin |
| GET | `/api/admin/workflow-statuses` | Admin |
| GET | `/api/admin/notification-templates` | Admin |
| PATCH | `/api/admin/notification-templates/:key` | Admin |
| POST | `/api/admin/backup/run` | Admin |
| GET | `/api/admin/backup/jobs` | Admin |
| POST | `/api/admin/backup/restore` | Admin |

Backup restore rejects an empty request with `400`; restoring a valid returned artifact pair completes with `200` and a `Restored` status.

### Documents and versions

| Method | Route | Access |
|---|---|---|
| GET | `/api/documents` | authenticated + object policy |
| POST | `/api/documents` | Editor |
| POST | `/api/documents/batch` | Editor |
| POST | `/api/documents/compare` | authenticated + version access |
| GET | `/api/documents/:id` | authenticated + object policy |
| GET | `/api/documents/:id/file` | authenticated + object policy |
| GET | `/api/documents/:id/versions` | authenticated + object policy |
| POST | `/api/documents/:id/rollback` | Editor + object policy |
| POST | `/api/documents/:id/finalize` | Editor + object policy |
| POST | `/api/documents/:id/workflow/transition` | Editor or Reviewer + policy |

Workflow status chain:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Finalized documents are immutable.

### Redactions, annotations, and Bates

| Method | Route | Access |
|---|---|---|
| POST | `/api/documents/:id/redactions` | Editor + object policy |
| GET | `/api/documents/:id/redactions` | authenticated + object policy |
| POST | `/api/documents/:id/redactions/:redaction_id/confirm` | Editor + object policy |
| POST | `/api/documents/:id/annotations` | Reviewer + object policy |
| GET | `/api/documents/:id/annotations` | authenticated + object policy |
| PATCH | `/api/annotations/:id/disposition` | Reviewer + object policy |
| POST | `/api/documents/:id/bates` | Editor + object policy |

Redaction confirmation creates a new version using strict burn-in. Bates creation uses prefix, suffix, padding, and start number inputs; the resulting job is audited.

### Audit and notifications

| Method | Route | Access |
|---|---|---|
| GET | `/api/audit-logs` | Admin |
| GET | `/api/notifications` | authenticated |
| POST | `/api/notifications/:id/read` | authenticated |

Audit logs support filtering by user, document, action type, and time range where supplied by the endpoint.

## Local verification

```bash
./run_tests.sh
bash scripts/generate_swagger.sh
```

For exact request and response schemas, use generated Swagger UI or `docs/swagger/swagger.yaml` after generation.
