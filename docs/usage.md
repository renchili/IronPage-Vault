# Usage and Acceptance Guide

This document covers secure startup, API use, and isolated acceptance runs. IronPage Vault is a pure backend API project. The browser UI is an acceptance-only backend testing aid.

## Required runtime configuration

Supply these values externally before startup:

```text
DB_PASSWORD
JWT_SECRET
AES_KEY
```

The service provides no fallback values for them.

## First normal-mode startup

A new empty database also requires this pair:

```text
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

Start the service:

```bash
docker compose up --build
```

The pair creates one initial Admin only when the user table is empty. After verifying the account and creating the required local users, remove the bootstrap pair from the deployment environment. An existing user database does not require it.

## Subsequent normal-mode startup

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

A successful startup means runtime configuration passed validation, PostgreSQL is reachable, local data paths are available, and the health endpoint returns OK.

Normal mode does not create acceptance users and does not serve the backend test UI.

## Explicit acceptance mode

An isolated acceptance run uses:

```text
ACCEPTANCE_MODE=true
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

All three fixture values must be supplied externally. The CI acceptance script generates execution-scoped values. Bootstrap Admin variables and acceptance fixture variables cannot be used together.

## Backend test UI

Only acceptance mode serves:

```text
http://localhost:8080/ui/
```

The page contains blank identity fields and no embedded credentials. Screenshot acceptance proves rendering only; it does not prove complete UI interaction or recovery behavior.

## API authentication flow

Use an identity supplied by the current local deployment. Store the returned bearer token locally:

```bash
TOKEN='paste_token_here'
```

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

Local unit and static checks:

```bash
go test ./...
bash unit_tests/test_rules.sh
```

Docker/API acceptance with generated execution-scoped fixture values:

```bash
bash ci/docker_acceptance.sh
```

## Local data locations

```text
/var/lib/postgresql/data
/var/lib/ironpage/storage
/var/lib/ironpage/backups
```

## Stop and reset

```bash
docker compose down
```

Clean local re-test:

```bash
docker compose down -v
```

## Evidence boundary

These commands describe the supported local path. Validation reports must still identify the exact commit SHA, run ID, and artifact digest when available.
