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

Collection endpoints use `page` and `page_size`; default is `25` and maximum is `100`. Values below one normalize to the default rules, values above the configured maximum clamp to that maximum, and extremely large page values clamp before offset multiplication.

During restore maintenance, non-owner requests return `503` with `error.code=MAINTENANCE_MODE`. A second restore request returns `409` with `error.code=RESTORE_ALREADY_RUNNING`.

## Roles

| Role | Responsibilities |
|---|---|
| Admin | users, validated configuration, persisted workflow definitions, templates, audit logs, backup and restore |
| Editor | owned-document upload, batch import, rollback, redaction confirmation, Bates, finalization |
| Reviewer | readable-document review, annotations/dispositions, permitted workflow transitions |

Object policy is evaluated in addition to route role checks. Admin is not implicitly an Editor. The reserved system scheduler identity is used only for automated audit attribution and is omitted from the user collection.

## Route inventory

### Health and identity

| Method | Route | Access |
|---|---|---|
| GET | `/healthz` | public, except restore maintenance |
| POST | `/api/auth/login` | public, except restore maintenance |
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
| POST | `/api/admin/backup/restore/:id/resolve` | Admin |

#### Configuration ownership

The generic config PATCH accepts only:

```text
pagination.default_page_size
pagination.max_page_size
```

The request is:

```json
{"value": "25"}
```

Both persisted pagination rows are locked and validated as one pair before update. The pair must satisfy `1 <= default <= max <= 100`.

| Condition | Status | Error code |
|---|---:|---|
| non-integer or invalid pagination pair | 400 | `INVALID_CONFIG_VALUE` |
| unknown generic key | 400 | `CONFIG_KEY_NOT_MANAGED` |
| deployment-owned `backup.local_volume` | 409 | `CONFIG_KEY_READ_ONLY` |

#### Workflow definitions

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

#### Backup and restore

Manual backup acquires an exclusive application mutation barrier before metadata collection, PostgreSQL dump and filesystem snapshot. It returns both strict artifact paths. Scheduled backup uses the same barrier and attributes its job/audit to the reserved system principal.

Restore requires both paths:

```json
{
  "database_dump_path": "/configured/backup/bak_example.dump",
  "file_snapshot_path": "/configured/backup/bak_example_files.tar"
}
```

The route owns code-enforced maintenance for the full authentication and restore lifecycle. It records a restore ID and Requested state, then Completed or Failed only after the platform result is known. `200` is returned only after Completed state and audit are persisted.

A process interruption before a durable result is recorded becomes `Interrupted` with unknown outcome. After inspecting the database and filesystem, an Admin resolves it without rerunning restore:

```http
POST /api/admin/backup/restore/rst_example/resolve
```

```json
{
  "status": "Completed",
  "verification_note": "Verified representative rows, paths, audit records, and PDF hashes"
}
```

`status` must be `Completed` or `Failed`; `verification_note` is required. Missing journal returns `RESTORE_RECONCILIATION_NOT_FOUND`; a non-Interrupted record returns `RESTORE_NOT_INTERRUPTED`.

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

Time values are RFC3339. `source_ip` is converted to a deterministic lookup key; returned source IP and `metadata` are decrypted from protected columns. Metadata is returned as structured JSON. Notification acknowledgement and its audit commit together. Every audit row requires a non-empty acting user, including scheduled and reconciliation work.

## Verification commands

The repository defines these commands for normal lifecycle use:

```bash
bash run_tests.sh
bash scripts/generate_swagger.sh
```

A static reviewer does not execute them. Generated Swagger is consulted only when it already belongs to the exact inspected revision.
