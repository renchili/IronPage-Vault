# IronPage Vault

IronPage Vault is an offline legal PDF lifecycle management **backend API** for air-gapped legal, compliance, and regulated-document environments. It uses Go, Echo, sqlx, PostgreSQL, and local filesystem storage in one standalone container.

This is not a full-stack product. `public/index.html` is the single acceptance-only browser probe and is served only when acceptance mode is explicitly enabled.

## Capabilities

The backend implements:

- local identity, rolling failed-login lockout, server-side sessions, token revocation, freshness validation, and replay protection;
- Admin, Editor, and Reviewer authorization boundaries;
- PDF intake, local binary storage, version history, and revision limits;
- an Admin-managed persisted workflow chain with terminal Finalized immutability;
- staged redaction followed by confirmed strict PDF burn-in;
- reviewer annotations, dispositions, and local mention notifications;
- visible Bates numbering with transactional page-range allocation;
- structured text/page/bounding-box comparison;
- encrypted audit source/metadata with Admin filtering and decrypted responses;
- validated configuration, mutation-isolated backup, maintenance-mode restore, and explicit restore lifecycle reconciliation.

Material database mutations include their required audit, history, and notification side effects in the same transaction. File-producing mutations remove generated output when database persistence fails.

Implementation claims and execution evidence are revision-specific. A historical successful run is not evidence for another commit.

## Repository layout

```text
cmd/server/          process entrypoint
internal/app/        Echo routes, transaction assembly, HTTP mapping, runtime configuration
internal/core/       domain rules, roles, default workflow compatibility, access policy, validation
internal/service/    use-case orchestration
internal/repository/ persistence operations and query models
internal/store/      SQL-facing helpers
internal/platform/   PDF, crypto, digest, filesystem, backup, staged restore adapters
migrations/          PostgreSQL schema and upgrade migrations
tests/api/           stateful HTTP, Docker, and browser acceptance definitions
tests/contracts/     repository, structure, and generated-contract checks
testdata/            local PDF and CSV fixtures
ci/                  static workflow contracts and manual regression helpers
scripts/deploy.sh    one-command secure local deployment
docs/                API, design, security, deployment, testing, and operations docs
public/index.html    canonical acceptance-only browser probe
```

## Prerequisites

The supported deployment path requires Docker with Docker Compose v2, Bash, and standard local tools including `od`, `sed`, `tr`, `cut`, `awk`, `getent`, and `timeout`. No external database, identity provider, PDF service, object store, notification service, or runtime internet connection is required.

## One-command deployment

From the repository root:

```bash
bash scripts/deploy.sh
```

The first run resolves an IPv4 loopback address, selects a currently unused loopback host port, creates `.env` with mode `0600`, generates installation-specific database identity/ports/paths/secrets/Admin credentials, builds the image, starts the one service, waits for health, and prints the actual URLs and initial Admin pair.

The generated `.env` is excluded from Git and the Docker context. Re-running the command reuses it. Do not assume `localhost:8080`; read `HOST_BIND_ADDRESS` and `HOST_PORT`. The availability probe reduces initial collisions, while Docker remains the final bind authority.

## Runtime configuration ownership

The deployment layer supplies every local runtime value. The schema does not seed a fixed machine backup path. After migrations, startup persists the generated `BACKUP_DIR` and paging limits into `config_entries`.

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

`backup.local_volume` is deployment-owned and read-only through the Admin API. The only Admin-managed generic configuration keys are `pagination.default_page_size` and `pagination.max_page_size`; every update is validated transactionally against `1 <= default <= max <= 100`. Unknown keys are rejected.

To generate and inspect a clean installation file before startup:

```bash
IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
```

Do not change database identity/password, AES key, persistent paths, or listener ports after data exists without a deliberate migration. After verifying the first Admin login, remove the bootstrap pair from `.env`.

## Normal and acceptance modes

Normal mode creates no fixture identities and does not serve `/ui/`. Acceptance mode requires execution-scoped fixture passwords and serves the canonical probe at:

```text
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/ui/
```

The browser page contains no fixture credential.

## Roles and workflow

| Role | Responsibility |
|---|---|
| Admin | local users, configuration, persisted workflow definitions, templates, backup/restore operations |
| Editor | owned-document upload, versions, redaction confirmation, Bates numbering, finalization |
| Reviewer | document retrieval, annotations, dispositions, permitted review transitions |

Admin is not treated as a document editor.

The initial chain is:

