#!/usr/bin/env bash
set -euo pipefail

# This script is intentionally Docker-based. It does not require a local Go toolchain.

docker compose build
docker compose up -d

cleanup() {
  docker compose down
}
trap cleanup EXIT

for i in $(seq 1 60); do
  if curl -s http://localhost:8080/healthz >/tmp/ironpage_health.out 2>&1; then
    break
  fi
  sleep 1
  if [ "$i" = "60" ]; then
    echo "service did not become healthy"
    cat /tmp/ironpage_health.out || true
    exit 1
  fi
done

if ! bash ./run_tests.sh || ! bash API_tests/test_request_guard_edges.sh; then
  echo "Docker acceptance failed; dumping compose logs"
  docker compose logs --no-color || true
  exit 1
fi
