#!/usr/bin/env bash
set -euo pipefail

APP_SERVICE=${APP_SERVICE:-ironpage}
BASE_URL=${BASE_URL:-http://localhost:8080}
BODY=${BODY:-/tmp/ironpage_bootstrap_body.json}

random_hex() {
  od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
}

wait_for_health() {
  local attempt
  for attempt in $(seq 1 60); do
    if curl -fsS "$BASE_URL/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "FAIL bootstrap: service did not become healthy" >&2
  docker compose logs --no-color "$APP_SERVICE" || true
  return 1
}

login_code() {
  local username="$1" password="$2"
  curl -sS -o "$BODY" -w '%{http_code}' "$BASE_URL/api/auth/login" \
    -H 'Content-Type: application/json' \
    -H "X-Request-ID: bootstrap-$(date +%s%N)-$RANDOM" \
    -d "{\"username\":\"$username\",\"password\":\"$password\"}"
}

expect_code() {
  local name="$1" expected="$2" actual="$3"
  if [ "$actual" != "$expected" ]; then
    echo "FAIL bootstrap: $name expected=$expected actual=$actual" >&2
    [ -s "$BODY" ] && cat "$BODY" >&2 && echo >&2
    return 1
  fi
  echo "PASS bootstrap: $name"
}

bootstrap_username="bootstrap_$(random_hex | cut -c1-12)"
bootstrap_password="$(random_hex)$(random_hex)"

export DB_PASSWORD="$(random_hex)"
export JWT_SECRET="$(random_hex)"
export AES_KEY="$(random_hex)"
export ACCEPTANCE_MODE=false
export BOOTSTRAP_ADMIN_USERNAME="$bootstrap_username"
export BOOTSTRAP_ADMIN_PASSWORD="$bootstrap_password"
unset SEED_ADMIN_PASSWORD SEED_EDITOR_PASSWORD SEED_REVIEWER_PASSWORD

cleanup() {
  docker compose down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

# The caller builds the image once. This flow exercises a clean normal-mode
# volume, then recreates the container against the same persisted database.
docker compose down -v --remove-orphans >/dev/null 2>&1 || true
docker compose up -d "$APP_SERVICE"
wait_for_health

code=$(login_code "$bootstrap_username" "$bootstrap_password")
expect_code "initial Admin login after empty-volume bootstrap" 200 "$code"

ui_code=$(curl -sS -o /tmp/ironpage_bootstrap_ui.out -w '%{http_code}' "$BASE_URL/ui/")
expect_code "normal mode does not serve acceptance UI" 404 "$ui_code"

# Remove the bootstrap pair and recreate the container while retaining volumes.
docker compose down --remove-orphans
unset BOOTSTRAP_ADMIN_USERNAME BOOTSTRAP_ADMIN_PASSWORD
docker compose up -d "$APP_SERVICE"
wait_for_health

code=$(login_code "$bootstrap_username" "$bootstrap_password")
expect_code "existing Admin login survives restart without bootstrap values" 200 "$code"

admin_count=$(docker compose exec -T "$APP_SERVICE" bash -lc '
  psql -X -qAt -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT COUNT(*) FROM users WHERE role='"'"'Admin'"'"';"
' | tr -d '\r')
if [ "$admin_count" != "1" ]; then
  echo "FAIL bootstrap: expected one Admin after restart, got $admin_count" >&2
  exit 1
fi
echo "PASS bootstrap: restart did not create or overwrite another Admin"

ui_code=$(curl -sS -o /tmp/ironpage_bootstrap_ui_after_restart.out -w '%{http_code}' "$BASE_URL/ui/")
expect_code "normal mode remains without acceptance UI after restart" 404 "$ui_code"

echo "PASS bootstrap-suite: normal-mode first start and restart"
