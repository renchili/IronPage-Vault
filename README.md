# IronPage Vault

IronPage Vault is an offline legal PDF lifecycle management backend for document upload, review, annotation, redaction workflow, Bates numbering, workflow transitions, audit logs, notifications, configuration, and local backup metadata.

`AGENT.md` is the authoritative project rule file. `CLAUDE.md` only references it so the project has one source of implementation rules.

## Stack

- Go
- Echo
- sqlx
- PostgreSQL
- local filesystem PDF storage
- Docker Compose startup
- informal manual test page only; no formal frontend requirement

## Start

```bash
docker compose up --build
```

API base URL:

```text
http://localhost:8080
```

Health check:

```bash
curl http://localhost:8080/healthz
```

## Seed Users

The application seeds three local accounts for acceptance testing:

| Role | Username | Password |
|---|---|---|
| Admin | admin | Admin123! |
| Editor | editor | Editor123! |
| Reviewer | reviewer | Reviewer123! |

## Required Auth Headers

Authenticated endpoints require:

```text
Authorization: Bearer <token>
X-Request-ID: unique request id
X-Request-Timestamp: RFC3339 UTC timestamp
```

## Login

```bash
curl -s http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -H "X-Request-ID: req_login_$(date +%s)" \
  -d '{"username":"editor","password":"Editor123!"}'
```

## Upload a PDF

Use the editor token and a local sample PDF:

```bash
curl -s http://localhost:8080/api/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: req_upload_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -F "title=Sample Contract" \
  -F "file=@testdata/pdfs/sample_contract.pdf"
```

## Main API Areas

- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`
- `POST /api/admin/users`
- `GET /api/admin/users`
- `GET /api/admin/config`
- `PATCH /api/admin/config/:key`
- `POST /api/admin/backup/run`
- `POST /api/documents`
- `POST /api/documents/batch`
- `GET /api/documents`
- `GET /api/documents/:id`
- `GET /api/documents/:id/versions`
- `POST /api/documents/:id/workflow/transition`
- `POST /api/documents/:id/finalize`
- `POST /api/documents/:id/redactions`
- `POST /api/documents/:id/redactions/:redaction_id/confirm`
- `POST /api/documents/:id/annotations`
- `PATCH /api/annotations/:id/disposition`
- `POST /api/documents/:id/bates`
- `GET /api/audit-logs`
- `GET /api/notifications`
- `POST /api/notifications/:id/read`

## Role Matrix

| Feature | Admin | Editor | Reviewer |
|---|---:|---:|---:|
| Users and config | yes | no | no |
| Backup job metadata | yes | no | no |
| PDF upload | no | yes | no |
| Redaction and Bates | no | yes | no |
| Annotation | no | no | yes |
| Workflow movement | no | yes | yes |
| Finalization | no | yes | no |
| Read documents | yes | yes | yes |

## Workflow

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Finalized documents are immutable. Mutation APIs must reject changes after finalization.

## Test Data

Test resources are kept under:

```text
testdata/pdfs/
testdata/csv/
```

The repository includes small local files so acceptance can run without internet access.

## Test Runner

Acceptance tests are organized as:

```text
unit_tests/
API_tests/
run_tests.sh
```

Run all tests with:

```bash
./run_tests.sh
```

The script should print a clear summary with total, passed, and failed counts.

## Informal Test Frontend

A non-production manual test page is provided at:

```text
public/manual-test.html
```

It is only for manual verification. The backend API and test scripts remain the acceptance source.

## Documentation

Detailed documentation should be kept in `docs/` and must stay consistent with implementation.
