# Usage and Acceptance Guide

This document contains startup, operation, manual backend testing, and acceptance commands. The README focuses on project purpose and implementation.

IronPage Vault is a pure backend API project. The UI mentioned here is only a manual backend testing aid.

## Start the Backend System

```bash
docker compose up --build
```

The API listens on:

```text
http://localhost:8080
```

Health check:

```bash
curl http://localhost:8080/healthz
```

## Seed Users

The application creates local acceptance users on startup if they do not already exist.

| Role | Username | Password |
|---|---|---|
| Admin | admin | Admin123! |
| Editor | editor | Editor123! |
| Reviewer | reviewer | Reviewer123! |

## Backend Test UI

The repository includes a lightweight backend test page at:

```text
public/manual-test.html
```

Runtime route:

```text
http://localhost:8080/ui/manual-test.html
```

This page is only for manual acceptance of backend APIs. It is not a production frontend, not a formal UI requirement, and not a separate fullstack deliverable.

The page documents the main manual backend flow:

1. log in as Editor
2. upload a sample PDF
3. list documents
4. log in as Reviewer
5. create annotations
6. update annotation disposition
7. log in as Editor
8. stage redaction
9. confirm redaction
10. apply Bates numbering
11. transition workflow
12. finalize the document
13. verify mutation rejection after Finalized
14. view audit logs
15. view notifications
16. create backup job metadata as Admin

## Login with Curl

```bash
curl -s http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -H "X-Request-ID: req_login_$(date +%s)" \
  -d '{"username":"editor","password":"Editor123!"}'
```

Store the returned token:

```bash
TOKEN="paste_token_here"
```

## Authenticated Request Headers

Authenticated endpoints require:

```text
Authorization: Bearer <token>
X-Request-ID: unique value per request
X-Request-Timestamp: RFC3339 UTC timestamp
```

Example:

```bash
curl -s http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: req_me_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## Upload Sample PDF

```bash
curl -s http://localhost:8080/api/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: req_upload_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -F "title=Sample Contract" \
  -F "file=@testdata/pdfs/sample_contract.pdf"
```

## List Documents

```bash
curl -s http://localhost:8080/api/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: req_docs_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## Workflow Transition

```bash
curl -s http://localhost:8080/api/documents/$DOC_ID/workflow/transition \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req_workflow_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -d '{"status":"Under Review"}'
```

## Run Tests

```bash
./run_tests.sh
```

The test runner should execute:

```text
unit_tests/
API_tests/
```

and print a summary with total, passed, and failed counts.

## Local Data Locations

Inside the container:

```text
/var/lib/postgresql/data
/var/lib/ironpage/storage
/var/lib/ironpage/backups
```

Docker volumes are declared in `docker-compose.yml`.

## Stop the System

```bash
docker compose down
```

To remove local volumes during a clean re-test:

```bash
docker compose down -v
```
