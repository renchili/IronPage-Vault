#!/usr/bin/env bash
set -euo pipefail

# Verifies the local/manual test entrypoint remains runnable from a fresh
# checkout with generated Swagger artifacts absent. The probe intentionally
# runs only the lightweight contract path in run_tests.sh, not the full local
# API acceptance suite.

rm -rf docs/swagger artifacts/local-acceptance
IRONPAGE_RUN_TESTS_CONTRACT_PROBE=1 bash run_tests.sh

test -s docs/swagger/docs.go
test -s docs/swagger/swagger.yaml
test -s artifacts/local-acceptance/results.tsv
test -s artifacts/local-acceptance/summary.json
test -s artifacts/local-acceptance/summary.md
test -s artifacts/local-acceptance/report.html
grep -q 'IronPage Local Acceptance Report' artifacts/local-acceptance/report.html
grep -q 'local_entrypoint_contract' artifacts/local-acceptance/results.tsv

echo "PASS run_tests local entrypoint contract"
