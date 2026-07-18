# Offline Deployment

IronPage Vault runs as one local Compose service in an air-gapped environment. The service container includes PostgreSQL, the Go API, migrations, local PDF storage, local backup storage, and acceptance-only browser assets.

## One-command first deployment

From the repository root:

```bash
bash scripts/deploy.sh
```

No database, signing, encryption, or initial Admin values need to be exported before this command. On the first run, the deployer:

1. creates `.env` with file mode `0600`;
2. generates random `DB_PASSWORD`, `JWT_SECRET`, `AES_KEY`, and initial Admin password values;
3. writes the complete embedded PostgreSQL and API configuration;
4. builds and starts the single Compose service in the background;
5. waits until `http://localhost:8080/healthz` succeeds; and
6. prints the initial Admin credentials once.

The generated `.env` is excluded by both `.gitignore` and `.dockerignore`. It is not committed and is not sent into the Docker build context.

Running the same command again reuses the existing runtime file:

```bash
bash scripts/deploy.sh
```

This prevents an ordinary restart from silently rotating the database password, JWT signing material, or AES encryption key.

## Verify startup

```text
http://localhost:8080/healthz
http://localhost:8080/swagger/index.html
```

A successful health response means application configuration passed validation and the embedded PostgreSQL instance is reachable.

## Generated database configuration

The one-command deployer creates these database values in `.env`:

| Variable | Initial value | Use |
|---|---|---|
| `DB_HOST` | `127.0.0.1` | PostgreSQL inside the same container |
| `DB_PORT` | `5432` | Embedded PostgreSQL port |
| `DB_USER` | `ironpage` | PostgreSQL role and API connection user |
| `DB_PASSWORD` | Randomly generated | PostgreSQL and API connection password |
| `DB_NAME` | `ironpage` | PostgreSQL database and API connection database |

`docker-compose.yml` maps the PostgreSQL initialization variables from the same application values:

```text
POSTGRES_USER     <- DB_USER
POSTGRES_PASSWORD <- DB_PASSWORD
POSTGRES_DB       <- DB_NAME
```

The container entrypoint checks those pairs before PostgreSQL starts. A mismatch fails startup instead of allowing PostgreSQL and the API to use different credentials.

The runtime file also contains randomly generated `JWT_SECRET` and `AES_KEY` values. The application itself still rejects missing or weak sensitive values and contains no built-in sensitive fallback.

## Initial Admin

A new empty database needs one initial Admin. The one-command deployer generates:

```text
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

The pair is used only while the user table is empty. After the initial Admin can log in, remove both lines from `.env`. Existing users are preserved and subsequent restarts do not create or overwrite another Admin.

## Custom clean installation

To generate the runtime file without starting the service:

```bash
IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
```

Edit `.env`, preserving strong unique secret values, then start:

```bash
bash scripts/deploy.sh
```

Database identity changes are intended for a clean installation. Do not edit only `DB_PASSWORD`, `DB_USER`, or `DB_NAME` after PostgreSQL has initialized persistent data. Existing PostgreSQL credentials must be migrated deliberately. Do not change `AES_KEY` for an existing installation because previously encrypted metadata depends on it.

## Manual Compose operation

The deployment script is the supported first-run path. Once `.env` exists, manual operation is available:

```bash
docker compose --env-file .env up --build -d
docker compose --env-file .env down
```

## Acceptance mode

Acceptance mode is for isolated CI or local acceptance only:

```text
ACCEPTANCE_MODE=true
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

All three fixture values must be supplied externally for an acceptance run. Bootstrap variables and acceptance fixture variables cannot be used together.

Only acceptance mode serves:

```text
http://localhost:8080/ui/
```

The repository and browser pages contain no fixture values.

## Runtime layout

Persistent volumes:

```text
ironpage_pgdata    PostgreSQL data
ironpage_storage   PDF binary storage
ironpage_backups   local backup outputs
```

Container paths:

```text
/var/lib/postgresql/data
/var/lib/ironpage/storage
/var/lib/ironpage/backups
```

Other supported runtime variables:

```text
STORAGE_DIR
BACKUP_DIR
MIGRATIONS_DIR
PUBLIC_DIR
HTTP_ADDR
```

## Stop and reset

Stop while preserving data:

```bash
docker compose --env-file .env down
```

Remove local volumes for a clean re-test:

```bash
docker compose --env-file .env down -v
```

To generate a completely new disposable installation, remove the volumes and the old `.env`, then run:

```bash
bash scripts/deploy.sh
```

Do not remove `.env` while retaining real persistent data unless the database password and encryption keys have been backed up securely.

## Acceptance checks

- a fresh checkout can generate complete runtime configuration without pre-exported secrets;
- generated sensitive values are random, meet application length requirements, and are stored with mode `0600`;
- the runtime file is excluded from source control and Docker build context;
- the one-command path starts a healthy single-container deployment;
- an empty normal-mode database creates one generated initial Admin;
- rerunning deployment reuses the same database and cryptographic values;
- restart without bootstrap values preserves the existing Admin;
- acceptance fixtures still require explicit acceptance mode;
- the test UI remains unavailable in normal mode; and
- no external identity, PDF, database, storage, notification, or network service is required.
