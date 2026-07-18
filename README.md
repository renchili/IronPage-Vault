# IronPage Vault

IronPage Vault is an offline legal PDF lifecycle management **backend API** for air-gapped legal, compliance, and regulated document environments. It uses Go, Echo, sqlx, PostgreSQL, and local filesystem storage.

This is not a fullstack product. Files under `public/` are acceptance-only backend testing aids and are served only when explicit acceptance mode is enabled.

## What the backend provides

The API covers:

- local identity, rolling failed-login lockout, server-side sessions, token revocation, freshness checks, and replay protection.
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
scripts/deploy.sh    one-command secure local deployment
API_tests/           stateful API, bootstrap, authentication, and browser acceptance
unit_tests/          static and repository contract checks
ci/                  Docker acceptance and full-regression entrypoints
testdata/            local PDF and CSV fixtures
docs/                API, design, security, deployment, testing, and operations docs
public/              acceptance-only browser aid
```

Start with `cmd/server/main.go` for process startup, `internal/app/server.go` for routes and runtime assembly, `internal/app/config.go` for application configuration gates, `docker-compose.yml` for the single-container runtime, and `docs/api-spec.md` for endpoint behavior.

## Prerequisites

The supported local runtime path requires:

- Docker with Docker Compose v2.
- Bash and standard local tools including `od`, `sed`, `tr`, and `cut`.

The application image contains PostgreSQL and the Go API process in one service container. No external database or network service is required.

For source-level Go checks, use the Go version declared by the repository build configuration.

## First 10 minutes: one-command deployment

From the repository root, run:

```bash
bash scripts/deploy.sh
```

On the first run, the deployer:

1. creates a local `.env` runtime file with mode `0600`;
2. generates random database, JWT-signing, AES-encryption, and initial Admin credentials;
3. configures the embedded PostgreSQL instance and API from that file;
4. builds and starts the single Compose service in the background;
5. waits for `/healthz` to succeed; and
6. prints the initial Admin username and password once.

The generated `.env` is excluded from Git and from the Docker build context. Product code, image layers, Compose defaults, documentation, and browser assets do not contain a fixed credential or cryptographic key.

Repeated execution uses the existing `.env` instead of rotating the database password or encryption keys:

```bash
bash scripts/deploy.sh
```

## Verify startup

The API listens on port `8080` by default.

```bash
curl http://localhost:8080/healthz
```

A successful response confirms that runtime validation passed and PostgreSQL is reachable. Swagger UI is mounted at:

```text
http://localhost:8080/swagger/index.html
```

## Database and runtime configuration

The one-command deployer writes the complete local runtime configuration to `.env`. The embedded PostgreSQL process and the Go API use the same database identity.

| Variable | Generated/default value | Purpose |
|---|---|---|
| `DB_HOST` | `127.0.0.1` | Embedded PostgreSQL address inside the container |
| `DB_PORT` | `5432` | Embedded PostgreSQL port |
| `DB_USER` | `ironpage` | PostgreSQL role used by both PostgreSQL initialization and the API |
| `DB_PASSWORD` | Randomly generated | PostgreSQL password used by both PostgreSQL initialization and the API |
| `DB_NAME` | `ironpage` | PostgreSQL database used by both PostgreSQL initialization and the API |
| `JWT_SECRET` | Randomly generated | Local JWT signing material |
| `AES_KEY` | Randomly generated | Sensitive-column encryption material |
| `BOOTSTRAP_ADMIN_USERNAME` | Randomly generated | Initial empty-database Admin identity |
| `BOOTSTRAP_ADMIN_PASSWORD` | Randomly generated | Initial empty-database Admin password |

`docker-compose.yml` derives `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` from `DB_USER`, `DB_PASSWORD`, and `DB_NAME`. This keeps the embedded database initialization and the application connection configuration consistent.

The Go application still has no sensitive fallback values. Starting Compose without a complete runtime file or equivalent externally supplied values fails closed. The deployment script is responsible for creating those external values securely on first use.

To customize a clean installation, generate the file without starting containers, edit it, and then deploy:

```bash
IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
# Edit .env while preserving strong unique secret values.
bash scripts/deploy.sh
```

Do not casually change `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, or `AES_KEY` after persistent data exists. Database credentials must be migrated inside PostgreSQL, and changing `AES_KEY` makes existing encrypted metadata unreadable. For a disposable clean re-test, remove the volumes first.

After the initial Admin login is verified, remove `BOOTSTRAP_ADMIN_USERNAME` and `BOOTSTRAP_ADMIN_PASSWORD` from `.env`. Existing users are preserved and restart does not create or overwrite another Admin.

## Manual Compose operation

The deployment script is the supported first-run path. Once `.env` exists, the equivalent manual start is:

```bash
docker compose --env-file .env up --build -d
```

Stop while preserving data:

```bash
docker compose --env-file .env down
```

Remove local volumes for a clean re-test:

```bash
docker compose --env-file .env down -v
```

A clean normal-mode database again needs bootstrap values. Running `bash scripts/deploy.sh` with a missing or empty `.env` generates a new complete configuration.

## Normal mode and acceptance mode

Normal mode is the deployment default:

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

The UI is a backend probe surface only. Screenshot acceptance proves rendering only. `API_tests/test_ui_interaction_acceptance.sh` exercises actual submission, errors, recovery, keyboard focus, and retry behavior and writes evidence without recording the supplied password.

## Run checks

Go and static repository checks:

```bash
go test ./...
bash unit_tests/test_rules.sh
```

Docker build, one-command normal-mode bootstrap and restart, rolling lockout, authentication fault injection, browser interaction, and stateful API acceptance:

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

**`scripts/deploy.sh` reports that Docker is missing**  
Install Docker with the Compose v2 plugin, then rerun the same command.

**The generated runtime file already exists**  
This is expected. The deployer reuses `.env` so persistent database and encryption credentials remain stable.

**Compose reports a missing required variable**  
The existing `.env` is incomplete or malformed. Restore the required database and cryptographic variables, or remove the empty/disposable file and rerun the deployer to generate a complete one.

**A new normal-mode database exits during startup**  
The user table is empty but bootstrap variables were removed from `.env`. Restore an explicit bootstrap pair or regenerate the disposable installation after removing its volumes.

**`/ui/` returns 404**  
This is expected in normal mode. The test UI is acceptance-only.

**Health returns database unavailable**  
Check the container logs, the database variables in `.env`, and the persistent volume state. Do not change only one side of an initialized database credential.

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