```text
Draft -> Under Review -> Redaction Pending -> Approved -> Finalized
```

Admin can read and replace the ordered chain:

```text
GET /api/admin/workflow-statuses
PUT /api/admin/workflow-statuses
```

`Draft` must remain first and mutable; `Finalized` must remain last and immutable; existing document statuses cannot be removed. Runtime transitions read the persisted order. A transition/finalization commits document state, history, audit, and owner notification together.

## Audit and notification integrity

Audit source IP and structured metadata are stored in ciphertext columns. Source IP also has a deterministic equality lookup. Startup backfills the lookup for older rows. The Admin audit route decrypts source IP and JSON metadata before response; compatibility plaintext columns are not the source of truth.

Every audit write requires a non-empty acting user. Scheduled backup and startup reconciliation use a protected, non-interactive system principal instead of `NULL`; this principal is omitted from the Admin user collection. User/config/template/workflow changes, upload/rollback, workflow/finalization, redaction, annotation, Bates, backup, notification acknowledgement, failed login, successful login, and logout all fail rather than report success if their required audit cannot be persisted. Workflow and mention notifications share the parent mutation transaction.

## Storage, backup, and restore

PostgreSQL stores metadata/security/workflow/audit/notification/configuration/backup state. The local filesystem stores PDF versions and transformed output.

All unsafe API mutations acquire a shared PostgreSQL advisory lock. Manual and scheduled backup acquire the matching exclusive lock before collecting metadata, running `pg_dump`, and archiving `STORAGE_DIR`; no application mutation can cross the database-dump/filesystem-snapshot interval. This is the application recovery boundary for the supported single-container deployment. Failed job/audit persistence removes the generated artifacts.

Restore enters code-enforced maintenance before authentication and restore work: new non-restore requests receive `MAINTENANCE_MODE`, active requests drain, and an exclusive advisory lock blocks application mutations. Strict restore safely extracts the archive to staging, rejects path traversal, links, and special entries, swaps the storage directory with a rollback copy, and invokes `pg_restore --single-transaction`. PostgreSQL failure restores the previous filesystem directory.

The encrypted lifecycle records `Requested`, then `Completed` or `Failed` when the platform result is known. A process exit before that result is durable becomes `Interrupted` with an unknown outcome, never an inferred failure. An Admin resolves an Interrupted record through `POST /api/admin/backup/restore/:id/resolve` after verifying the restored database and files. PostgreSQL subprocess passwords are supplied through a short-lived mode-`0600` `PGPASSFILE`, not command-line arguments.

See `docs/backup-recovery.md` and `docs/pitr.md`.

## Verification entrypoints

Local staged report:

```bash
bash run_tests.sh
```

Without `BASE_URL` and all three `SEED_*_PASSWORD` values, stateful rows are `SKIP`, the report is `INCOMPLETE`, and exit status is `2`.

Manual/normal-lifecycle complete regression:

```bash
bash ci/run_full_regression.sh artifacts/regression
```

GitHub verification is defined only in `.github/workflows/ci.yml` and is static acceptance only. Admission precedes checkout. Automatic targets are derived from the event. A manual target must equal the selected branch or identify the same-repository open PR whose branch and head SHA match the selected ref. The workflow collapses active duplicates, paginates scoped history, applies cooldown/latching to the canonical target/revision, rejects ordinary reruns, and permits one exact reviewed unlock.

The later job runs static syntax, formatting, inventory, documentation, and contract gates. It does not run Docker, API, browser, deployment, or complete regression. GitHub creates the run object before YAML admission executes, so repository admission is pre-checkout rather than platform-level pre-dispatch prevention.

A static reviewer reads source and existing evidence only and must not trigger, run, retry, wait for, or validate execution to fill gaps.

## Generated API documentation

Swaggo annotations in Go source are authoritative. Supported execution entrypoints generate files under `docs/swagger/`:

```bash
bash scripts/generate_swagger.sh
```

Static review does not authorize generation.

## Documentation

- `docs/api-spec.md` — API contract and examples
- `docs/design.md` — architecture and transaction boundaries
- `docs/security.md` — security model
- `docs/rbac.md` — role and object-access rules
- `docs/usage.md` — operational API examples
- `docs/testing.md` — test and static evidence boundaries
- `docs/deployment-offline.md` — generated offline deployment
- `docs/backup-recovery.md` — strict backup and staged restore
- `docs/pitr.md` — recovery strategy and limitations
