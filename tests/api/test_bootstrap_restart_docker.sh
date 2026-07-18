#!/usr/bin/env bash
set -euo pipefail

APP_SERVICE=${APP_SERVICE:-ironpage}
BODY=${BODY:-/tmp/ironpage_bootstrap_body.json}
env_dir=$(mktemp -d)
env_file="$env_dir/runtime.env"

# The normal-mode test relies only on its generated runtime file.
unset HOST_BIND_ADDRESS HOST_PORT HTTP_PORT HTTP_ADDR
unset DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME
unset JWT_SECRET AES_KEY ACCEPTANCE_MODE
unset BOOTSTRAP_ADMIN_USERNAME BOOTSTRAP_ADMIN_PASSWORD
unset SEED_ADMIN_PASSWORD SEED_EDITOR_PASSWORD SEED_REVIEWER_PASSWORD
unset IRONPAGE_APP_ROOT MIGRATIONS_DIR PUBLIC_DIR
unset POSTGRES_VOLUME_ROOT PGDATA IRONPAGE_VOLUME_ROOT STORAGE_DIR BACKUP_DIR

compose() {
  docker compose --env-file "$env_file" "$@"
}

read_env_value() {
  local key="$1"
  sed -n "s/^${key}=//p" "$env_file" | tail -n 1
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

cleanup() {
  if [ -s "$env_file" ]; then
    compose down -v --remove-orphans >/dev/null 2>&1 || true
  fi
  rm -rf "$env_dir"
}
trap cleanup EXIT

IRONPAGE_ENV_FILE="$env_file" IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
compose down -v --remove-orphans >/dev/null 2>&1 || true

bootstrap_username=$(read_env_value BOOTSTRAP_ADMIN_USERNAME)
bootstrap_password=$(read_env_value BOOTSTRAP_ADMIN_PASSWORD)
host_address=$(read_env_value HOST_BIND_ADDRESS)
host_port=$(read_env_value HOST_PORT)
BASE_URL="http://${host_address}:${host_port}"
if [ -z "$bootstrap_username" ] || [ -z "$bootstrap_password" ]; then
  echo "FAIL bootstrap: generated runtime file has no initial Admin credentials" >&2
  exit 1
fi

# The image build consumes installation-specific asset paths and HTTP port, so
# the first start builds from this same generated runtime file.
IRONPAGE_ENV_FILE="$env_file" IRONPAGE_DEPLOY_BUILD=true bash scripts/deploy.sh

code=$(login_code "$bootstrap_username" "$bootstrap_password")
expect_code "initial Admin login after one-command deployment" 200 "$code"

ui_code=$(curl -sS -o /tmp/ironpage_bootstrap_ui.out -w '%{http_code}' "$BASE_URL/ui/")
expect_code "normal mode does not serve acceptance UI" 404 "$ui_code"

cp "$env_file" "$env_dir/runtime.before"
IRONPAGE_ENV_FILE="$env_file" IRONPAGE_DEPLOY_BUILD=false bash scripts/deploy.sh
cmp "$env_dir/runtime.before" "$env_file"
echo "PASS bootstrap: repeated deployment reused persisted runtime configuration"

compose down --remove-orphans
python3 - "$env_file" <<'PY'
from pathlib import Path
import sys
path = Path(sys.argv[1])
lines = [
    line for line in path.read_text(encoding="utf-8").splitlines()
    if not line.startswith("BOOTSTRAP_ADMIN_USERNAME=")
    and not line.startswith("BOOTSTRAP_ADMIN_PASSWORD=")
]
path.write_text("\n".join(lines) + "\n", encoding="utf-8")
path.chmod(0o600)
PY
IRONPAGE_ENV_FILE="$env_file" IRONPAGE_DEPLOY_BUILD=false bash scripts/deploy.sh

code=$(login_code "$bootstrap_username" "$bootstrap_password")
expect_code "existing Admin login survives restart without bootstrap values" 200 "$code"

admin_count=$(compose exec -T "$APP_SERVICE" bash -lc '
  psql -X -qAt -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT COUNT(*) FROM users WHERE role='"'"'Admin'"'"';"
' | tr -d '\r')
if [ "$admin_count" != "1" ]; then
  echo "FAIL bootstrap: expected one Admin after restart, got $admin_count" >&2
  exit 1
fi
echo "PASS bootstrap: restart did not create or overwrite another Admin"

ui_code=$(curl -sS -o /tmp/ironpage_bootstrap_ui_after_restart.out -w '%{http_code}' "$BASE_URL/ui/")
expect_code "normal mode remains without acceptance UI after restart" 404 "$ui_code"

echo "PASS bootstrap-suite: one-command normal-mode first start and restart"
