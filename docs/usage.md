# Usage and Acceptance Guide

IronPage Vault is a backend API. The browser surface is an acceptance-only backend probe, not a product frontend.

## Start a normal installation

From the repository root:

```bash
bash scripts/deploy.sh
```

A fresh checkout needs no manual exports. The deployer creates a mode-`0600` `.env` containing the complete installation configuration:

```text
HOST_BIND_ADDRESS
HOST_PORT
HTTP_PORT
HTTP_ADDR
DB_PORT
DB_USER
DB_PASSWORD
DB_NAME
JWT_SECRET
AES_KEY
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
IRONPAGE_APP_ROOT
MIGRATIONS_DIR
PUBLIC_DIR
POSTGRES_VOLUME_ROOT
PGDATA
IRONPAGE_VOLUME_ROOT
STORAGE_DIR
BACKUP_DIR
```

Ports, database identity, credentials, cryptographic material, application paths, and persistent-data targets are generated for the installation. The script builds and starts the single Compose service, waits for health, and prints the actual URLs and initial Admin pair.

The runtime file is excluded from Git and the Docker build context. The image, Compose file, and application do not replace missing values with fixed local configuration.

## Use the generated URL

Read the URL printed by `scripts/deploy.sh` or inspect:

```text
HOST_BIND_ADDRESS
HOST_PORT
```

API requests use:

```text
http://<HOST_BIND_ADDRESS>:<HOST_PORT>
```

For example:

```bash
set -a
. ./.env
set +a
curl "http://${HOST_BIND_ADDRESS}:${HOST_PORT}/healthz"
```

Do not assume a particular port.

## Customize before first start

```bash
IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
# Edit .env before the first build.
bash scripts/deploy.sh
```

Database identity, database password, AES key, image paths, persistent-volume targets, and ports are installation state. Changing them after data exists requires the corresponding database, encryption, image, or storage migration.

## Initial Admin lifecycle

The bootstrap pair creates one initial Admin only when the user table is empty. After verifying login, remove:

```text
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

Existing users are not overwritten. Restarting without the pair continues to use the same generated installation configuration.

## Subsequent startup

```bash
bash scripts/deploy.sh
```

The existing `.env` is reused. The deployment does not rotate database credentials, JWT material, AES material, ports, paths, or persistent targets.

## Manual Compose operation

After `.env` exists:

```bash
docker compose --env-file .env up --build -d
docker compose --env-file .env down
```

The deployment script remains the supported first-run path because it creates, protects, validates, and reports the installation configuration.

## Acceptance mode

An isolated acceptance environment requires:

```text
ACCEPTANCE_MODE=true
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

All fixture values are execution-scoped. Bootstrap values and acceptance fixture values cannot be combined.

Only acceptance mode serves the canonical probe:

```text
http://<HOST_BIND_ADDRESS>:<HOST_PORT>/ui/
```

The only served acceptance HTML source is `public/index.html`; it contains no credential.

## API authentication

Use `POST /api/auth/login` with a local username and password. Authenticated requests include the returned bearer token, a fresh request timestamp, and a unique request ID as documented in `docs/api-spec.md`.

`POST /api/auth/logout` revokes server-side session state and the token identifier. Authentication-state persistence failures use the standard error envelope and fail closed.

## Validation paths

The normal generated deployment and restart flow is defined in:

```text
tests/api/test_bootstrap_restart_docker.sh
```

Complete Docker/API/browser regression is orchestrated by:

```bash
bash ci/run_full_regression.sh artifacts/regression
```

A static source review or lightweight report probe is not equivalent to executed complete-regression evidence.
