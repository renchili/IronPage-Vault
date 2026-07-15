# Offline Deployment

IronPage Vault runs as one local Compose service in an air-gapped environment.

## Required configuration

The service requires externally supplied values for:

```text
DB_PASSWORD
JWT_SECRET
AES_KEY
```

There are no built-in fallback values. Compose and application startup reject missing or weak values.

## First normal-mode startup

A new empty database also requires:

```text
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

The application uses this pair only when the user table is empty. After the first Admin is created and verified, remove the bootstrap variables from the deployment environment. Existing databases with users do not require them.

Normal mode does not create acceptance fixture users and does not serve the backend test UI.

## Startup

```bash
docker compose up --build
```

After startup, verify:

```text
http://localhost:8080/healthz
http://localhost:8080/swagger/index.html
```

## Acceptance mode

Acceptance mode is for isolated CI or local acceptance only:

```text
ACCEPTANCE_MODE=true
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

All three fixture values are required and must be supplied externally. Bootstrap variables and acceptance fixture variables cannot be used together.

Only acceptance mode serves:

```text
http://localhost:8080/ui/
```

The repository and browser pages contain no fixture values.

## Runtime layout

The image contains PostgreSQL, the Go API, migrations, local PDF storage, local backup storage, and acceptance UI assets.

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

## Other runtime variables

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

## Stop and reset

Stop while preserving data:

```bash
docker compose down
```

Remove local volumes for a clean re-test:

```bash
docker compose down -v
```

A clean normal-mode database again requires the bootstrap pair. A clean acceptance run again requires externally supplied acceptance fixture values.

## Acceptance checks

- missing required runtime values prevent startup.
- an empty normal-mode database requires explicit initial Admin configuration.
- seed users require explicit acceptance mode.
- bootstrap and acceptance identity modes are mutually exclusive.
- the test UI is unavailable in normal mode and contains no embedded credentials.
- no external identity, PDF, database, storage, notification, or network service is required.
