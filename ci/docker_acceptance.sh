#!/usr/bin/env bash
set -euo pipefail

APP_SERVICE=${APP_SERVICE:-ironpage}
ACCEPTANCE_IMAGE=${ACCEPTANCE_IMAGE:-ironpage-vault-ci-acceptance}
env_dir=$(mktemp -d)
env_file="$env_dir/runtime.env"

random_hex() {
  od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
}

read_env_value() {
  local key="$1"
  sed -n "s/^${key}=//p" "$env_file" | tail -n 1
}

compose() {
  env \
    -u HOST_BIND_ADDRESS -u HOST_PORT -u HTTP_PORT -u HTTP_ADDR \
    -u POSTGRES_USER -u POSTGRES_PASSWORD -u POSTGRES_DB \
    -u DB_HOST -u DB_PORT -u DB_USER -u DB_PASSWORD -u DB_NAME \
    -u JWT_SECRET -u AES_KEY -u ACCEPTANCE_MODE \
    -u BOOTSTRAP_ADMIN_USERNAME -u BOOTSTRAP_ADMIN_PASSWORD \
    -u SEED_ADMIN_PASSWORD -u SEED_EDITOR_PASSWORD -u SEED_REVIEWER_PASSWORD \
    -u IRONPAGE_APP_ROOT -u MIGRATIONS_DIR -u PUBLIC_DIR \
    -u POSTGRES_VOLUME_ROOT -u PGDATA -u IRONPAGE_VOLUME_ROOT \
    -u STORAGE_DIR -u BACKUP_DIR \
    docker compose --env-file "$env_file" "$@"
}

cleanup() {
  if [ -s "$env_file" ]; then
    compose down -v --remove-orphans >/dev/null 2>&1 || true
  fi
  rm -rf "$env_dir"
}
trap cleanup EXIT

# Generate the complete build/runtime configuration rather than relying on
# application, Compose, or image defaults.
IRONPAGE_ENV_FILE="$env_file" IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
compose build "$APP_SERVICE"

# Prove normal-mode one-command deployment, initial Admin login, idempotent
# rerun, and restart against an independently generated clean configuration.
bash tests/api/test_bootstrap_restart_docker.sh

seed_admin_password=$(random_hex)
seed_editor_password=$(random_hex)
seed_reviewer_password=$(random_hex)
sed -i \
  -e '/^BOOTSTRAP_ADMIN_USERNAME=/d' \
  -e '/^BOOTSTRAP_ADMIN_PASSWORD=/d' \
  -e 's/^ACCEPTANCE_MODE=.*/ACCEPTANCE_MODE=true/' \
  -e "s/^SEED_ADMIN_PASSWORD=.*/SEED_ADMIN_PASSWORD=$seed_admin_password/" \
  -e "s/^SEED_EDITOR_PASSWORD=.*/SEED_EDITOR_PASSWORD=$seed_editor_password/" \
  -e "s/^SEED_REVIEWER_PASSWORD=.*/SEED_REVIEWER_PASSWORD=$seed_reviewer_password/" \
  "$env_file"
chmod 600 "$env_file"

compose down -v --remove-orphans >/dev/null 2>&1 || true
compose up -d "$APP_SERVICE"

host_address=$(read_env_value HOST_BIND_ADDRESS)
host_port=$(read_env_value HOST_PORT)
http_port=$(read_env_value HTTP_PORT)
host_base_url="http://${host_address}:${host_port}"
host_health_url="$host_base_url/healthz"
container_base_url="http://${APP_SERVICE}:${http_port}"

for i in $(seq 1 60); do
  if curl -fsS "$host_health_url" >/tmp/ironpage_health.out 2>&1; then
    break
  fi
  if [ "$i" = "60" ]; then
    echo "service did not become healthy"
    cat /tmp/ironpage_health.out || true
    compose logs --no-color || true
    exit 1
  fi
  sleep 1
done

BASE_URL="$host_base_url" \
SEED_ADMIN_PASSWORD="$seed_admin_password" \
  bash tests/api/test_auth_lockout_docker.sh

BASE_URL="$host_base_url" \
IRONPAGE_UI_EVIDENCE_DIR="${IRONPAGE_UI_EVIDENCE_DIR:-artifacts/regression/ui-interaction}" \
SEED_EDITOR_PASSWORD="$seed_editor_password" \
  bash tests/api/test_ui_interaction_acceptance.sh

container_id=$(compose ps -q "$APP_SERVICE")
if [ -z "$container_id" ]; then
  echo "compose service $APP_SERVICE is not running"
  compose logs --no-color || true
  exit 1
fi

network=$(docker inspect -f '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' "$container_id" | head -n1)
if [ -z "$network" ]; then
  echo "could not resolve compose network for $APP_SERVICE"
  compose logs --no-color || true
  exit 1
fi

docker build -f ci/Dockerfile.acceptance -t "$ACCEPTANCE_IMAGE" .

if ! docker run --rm --network "$network" \
  -e BASE_URL="$container_base_url" \
  -e SEED_ADMIN_PASSWORD="$seed_admin_password" \
  -e SEED_EDITOR_PASSWORD="$seed_editor_password" \
  -e SEED_REVIEWER_PASSWORD="$seed_reviewer_password" \
  "$ACCEPTANCE_IMAGE"; then
  echo "Docker acceptance failed; dumping compose logs"
  compose logs --no-color || true
  exit 1
fi
