#!/usr/bin/env bash
set -euo pipefail

# Verifies the local/manual test entrypoint remains runnable from a fresh
# checkout with generated Swagger artifacts absent. The probe intentionally
# runs only the lightweight contract path in run_tests.sh, not the full local
# API acceptance suite.

rm -rf docs/swagger
IRONPAGE_RUN_TESTS_CONTRACT_PROBE=1 bash run_tests.sh

test -s docs/swagger/docs.go
test -s docs/swagger/swagger.yaml

echo "PASS run_tests local entrypoint contract"
