# API Guide

IronPage Vault is a Go/Echo backend using PostgreSQL for metadata and local filesystem storage for PDF binaries.

## Contract source of truth

Route-level Swaggo annotations are authoritative. Supported execution entrypoints generate `docs/swagger/swagger.yaml` and `docs/swagger/swagger.json` with:

```bash
bash scripts/generate_swagger.sh
```

Swagger UI is served at `GET /swagger/index.html`. This Markdown file is an operational guide, not a parallel OpenAPI definition. Static review does not authorize generation.

## Common requirements

Authenticated endpoints require:

```text
Authorization: Bearer <token>
X-Request-ID: unique request id
X-Request-Timestamp: RFC3339 timestamp within 60 seconds
```

Errors use:

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

Collection endpoints use `page` and `page_size`; default is `25` and maximum is `100`.

## Roles

| Role | Responsibilities |
|---|---|
| Admin | users, configuration, persisted workflow definitions, templates, audit logs, backup and restore |
| Editor | owned-document upload, batch import, rollback, redaction confirmation, Bates, finalization |
| Reviewer | readable-document review, annotations/dispositions, permitted workflow transitions |

Object policy is evaluated in addition to route role checks. Admin is not implicitly an Editor.

## Route inventory

### Health and identity

| Method | Route | Access |
|---|---|---|
| GET | `/healthz` | public |
| POST | `/api/auth/login` | public |
| POST | `/api/auth/logout` | authenticated |
| GET | `/api/auth/me` | authenticated |

Failed-login state/lock audit, successful reset/session/audit, and logout blacklist/session/audit are transactional. Authentication persistence errors fail closed.

### Administration

| Method | Route | Access |
|---|---|---|
| POST | `/api/admin/users` | Admin |
| GET | `/api/admin/users` | Admin |
| GET | `/api/admin/config` | Admin |
| PATCH | `/api/admin/config/:key` | Admin |
| GET | `/api/admin/workflow-statuses` | Admin |
| PUT | `/api/admin/workflow-statuses` | Admin |
| GET | `/api/admin/notification-templates` | Admin |
| PATCH | `/api/admin/notification-templates/:key` | Admin |
| POST | `/api/admin/backup/run` | Admin |
| GET | `/api/admin/backup/jobs` | Admin |
| POST | `/api/admin/backup/restore` | Admin |

Workflow replacement body:

```json
{
  "statuses": [
    {"name": "Draft", "mutable": true},
    {"name": "Under Review", "mutable": true},
    {"name": "Redaction Pending", "mutable": true},
    {"name": "Approved", "mutable": true},
    {"name": "Finalized", "mutable": false}
  ]
}
```

The array is the complete ordered chain. `Draft` must be first/mutable, `Finalized` last/immutable, names are case-insensitively unique, intermediate states are mutable, and a status used by an existing document cannot be removed. Replacement and its audit commit together.

Backup returns both strict artifact paths. Restore requires both paths, records a restore ID and Requested state, then records Completed or Failed. `200` is returned only after Completed state and audit are persisted.

### Documents and versions

| Method | Route | Access |
|---|---|---|
| GET | `/api/documents` | authenticated + object policy |
| POST | `/api/documents` | Editor + object policy |
| POST | `/api/documents/batch` | Editor + object policy |
| POST | `/api/documents/compare` | authenticated + version access |
| GET | `/api/documents/:id` | authenticated + object policy |
| GET | `/api/documents/:id/file` | authenticated + object policy |
| GET | `/api/documents/:id/versions` | authenticated + object policy |
| POST | `/api/documents/:id/rollback` | Editor + object policy |
| POST | `/api/documents/:id/finalize` | Editor + object policy |
| POST | `/api/documents/:id/workflow/transition` | Editor or Reviewer + policy |

The initial chain is `Draft -> Under Review -> Redaction Pending -> Approved -> Finalized`. Runtime transition validation reads the persisted ordered definitions. A successful transition/finalization includes document state, status history, audit and owner notification in one transaction. Finalized is terminal.

Comparison returns structured text changes with page and bounding-box data.

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

Redaction confirmation creates and verifies a strict new PDF version. Failed persistence removes the generated file. Annotation creation commits encrypted comment, audit, and mention notifications together; list responses decrypt the comment.

Bates reserves a global range equal to the source page count. The response includes `start_number` and `end_number`. Range reservation, Bates job, version, document pointer and audit commit together; failure rolls back the range and removes the generated file.

### Audit and notifications

| Method | Route | Access |
|---|---|---|
| GET | `/api/audit-logs` | Admin |
| GET | `/api/notifications` | authenticated |
| POST | `/api/notifications/:id/read` | authenticated owner |

Audit filters:

```text
actor_user_id
document_id
action_type
request_id
source_ip
created_from
created_to
```

Time values are RFC3339. `source_ip` is converted to a deterministic lookup key; returned source IP and `metadata` are decrypted from protected columns. Metadata is returned as structured JSON. Notification acknowledgement and its audit commit together.

## Verification commands

The repository defines these commands for normal lifecycle use:

```bash
bash run_tests.sh
bash scripts/generate_swagger.sh
```

A static reviewer does not execute them. Generated Swagger is consulted only when it already belongs to the exact inspected revision.
