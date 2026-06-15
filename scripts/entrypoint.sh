#!/usr/bin/env bash
set -euo pipefail

export PGDATA=${PGDATA:-/var/lib/postgresql/data}

docker-entrypoint.sh postgres &

until pg_isready -h 127.0.0.1 -p 5432 -U "${POSTGRES_USER:-ironpage}" >/dev/null 2>&1; do
  sleep 1
done

/usr/local/bin/ironpage
