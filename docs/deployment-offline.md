# Offline Deployment

IronPage Vault is designed for standalone local deployment in an air-gapped environment.

## Deployment Model

The project uses one Compose service:

```text
ironpage
```

The image contains:

- PostgreSQL runtime
- Go API binary
- database migrations
- seed logic
- local PDF storage directories
- manual test UI files
- startup script

## Startup

```bash
docker compose up --build
```

## Runtime Ports

```text
8080/tcp API and manual test UI
```

## Local URLs

```text
http://localhost:8080/healthz
http://localhost:8080/ui/manual-test.html
```

## Volumes

```text
ironpage_pgdata    PostgreSQL data
ironpage_storage   PDF binary storage
ironpage_backups   local backup outputs
```

## Environment Variables

Common variables:

```text
POSTGRES_USER
POSTGRES_PASSWORD
POSTGRES_DB
DB_HOST
DB_PORT
DB_USER
DB_PASSWORD
DB_NAME
JWT_SECRET
STORAGE_DIR
BACKUP_DIR
MIGRATIONS_DIR
PUBLIC_DIR
HTTP_ADDR
```

Seed password variables:

```text
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

## No External Runtime Dependency

The running service should not require:

- external identity provider
- external PDF API
- external database
- cloud storage
- remote notification provider
- internet access

## Rebuild

```bash
docker compose build --no-cache
```

## Stop

```bash
docker compose down
```

## Clean Local Re-test

```bash
docker compose down -v
```

This removes local volumes and allows seed data to be recreated on the next startup.

## Acceptance Checks

- Compose defines one service container.
- API health check works locally.
- Manual UI is served locally.
- PDF upload stores files in the storage volume.
- PostgreSQL metadata persists across container restart when volumes are kept.
- No external runtime service is required.
