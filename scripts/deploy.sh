#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
ENV_FILE=${IRONPAGE_ENV_FILE:-"$ROOT_DIR/.env"}
COMPOSE_FILE=${IRONPAGE_COMPOSE_FILE:-"$ROOT_DIR/docker-compose.yml"}
APP_SERVICE=${APP_SERVICE:-ironpage}
DEPLOY_BUILD=${IRONPAGE_DEPLOY_BUILD:-true}
STARTUP_WAIT_SECONDS=${IRONPAGE_STARTUP_WAIT_SECONDS:-120}
DRY_RUN=${IRONPAGE_DEPLOY_DRY_RUN:-false}

random_hex() {
  od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
}

random_port() {
  local value
  value=$(od -An -N2 -tu2 /dev/urandom | tr -d ' ')
  printf '%s\n' "$((20000 + value % 30000))"
}

read_env_value() {
  local key="$1"
  sed -n "s/^${key}=//p" "$ENV_FILE" | tail -n 1
}

validate_runtime_env() {
  local name value
  for name in \
    HOST_BIND_ADDRESS HOST_PORT HTTP_PORT HTTP_ADDR \
    DB_PORT DB_USER DB_PASSWORD DB_NAME \
    JWT_SECRET AES_KEY ACCEPTANCE_MODE \
    IRONPAGE_APP_ROOT MIGRATIONS_DIR PUBLIC_DIR \
    POSTGRES_VOLUME_ROOT PGDATA IRONPAGE_VOLUME_ROOT STORAGE_DIR BACKUP_DIR
  do
    value=$(read_env_value "$name")
    if [ -z "$value" ]; then
      echo "ERROR: runtime configuration is missing $name: $ENV_FILE" >&2
      exit 1
    fi
  done
}

create_runtime_env() {
  local installation_id db_password jwt_secret aes_key
  local admin_username admin_password db_port http_port host_port
  local app_root postgres_root data_root

  installation_id=$(random_hex | cut -c1-12)
  db_password=$(random_hex)
  jwt_secret=$(random_hex)
  aes_key=$(random_hex)
  admin_username="admin_$installation_id"
  # bcrypt accepts at most 72 bytes; one random_hex value is 64 ASCII bytes.
  admin_password=$(random_hex)
  db_port=$(random_port)
  http_port=$(random_port)
  while [ "$http_port" = "$db_port" ]; do
    http_port=$(random_port)
  done
  host_port=$(random_port)
  app_root="/opt/ironpage-$installation_id"
  postgres_root="/var/lib/postgresql-$installation_id"
  data_root="/var/lib/ironpage-$installation_id"

  mkdir -p "$(dirname -- "$ENV_FILE")"
  umask 077
  cat >"$ENV_FILE" <<EOF
HOST_BIND_ADDRESS=127.0.0.1
HOST_PORT=$host_port
HTTP_PORT=$http_port
HTTP_ADDR=0.0.0.0:$http_port
DB_PORT=$db_port
DB_USER=ironpage_$installation_id
DB_PASSWORD=$db_password
DB_NAME=ironpage_$installation_id
JWT_SECRET=$jwt_secret
AES_KEY=$aes_key
ACCEPTANCE_MODE=false
BOOTSTRAP_ADMIN_USERNAME=$admin_username
BOOTSTRAP_ADMIN_PASSWORD=$admin_password
SEED_ADMIN_PASSWORD=
SEED_EDITOR_PASSWORD=
SEED_REVIEWER_PASSWORD=
IRONPAGE_APP_ROOT=$app_root
MIGRATIONS_DIR=$app_root/migrations
PUBLIC_DIR=$app_root/public
POSTGRES_VOLUME_ROOT=$postgres_root
PGDATA=$postgres_root/data
IRONPAGE_VOLUME_ROOT=$data_root
STORAGE_DIR=$data_root/storage
BACKUP_DIR=$data_root/backups
EOF
  chmod 600 "$ENV_FILE"
}

