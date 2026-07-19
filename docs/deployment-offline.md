# Offline Deployment

IronPage Vault runs as one local Compose service for an air-gapped environment. The service image contains PostgreSQL, the Go API, migrations, PDF processing dependencies, local storage support, and the acceptance-only browser asset.

## First deployment

From the repository root:

```bash
bash scripts/deploy.sh
```

The first run creates `.env` with mode `0600`, generates an installation identifier, writes every local runtime setting, builds the image, starts the service, waits for health, and prints the actual API URL plus the initial Admin credentials.

No database identity, port, filesystem path, credential, signing key, encryption key, container name, or host port must be exported before this command. `.env` is excluded from Git and from the Docker build context.

Running the same command again reuses the existing `.env`:

```bash
bash scripts/deploy.sh
```

This preserves the PostgreSQL identity and password, JWT signing material, AES encryption material, generated ports, paths, and persistent-volume targets for the installation.

## Find the service URL

The deployer prints the actual values after a successful start. The same values are stored in `.env`:

```text
HOST_BIND_ADDRESS
HOST_PORT
```

The health and Swagger URLs are:

```text
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/healthz
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/swagger/index.html
```

Do not assume port `8080`; the host and container ports are installation configuration.

## Generated runtime configuration

The generated file contains:

| Area | Variables | Behavior |
|---|---|---|
| Host exposure | `HOST_BIND_ADDRESS`, `HOST_PORT` | Local address and host port used by Compose |
| API listener | `HTTP_PORT`, `HTTP_ADDR` | Container listener port and Go bind address |
| PostgreSQL | `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Installation-specific embedded database identity |
| Cryptography | `JWT_SECRET`, `AES_KEY` | Random signing and encrypted-column material |
| Bootstrap | `BOOTSTRAP_ADMIN_USERNAME`, `BOOTSTRAP_ADMIN_PASSWORD` | Initial empty-database Admin pair |
| Image layout | `IRONPAGE_APP_ROOT`, `MIGRATIONS_DIR`, `PUBLIC_DIR` | Installation-specific application asset paths |
| PostgreSQL data | `POSTGRES_VOLUME_ROOT`, `PGDATA` | Generated persistent PostgreSQL target |
| Product data | `IRONPAGE_VOLUME_ROOT`, `STORAGE_DIR`, `BACKUP_DIR` | Generated PDF and backup targets |
| Mode | `ACCEPTANCE_MODE`, `SEED_*_PASSWORD` | Normal mode by default; no fixture credentials |

`docker-compose.yml` maps PostgreSQL initialization from the same generated database values used by the API:

```text
POSTGRES_USER     <- DB_USER
POSTGRES_PASSWORD <- DB_PASSWORD
POSTGRES_DB       <- DB_NAME
```

The image, Compose file, and Go application do not supply alternative local defaults. An incomplete runtime file fails validation.

## Customize a clean installation

Generate the configuration without building or starting containers:

```bash
IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
```

Edit `.env` before the first build, then deploy:

```bash
bash scripts/deploy.sh
```

The values are installation state. Do not change database identity, database password, AES key, image paths, volume targets, or listener ports after persistent data exists unless the associated PostgreSQL, encryption, image, and storage migration is performed deliberately.

## Initial Admin

The generated bootstrap pair is used only while the user table is empty. After the initial Admin login is verified, remove both entries from `.env`:

```text
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

Existing users remain unchanged across restart.

## Manual Compose operation

After `.env` exists:

```bash
docker compose --env-file .env up --build -d
docker compose --env-file .env down
```

For a disposable clean reset:

```bash
docker compose --env-file .env down -v
rm .env
bash scripts/deploy.sh
```

Do not remove `.env` while retaining real persistent data unless its database and encryption values have been backed up securely.

## Acceptance mode

Acceptance mode is only for isolated validation. It requires externally generated values for all fixture identities and cannot be combined with bootstrap values:

```text
ACCEPTANCE_MODE=true
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

Only acceptance mode serves the canonical browser probe at:

```text
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/ui/
```

The sole deployed acceptance HTML file is `public/index.html`. It contains no credential.

## Evidence boundary

Static source inspection can establish configuration ownership and path consistency. Runtime startup, login, restart, PDF, RBAC, backup/restore, browser interaction, and complete-regression claims require an executed artifact tied to the exact revision. The supported complete command is:

```bash
bash ci/run_full_regression.sh artifacts/regression
```
