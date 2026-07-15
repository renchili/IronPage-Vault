# Offline Deployment

IronPage Vault is designed for standalone local deployment in an air-gapped environment.

## Deployment model

The project uses one Compose service:

```text
ironpage
```

The image contains:

- PostgreSQL runtime.
- Go API binary.
- database migrations.
- local PDF storage directories.
- local backup directories.
- acceptance UI files that are served only in explicit acceptance mode.
- startup script.

## Required runtime values

The service has no built-in fallback for database authentication, JWT signing, or sensitive-field encryption. Supply these values externally before Compose resolves the service:

```bash
export DB_PASSWORD='<strong local database password>'
export JWT_SECRET='<at least 32 characters of local signing material>'
export AES_KEY='<at least 32 characters of local encryption material>'
```

Do not commit these values to the repository or copy them into shared documentation.

## Startup

```bash
docker compose up --build
```

Compose fails before startup when any required runtime value is absent. The application also validates the values before creating directories, connecting to PostgreSQL, running migrations, or serving HTTP routes.

## Production/default mode

The default mode is:

```text
ACCEPTANCE_MODE=false
```

In default mode:

- no seed users are created.
- seed-user password variables are rejected when supplied.
- the backend test UI is not served.
- administrators must create and manage identities through an approved local provisioning path.

## Explicit acceptance mode

Acceptance mode is only for isolated local tests and CI. It requires all three acceptance identity values to be supplied externally:

```bash
export ACCEPTANCE_MODE=true
export SEED_ADMIN_PASSWORD='<acceptance-only value>'
export SEED_EDITOR_PASSWORD='<acceptance-only value>'
export SEED_REVIEWER_PASSWORD='<acceptance-only value>'
docker compose up --build
```

Acceptance identity values are fixtures, not deployment defaults. The repository and browser UI do not contain their values.

When acceptance mode is enabled, the backend test UI is available at:

```text
http://localhost:8080/ui/
```

## Runtime ports

```text
8080/tcp API
```

The same port serves the acceptance UI only when acceptance mode is enabled.

## Local URLs

Always available after successful startup:

```text
http://localhost:8080/healthz
http://localhost:8080/swagger/index.html
```

Acceptance mode only:

```text
http://localhost:8080/ui/
```

## Volumes

```text
ironpage_pgdata    PostgreSQL data
ironpage_storage   PDF binary storage
ironpage_backups   local backup outputs
```

## Environment variables

Required sensitive values:

```text
DB_PASSWORD
JWT_SECRET
AES_KEY
```

Common runtime values:

```text
POSTGRES_USER
POSTGRES_DB
DB_HOST
DB_PORT
DB_USER
DB_NAME
STORAGE_DIR
BACKUP_DIR
MIGRATIONS_DIR
PUBLIC_DIR
HTTP_ADDR
```

Acceptance-only values:

```text
ACCEPTANCE_MODE
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

## No external runtime dependency

The running service does not require:

- external identity provider.
- external PDF API.
- external database.
- cloud storage.
- remote notification provider.
- internet access.

## Rebuild

```bash
docker compose build --no-cache
```

## Stop

```bash
docker compose down
```

## Clean local re-test

```bash
docker compose down -v
```

This removes local volumes. Acceptance users are recreated only when the next startup explicitly enables acceptance mode and supplies new fixture values.

## Acceptance checks

- Compose refuses missing required runtime values.
- application configuration validation rejects missing or weak sensitive values.
- seed users require explicit acceptance mode.
- the test UI is unavailable in default mode and contains no embedded credentials.
- API health check works locally.
- PDF upload stores files in the storage volume.
- PostgreSQL metadata persists across container restart when volumes are kept.
- no external runtime service is required.
