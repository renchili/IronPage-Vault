# IronPage Vault

IronPage Vault is an offline legal PDF lifecycle management **backend API** for air-gapped legal, compliance, and regulated-document environments. It uses Go, Echo, sqlx, PostgreSQL, and local filesystem storage in one standalone container.

This is not a full-stack product. `public/index.html` is the single acceptance-only browser probe and is served only when acceptance mode is explicitly enabled.

## Capabilities

The backend implements:

- local identity, rolling failed-login lockout, server-side sessions, token revocation, freshness validation, and replay protection;
- Admin, Editor, and Reviewer authorization boundaries;
- PDF intake, local binary storage, version history, and revision limits;
- Draft → Under Review → Redaction Pending → Approved → Finalized workflow with terminal immutability;
- staged redaction followed by confirmed PDF burn-in;
- reviewer annotations and dispositions;
- visible Bates numbering and auditable sequence allocation;
- document comparison, audit records, notifications, configuration, backup, and restore.

Implementation claims and test evidence are revision-specific. A historical successful run is not evidence for another commit.

## Repository layout

```text
cmd/server/          process entrypoint
internal/app/        Echo routes, middleware, HTTP mapping, runtime configuration
internal/core/       domain rules, roles, workflow, access policy, validation
internal/service/    use-case orchestration
internal/repository/ repository interfaces and persistence operations
internal/store/      SQL-facing storage helpers
internal/platform/   PDF, crypto, digest, filesystem, backup, restore adapters
migrations/          PostgreSQL schema
tests/api/           stateful HTTP, Docker, and browser acceptance flows
tests/contracts/     repository, structure, and generated-contract checks
testdata/            local PDF and CSV fixtures
ci/                  serialized verification and full-regression orchestration
scripts/deploy.sh    one-command secure local deployment
docs/                API, design, security, deployment, testing, and operations docs
public/index.html    canonical acceptance-only browser probe
```

Start with `cmd/server/main.go`, `internal/app/server.go`, `internal/app/config.go`, `docker-compose.yml`, and `docs/api-spec.md`.

## Prerequisites

The supported deployment path requires:

- Docker with Docker Compose v2;
- Bash and standard local tools including `od`, `sed`, `tr`, `cut`, `awk`, and `getent`.

No external database, identity provider, PDF service, object store, notification service, or runtime internet connection is required.

## One-command deployment

From the repository root:

```bash
bash scripts/deploy.sh
```

The first run:

1. creates `.env` with mode `0600`;
2. generates installation-specific database identity, ports, filesystem targets, JWT material, AES material, and initial Admin credentials;
3. builds the image with the generated application root and HTTP port;
4. starts the single Compose service;
5. waits for health; and
6. prints the actual API, health, and Swagger URLs plus the initial Admin pair.

The generated `.env` is excluded from Git and from the Docker build context. Re-running the command reuses it instead of rotating persistent configuration.

Do not assume `localhost:8080`. Read the printed URL or inspect these generated values:

```text
HOST_BIND_ADDRESS
HOST_PORT
```

The corresponding endpoints are:

```text
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/healthz
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/swagger/index.html
```

## Runtime configuration ownership

The deployment layer writes every local runtime value to `.env`; the image, Compose file, and Go application do not provide an alternative fixed local configuration.

| Area | Variables |
|---|---|
| Host exposure | `HOST_BIND_ADDRESS`, `HOST_PORT` |
| API listener | `HTTP_PORT`, `HTTP_ADDR` |
| PostgreSQL | `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` |
| Cryptography | `JWT_SECRET`, `AES_KEY` |
| Initial Admin | `BOOTSTRAP_ADMIN_USERNAME`, `BOOTSTRAP_ADMIN_PASSWORD` |
| Application assets | `IRONPAGE_APP_ROOT`, `MIGRATIONS_DIR`, `PUBLIC_DIR` |
| PostgreSQL storage | `POSTGRES_VOLUME_ROOT`, `PGDATA` |
| Product storage | `IRONPAGE_VOLUME_ROOT`, `STORAGE_DIR`, `BACKUP_DIR` |
| Acceptance fixtures | `ACCEPTANCE_MODE`, `SEED_ADMIN_PASSWORD`, `SEED_EDITOR_PASSWORD`, `SEED_REVIEWER_PASSWORD` |

Compose maps `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, and `PGPORT` from the same generated `DB_*` values used by the API. Startup fails if required values are missing or inconsistent.

To inspect or customize a clean installation before startup:

```bash
IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
# Edit .env before the first build.
bash scripts/deploy.sh
```

Do not change database identity, database password, AES key, image paths, volume targets, or listener ports after data exists without a deliberate migration.

After verifying the first Admin login, remove the bootstrap pair from `.env`. Existing users remain unchanged across restart.

See `docs/deployment-offline.md` for the complete deployment contract.

## Normal and acceptance modes

Normal mode is generated by default and does not create fixture identities or serve `/ui/`.

Acceptance mode is isolated validation only. It requires execution-scoped values for all three fixture passwords and cannot be combined with bootstrap values. In that mode, the canonical browser probe is available at:

```text
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/ui/
```

The browser page contains no fixture credential.

## Verification entrypoints

Local source and report entrypoint:

```bash
bash run_tests.sh
```

Without `BASE_URL` and all three execution-scoped `SEED_*_PASSWORD` values, stateful rows are recorded as `SKIP`, the report is `INCOMPLETE`, and the command exits with status `2`. A skipped stage is never reported as a local PASS.

Complete serialized regression:

```bash
bash ci/run_full_regression.sh artifacts/regression
```

GitHub verification is defined only in `.github/workflows/ci.yml`. The workflow:

- uses one repository-and-target concurrency group;
- waits out any remaining ten-minute target cooldown before validation;
- latches a failed revision until a new commit or explicit reviewed manual unlock;
- runs one sequential job;
- stops after the first failed stage; and
- uploads evidence only after the complete regression succeeds.

`run_tests.sh` reports only stage rows that actually executed. Its lightweight contract probe is not full-regression evidence.

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

Finalized is terminal. Replacement upload, rollback, redaction, annotation mutation, Bates numbering, workflow transition, and metadata mutation must be rejected after finalization.

## Storage and recovery

PostgreSQL stores metadata, identity/session state, workflow history, audit records, notifications, configuration, and backup records. The local filesystem stores PDF binaries and transformed versions. All concrete container paths are generated into `.env` for each installation.

Strict backup requires both a PostgreSQL custom-format dump and a tar snapshot of PDF storage. Restore requires both artifacts.

See `docs/backup-recovery.md` and `docs/pitr.md`.

## Generated API documentation

Swaggo annotations in Go source are authoritative:

```bash
bash scripts/generate_swagger.sh
```

Generated files are written under `docs/swagger/`.

## Documentation

- `docs/api-spec.md` — API contract and examples
- `docs/design.md` — architecture and boundaries
- `docs/security.md` — security model
- `docs/rbac.md` — role and object-access rules
- `docs/usage.md` — operational API examples
- `docs/testing.md` — test and evidence boundaries
- `docs/deployment-offline.md` — generated offline deployment
- `docs/backup-recovery.md` — strict backup and restore
- `docs/pitr.md` — recovery strategy and limitations
