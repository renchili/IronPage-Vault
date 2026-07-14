# Usage and Acceptance Guide

This document contains startup, operation, manual backend testing, and acceptance commands. The README focuses on project purpose and implementation.

IronPage Vault is a pure backend API project. The UI mentioned here is only a manual backend testing aid.

## Start the backend system

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

A successful local startup is not just a running process. It should also prove that the configured database connection, storage directory, and required runtime settings are available.

## Runtime-sensitive values

Do not publish reusable authentication material, signing material, encryption material, or deployment-only values in documentation.

For secure operation, sensitive runtime values must be supplied externally through the deployment environment. Local acceptance fixture values must be treated as acceptance-only and must not be documented as safe deployment defaults.

If the current code or Compose configuration still provides predictable local fixtures, that is a security remediation item, not a recommended usage pattern.

## Backend test UI

The repository includes a lightweight backend test page served at:

```text
http://localhost:8080/ui/
```

This page is only for manual acceptance of backend APIs. It is not a production frontend, not a formal UI requirement, and not a separate fullstack deliverable.

The screenshot acceptance script verifies that `/ui/` loads and can be rendered by a headless browser. It does not prove full login interaction, retry behavior, accessibility, or API state transitions from UI clicks.

## API authentication flow

Use a local acceptance identity only in an explicit acceptance environment. Do not copy fixture values into production or shared documentation.

The login request returns a bearer token. Store that token locally for subsequent API calls:

```bash
TOKEN="paste_token_here"
```

## Authenticated request headers

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

## Upload sample PDF

```bash
curl -s http://localhost:8080/api/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: req_upload_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -F "title=Sample Contract" \
  -F "file=@testdata/pdfs/sample_contract.pdf"
```

## List documents

```bash
curl -s http://localhost:8080/api/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Request-ID: req_docs_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## Workflow transition

```bash
curl -s http://localhost:8080/api/documents/$DOC_ID/workflow/transition \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req_workflow_$(date +%s%N)" \
  -H "X-Request-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -d '{"status":"Under Review"}'
```

## Run tests

```bash
./run_tests.sh
```

The test runner should execute local unit/static checks and API acceptance checks, then print a summary with total, passed, and failed counts.

## Local data locations

Inside the container:

```text
/var/lib/postgresql/data
/var/lib/ironpage/storage
/var/lib/ironpage/backups
```

Docker volumes are declared in `docker-compose.yml`.

## Stop the system

```bash
docker compose down
```

To remove local volumes during a clean re-test:

```bash
docker compose down -v
```

## Evidence boundary

The usage commands above describe the intended local acceptance path. They are not a replacement for a current-HEAD full regression. When reporting validation, include the exact commit SHA, run ID, and artifact digest when available.
