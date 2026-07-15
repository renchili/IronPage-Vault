#!/usr/bin/env bash
set -euo pipefail

APP_SERVICE=${APP_SERVICE:-ironpage}
ACCEPTANCE_IMAGE=${ACCEPTANCE_IMAGE:-ironpage-vault-ci-acceptance}
HOST_HEALTH_URL=${HOST_HEALTH_URL:-http://localhost:8080/healthz}
CONTAINER_BASE_URL=${CONTAINER_BASE_URL:-http://ironpage:8080}

random_hex() {
  od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
}

# Acceptance identities and runtime secrets are generated for this execution.
# They are not application defaults and are not persisted in the repository.
export DB_PASSWORD=${DB_PASSWORD:-$(random_hex)}
export JWT_SECRET=${JWT_SECRET:-$(random_hex)}
export AES_KEY=${AES_KEY:-$(random_hex)}
export ACCEPTANCE_MODE=true
export SEED_ADMIN_PASSWORD=${SEED_ADMIN_PASSWORD:-$(random_hex)}
export SEED_EDITOR_PASSWORD=${SEED_EDITOR_PASSWORD:-$(random_hex)}
export SEED_REVIEWER_PASSWORD=${SEED_REVIEWER_PASSWORD:-$(random_hex)}

docker compose build "$APP_SERVICE"
docker compose up -d "$APP_SERVICE"

cleanup() {
  docker compose down -v
}
trap cleanup EXIT

for i in $(seq 1 60); do
  if curl -s "$HOST_HEALTH_URL" >/tmp/ironpage_health.out 2>&1; then
    break
  fi
  sleep 1
  if [ "$i" = "60" ]; then
    echo "service did not become healthy"
    cat /tmp/ironpage_health.out || true
    docker compose logs --no-color || true
    exit 1
  fi
done

container_id="$(docker compose ps -q "$APP_SERVICE")"
if [ -z "$container_id" ]; then
  echo "compose service $APP_SERVICE is not running"
  docker compose logs --no-color || true
  exit 1
fi

network="$(docker inspect -f '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' "$container_id" | head -n1)"
if [ -z "$network" ]; then
  echo "could not resolve compose network for $APP_SERVICE"
  docker compose logs --no-color || true
  exit 1
fi

docker build -f ci/Dockerfile.acceptance -t "$ACCEPTANCE_IMAGE" .

if ! docker run --rm --network "$network" \
  -e BASE_URL="$CONTAINER_BASE_URL" \
  -e SEED_ADMIN_PASSWORD="$SEED_ADMIN_PASSWORD" \
  -e SEED_EDITOR_PASSWORD="$SEED_EDITOR_PASSWORD" \
  -e SEED_REVIEWER_PASSWORD="$SEED_REVIEWER_PASSWORD" \
  "$ACCEPTANCE_IMAGE"; then
  echo "Docker acceptance failed; dumping compose logs"
  docker compose logs --no-color || true
  exit 1
fi
