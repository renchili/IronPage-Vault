# Usage and Acceptance Guide

This document covers secure startup, API use, and isolated acceptance runs. IronPage Vault is a pure backend API project. The browser UI is an acceptance-only backend testing aid.

## First normal-mode startup

From the repository root:

```bash
bash scripts/deploy.sh
```

A fresh checkout does not require manual `export` commands. The deployer creates a mode-`0600` `.env` file containing:

```text
DB_HOST
DB_PORT
DB_USER
DB_PASSWORD
DB_NAME
JWT_SECRET
AES_KEY
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

Database, JWT, AES, and bootstrap password values are generated locally. The script then builds and starts the one-container Compose service, waits for the health endpoint, and prints the initial Admin credentials once.

The generated `.env` is excluded from Git and the Docker build context. The application and image still contain no sensitive fallback values.

## Database configuration

The default generated database configuration is:

```text
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=ironpage
DB_NAME=ironpage
```

`DB_PASSWORD` is random. Compose uses `DB_USER`, `DB_PASSWORD`, and `DB_NAME` for both PostgreSQL initialization and the API connection, so the two sides cannot drift when using the supported configuration path.

Generate configuration without starting containers when a clean installation needs customization:

```bash
IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
```

After editing `.env`, deploy with the normal command:

```bash
bash scripts/deploy.sh
```

Do not change database credentials or `AES_KEY` casually after persistent data exists. Existing PostgreSQL credentials require an explicit migration, and encrypted metadata requires the original AES key.

## Initial Admin lifecycle

The generated bootstrap pair creates one initial Admin only when the user table is empty. After verifying the account, remove these lines from `.env`:

```text
BOOTSTRAP_ADMIN_USERNAME
BOOTSTRAP_ADMIN_PASSWORD
```

Existing users are not overwritten. Restarting without the pair continues to use the existing database identities.

## Subsequent normal-mode startup

Run the same command:

```bash
bash scripts/deploy.sh
```

The existing `.env` is reused. Repeated deployment does not rotate the database password, JWT signing material, AES key, or initial account password.

The API listens on:

```text
http://localhost:8080
```

Health check:

```bash
curl http://localhost:8080/healthz
```

A successful startup means runtime configuration passed validation, PostgreSQL is reachable, local data paths are available, and the health endpoint returns OK.

Normal mode does not create acceptance users and does not serve the backend test UI.

## Manual Compose operation

Once `.env` exists, maintainers may operate Compose directly:

```bash
docker compose --env-file .env up --build -d
docker compose --env-file .env down
```

The deployment script remains the supported first-run path because it creates and protects the required runtime configuration and verifies health.

## Explicit acceptance mode

An isolated acceptance run uses:

```text
ACCEPTANCE_MODE=true
SEED_ADMIN_PASSWORD
SEED_EDITOR_PASSWORD
SEED_REVIEWER_PASSWORD
```

All three fixture values must be supplied externally. The CI acceptance script generates execution-scoped values. Bootstrap Admin variables and acceptance fixture variables cannot be used together.

## Backend test UI

Only acceptance mode serves:

```text
http://localhost:8080/ui/
```

The page contains blank identity fields and no embedded credentials. Screenshot acceptance proves rendering only; it does not prove complete UI interaction or recovery behavior.

## API authentication flow

Use `POST /api/auth/login` with a local username and password. Authenticated requests use the returned bearer token together with a fresh request timestamp and unique request ID as documented in `docs/api-spec.md`.

Logout uses `POST /api/auth/logout`. Successful logout revokes the server-side session and token identifier. Authentication state persistence failures return the standard error envelope and do not return authenticated or logged-out success.

## Deployment validation

The one-command normal-mode path is exercised by:

```text
API_tests/test_bootstrap_restart_docker.sh
```

The test verifies generated configuration, initial Admin login, idempotent repeated deployment, restart without bootstrap values, persistent identity count, and normal-mode `/ui/` absence. It is executed by `ci/docker_acceptance.sh` as part of the full regression path.