if [ -e "$ENV_FILE" ] && [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: runtime environment path is not a regular file: $ENV_FILE" >&2
  exit 1
fi

created_runtime_env=false
if [ ! -s "$ENV_FILE" ]; then
  create_runtime_env
  created_runtime_env=true
else
  chmod 600 "$ENV_FILE"
fi
validate_runtime_env

if [ "$DRY_RUN" = "true" ]; then
  echo "Runtime configuration ready: $ENV_FILE"
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "ERROR: docker is required" >&2
  exit 1
fi
if ! docker compose version >/dev/null 2>&1; then
  echo "ERROR: Docker Compose v2 is required" >&2
  exit 1
fi

# Values in the generated runtime file are the deployment source of truth.
# Unsetting matching ambient variables prevents an old shell export from silently
# overriding the persisted configuration on restart.
compose() {
  env \
    -u HOST_BIND_ADDRESS \
    -u HOST_PORT \
    -u HTTP_PORT \
    -u HTTP_ADDR \
    -u POSTGRES_USER \
    -u POSTGRES_PASSWORD \
    -u POSTGRES_DB \
    -u DB_HOST \
    -u DB_PORT \
    -u DB_USER \
    -u DB_PASSWORD \
    -u DB_NAME \
    -u JWT_SECRET \
    -u AES_KEY \
    -u ACCEPTANCE_MODE \
    -u BOOTSTRAP_ADMIN_USERNAME \
    -u BOOTSTRAP_ADMIN_PASSWORD \
    -u SEED_ADMIN_PASSWORD \
    -u SEED_EDITOR_PASSWORD \
    -u SEED_REVIEWER_PASSWORD \
    -u IRONPAGE_APP_ROOT \
    -u MIGRATIONS_DIR \
    -u PUBLIC_DIR \
    -u POSTGRES_VOLUME_ROOT \
    -u PGDATA \
    -u IRONPAGE_VOLUME_ROOT \
    -u STORAGE_DIR \
    -u BACKUP_DIR \
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

cd "$ROOT_DIR"
compose config >/dev/null

case "$DEPLOY_BUILD" in
  true|1|yes)
    compose up --build -d "$APP_SERVICE"
    ;;
  false|0|no)
    compose up -d "$APP_SERVICE"
    ;;
  *)
    echo "ERROR: IRONPAGE_DEPLOY_BUILD must be true or false" >&2
    exit 1
    ;;
esac

http_port=$(read_env_value HTTP_PORT)
ready=false
for _ in $(seq 1 "$STARTUP_WAIT_SECONDS"); do
  if compose exec -T "$APP_SERVICE" python3 -c \
    "import urllib.request; urllib.request.urlopen('http://127.0.0.1:${http_port}/healthz', timeout=2).read()" \
    >/dev/null 2>&1; then
    ready=true
    break
  fi
  sleep 1
done

if [ "$ready" != "true" ]; then
  echo "ERROR: IronPage Vault did not become healthy within ${STARTUP_WAIT_SECONDS} seconds" >&2
  compose logs --no-color "$APP_SERVICE" >&2 || true
  exit 1
fi

host_address=$(read_env_value HOST_BIND_ADDRESS)
host_port=$(read_env_value HOST_PORT)
base_url="http://${host_address}:${host_port}"
echo "IronPage Vault is running."
echo "API: $base_url"
echo "Health: $base_url/healthz"
echo "Swagger: $base_url/swagger/index.html"
echo "Runtime configuration: $ENV_FILE"

if [ "$created_runtime_env" = "true" ]; then
  echo "Initial administrator username: $(read_env_value BOOTSTRAP_ADMIN_USERNAME)"
  echo "Initial administrator password: $(read_env_value BOOTSTRAP_ADMIN_PASSWORD)"
  echo "Store these credentials securely, then remove BOOTSTRAP_ADMIN_USERNAME and BOOTSTRAP_ADMIN_PASSWORD from the runtime file after the account is verified."
fi
