#!/usr/bin/env bash
set -euo pipefail

APP_SERVICE=${APP_SERVICE:-ironpage}
ACCEPTANCE_IMAGE=${ACCEPTANCE_IMAGE:-ironpage-vault-ci-acceptance}
HOST_HEALTH_URL=${HOST_HEALTH_URL:-http://localhost:8080/healthz}
CONTAINER_BASE_URL=${CONTAINER_BASE_URL:-http://ironpage:8080}

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

if ! docker run --rm --network "$network" -e BASE_URL="$CONTAINER_BASE_URL" "$ACCEPTANCE_IMAGE"; then
  echo "Docker acceptance failed; dumping compose logs"
  docker compose logs --no-color || true
  exit 1
fi
