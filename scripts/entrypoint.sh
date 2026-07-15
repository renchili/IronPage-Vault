#!/usr/bin/env bash
set -euo pipefail

required_runtime_values=(
  POSTGRES_USER
  POSTGRES_PASSWORD
  POSTGRES_DB
  DB_USER
  DB_PASSWORD
  DB_NAME
  JWT_SECRET
  AES_KEY
)

for name in "${required_runtime_values[@]}"; do
  if [ -z "${!name:-}" ]; then
    echo "ERROR: $name is required" >&2
    exit 1
  fi
done

if [ "$POSTGRES_USER" != "$DB_USER" ]; then
  echo "ERROR: POSTGRES_USER and DB_USER must match in the single-container deployment" >&2
  exit 1
fi

if [ "$POSTGRES_PASSWORD" != "$DB_PASSWORD" ]; then
  echo "ERROR: POSTGRES_PASSWORD and DB_PASSWORD must match in the single-container deployment" >&2
  exit 1
fi

if [ "$POSTGRES_DB" != "$DB_NAME" ]; then
  echo "ERROR: POSTGRES_DB and DB_NAME must match in the single-container deployment" >&2
  exit 1
fi

case "${ACCEPTANCE_MODE:-false}" in
  1|true|TRUE|True)
    for name in \
      SEED_ADMIN_PASSWORD \
      SEED_EDITOR_PASSWORD \
      SEED_REVIEWER_PASSWORD
    do
      if [ -z "${!name:-}" ]; then
        echo "ERROR: $name is required when ACCEPTANCE_MODE is enabled" >&2
        exit 1
      fi
    done
    ;;
esac

export PGDATA="${PGDATA:-/var/lib/postgresql/data}"

docker-entrypoint.sh postgres &
postgres_pid=$!

cleanup() {
  if kill -0 "$postgres_pid" >/dev/null 2>&1; then
    kill "$postgres_pid" >/dev/null 2>&1 || true
    wait "$postgres_pid" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT
trap 'exit 143' TERM
trap 'exit 130' INT

ready=0
for _ in $(seq 1 60); do
  if pg_isready \
      -h 127.0.0.1 \
      -p 5432 \
      -U "$POSTGRES_USER" \
      -d "$POSTGRES_DB" >/dev/null 2>&1
  then
    ready=1
    break
  fi

  if ! kill -0 "$postgres_pid" >/dev/null 2>&1; then
    echo "ERROR: PostgreSQL exited before becoming ready" >&2
    wait "$postgres_pid"
    exit 1
  fi

  sleep 1
done

if [ "$ready" -ne 1 ]; then
  echo "ERROR: PostgreSQL did not become ready within 60 seconds" >&2
  exit 1
fi

/usr/local/bin/ironpage
