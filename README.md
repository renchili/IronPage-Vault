# IronPage Vault

IronPage Vault is an offline legal PDF lifecycle management **backend API** for air-gapped legal, compliance, and regulated document environments. It uses Go, Echo, sqlx, PostgreSQL, and local filesystem storage.

This is not a fullstack product. Files under `public/` are acceptance-only backend testing aids and are served only when explicit acceptance mode is enabled.

## What the backend provides

The API covers:

- local identity, server-side sessions, token revocation, freshness checks, and replay protection.
- Admin, Editor, and Reviewer authorization boundaries.
- PDF intake, version records, local binary storage, and revision limits.
- ordered workflow from Draft through Finalized, with terminal immutability.
- staged redaction followed by confirmed PDF burn-in.
- reviewer annotations and dispositions.
- visible Bates numbering and sequence records.
- document comparison, audit records, notifications, configuration, backup, and restore.

Implementation and acceptance claims must be read together with the exact test revision. Historical successful regression evidence is not the same as a fresh run for the current HEAD.

## Repository map

```text
cmd/server/          process entrypoint
internal/app/        Echo routes, middleware, HTTP mapping, runtime configuration
internal/core/       domain rules, roles, workflow, access policy, validation
internal/service/    use-case orchestration
internal/repository/ repository interfaces and persistence operations
internal/store/      SQL-facing storage helpers
internal/platform/   PDF, crypto, digest, filesystem, backup, restore adapters
migrations/          PostgreSQL schema
API_tests/           stateful API acceptance scripts
unit_tests/          static and repository contract checks
ci/                  Docker acceptance and full-regression entrypoints
testdata/            local PDF and CSV fixtures
docs/                API, design, security, deployment, testing, and operations docs
public/              acceptance-only browser aids
```

Start with `cmd/server/main.go` for process startup, `internal/app/server.go` for routes and runtime assembly, `internal/app/config.go` for configuration gates, and `docs/api-spec.md` for endpoint behavior.

## Prerequisites

The supported local runtime path uses Docker with Compose. The application image includes PostgreSQL and the Go API process in one service container.

For source-level Go checks, use the Go version declared by the repository build configuration.

## Required runtime configuration

IronPage Vault has no built-in fallback for database authentication, token signing, or sensitive-field encryption. Supply these values externally:

```text
DB_PASSWORD
JWT_SECRET
AES_KEY
```

Compose rejects missing values, and the application validates them again before creating directories, connecting to PostgreSQL, running migrations, or serving HTTP.

Do not commit runtime values to the repository.

## First secure startup

A new empty database needs one explicitly configured initial Admin:

```text
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

With all required values present, start the service:

```bash
docker compose up --build
```

The bootstrap pair is used only when the user table is empty. After the initial Admin is verified and the required local identities are created, remove the bootstrap pair from the deployment environment. Existing databases with users do not require it.

## Verify startup

The API listens on port `8080` by default.

```bash
curl http://localhost:8080/healthz
```

A successful response confirms that runtime validation passed and PostgreSQL is reachable. Swagger UI is mounted at:

```text
http://localhost:8080/swagger/index.html
```

## Normal mode and acceptance mode

Normal mode is the default:

```text
ACCEPTANCE_MODE=false
```

Normal mode:

- does not create fixture identities.
- rejects acceptance seed values.
- does not serve `/ui/`.

Acceptance mode is limited to isolated CI or local acceptance runs. It requires externally supplied values for all three fixture identities:

```text
ACCEPTANCE_MODE=true
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

Bootstrap Admin values and acceptance fixture values are mutually exclusive. The repository and browser pages contain no fixture passwords.

The acceptance UI is then available at:

```text
http://localhost:8080/ui/
```

The UI is a backend probe surface only. Screenshot acceptance proves that it loads and renders; it does not prove complete login, retry, accessibility, or recovery interactions.

## Run checks

Go and static repository checks:

```bash
go test ./...
bash unit_tests/test_rules.sh
```

Docker build, Compose startup, generated execution-scoped fixture values, and stateful API acceptance:

```bash
bash ci/docker_acceptance.sh
```

Full regression and retained artifacts:

```bash
bash ci/run_full_regression.sh artifacts/regression
```

A validation report must identify the exact commit SHA and corresponding run or artifact. Do not report current-HEAD PASS from an older run.

## Generated API documentation

OpenAPI documentation is generated from Swaggo annotations in Go source. Do not edit generated Swagger files as the source of truth.

```bash
bash scripts/generate_swagger.sh
```

Generated files are written under `docs/swagger/`.

## Roles

| Role | Responsibility |
|---|---|
| Admin | local user management, configuration, workflow definitions, templates, backup operations |
| Editor | PDF upload, versions, redaction confirmation, Bates numbering, finalization |
| Reviewer | document retrieval, annotations, dispositions, permitted review transitions |

Admin is intentionally not treated as a document editor.

## Document lifecycle

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Finalized is the terminal state. Upload replacement, rollback, redaction, annotation mutation, Bates numbering, workflow transition, and metadata mutation must be rejected after finalization.

## Storage and backup

PostgreSQL stores metadata, identity/session state, workflow history, audit records, notifications, configuration, and backup records. The local filesystem stores PDF binaries and transformed versions.

Persistent container paths:

```text
/var/lib/postgresql/data
/var/lib/ironpage/storage
/var/lib/ironpage/backups
```

Strict backup requires both a PostgreSQL custom-format dump and a tar snapshot of local PDF storage. Restore requires both artifacts so database paths and files return to a consistent recovery point.

See `docs/backup-recovery.md` and `docs/pitr.md` for the supported recovery scope and evidence boundaries.

## Troubleshooting

**Compose stops before building**  
A required runtime variable is absent. Check `DB_PASSWORD`, `JWT_SECRET`, and `AES_KEY` in the calling environment.

**A new normal-mode database exits during startup**  
The user table is empty and the initial Admin pair was not supplied.

**Acceptance users are not created**  
Confirm that acceptance mode is explicitly enabled and all three fixture values are present.

**`/ui/` returns 404**  
This is expected in normal mode. The test UI is acceptance-only.

**Health returns database unavailable**  
Check the local PostgreSQL process, database configuration, and persistent volume state.

## Deeper documentation

- `docs/api-spec.md` — API contract and examples.
- `docs/design.md` — architecture, boundaries, data flow, and validation strategy.
- `docs/security.md` — security model.
- `docs/rbac.md` — role and object-access rules.
- `docs/usage.md` — operational API examples.
- `docs/testing.md` — local and Docker test coverage.
- `docs/deployment-offline.md` — offline deployment configuration.
- `docs/backup-recovery.md` — strict backup and restore.
- `docs/pitr.md` — documented recovery strategy and current limitations.
