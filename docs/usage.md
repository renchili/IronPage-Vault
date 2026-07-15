# Usage and Acceptance Guide

This document contains startup, operation, manual backend testing, and acceptance commands. IronPage Vault is a pure backend API project. The browser UI is an acceptance-only backend testing aid.

## Supply required runtime values

The service does not provide fallback database, signing, or encryption values. Supply them externally before startup:

```bash
export DB_PASSWORD='<strong local database password>'
export JWT_SECRET='<at least 32 characters of local signing material>'
export AES_KEY='<at least 32 characters of local encryption material>'
```

Do not commit or publish the actual values.

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

A successful local startup proves that configuration validation passed, PostgreSQL is reachable, the storage and backup directories are available, and the health endpoint returns an OK status.

## Default operating mode

Default mode does not create acceptance identities and does not serve the backend test UI. Seed-user variables are not accepted unless acceptance mode is explicitly enabled.

## Explicit acceptance mode

For an isolated acceptance run, supply temporary fixture values and explicitly enable the mode:

```bash
export ACCEPTANCE_MODE=true
export SEED_ADMIN_PASSWORD='<acceptance-only value>'
export SEED_EDITOR_PASSWORD='<acceptance-only value>'
export SEED_REVIEWER_PASSWORD='<acceptance-only value>'
docker compose up --build
```

The CI acceptance script generates these values for each execution. They are not stored in the repository.

## Backend test UI

When acceptance mode is enabled, the backend test page is served at:

```text
http://localhost:8080/ui/
```

The page contains blank username and password fields. It does not embed acceptance credentials.

The screenshot acceptance script verifies that `/ui/` loads and can be rendered by a headless browser. It does not prove full login interaction, retry behavior, accessibility, or API state transitions from UI clicks.

## API authentication flow

Use a local identity supplied by the current deployment or acceptance environment. The login response returns a bearer token. Store it locally for subsequent requests:

```bash
TOKEN='paste_token_here'
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

Local unit and static checks:

```bash
go test ./...
bash unit_tests/test_rules.sh
```

Docker/API acceptance, including generated temporary runtime and fixture values:

```bash
bash ci/docker_acceptance.sh
```

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

These commands describe the supported local path. They are not a replacement for a current-HEAD full regression. Validation reports must include the exact commit SHA, run ID, and artifact digest when available.
