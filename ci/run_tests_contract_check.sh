#!/usr/bin/env bash
set -euo pipefail

# This probe verifies entrypoint/report generation only. It must not be reported
# as full local or full-regression evidence.
rm -rf docs/swagger artifacts/local-acceptance
IRONPAGE_RUN_TESTS_CONTRACT_PROBE=1 bash run_tests.sh

test -s docs/swagger/docs.go
test -s docs/swagger/swagger.yaml
test -s artifacts/local-acceptance/results.tsv
test -s artifacts/local-acceptance/summary.json
test -s artifacts/local-acceptance/summary.md
test -s artifacts/local-acceptance/report.html
grep -q 'IronPage Local Acceptance Report' artifacts/local-acceptance/report.html
grep -q 'claims coverage only for the stage rows' artifacts/local-acceptance/report.html

python3 - <<'PY'
import json
from pathlib import Path
payload = json.loads(Path('artifacts/local-acceptance/summary.json').read_text(encoding='utf-8'))
expected = ['prepare_swagger', 'local_entrypoint_contract']
if payload.get('executed_stages') != expected:
    raise SystemExit(f"local probe overstated or omitted stages: {payload.get('executed_stages')!r}")
if [stage.get('stage') for stage in payload.get('stages', [])] != expected:
    raise SystemExit('stage rows do not match executed_stages')
PY

if grep -q 'redaction/Bates/compare flows' artifacts/local-acceptance/report.html; then
  echo "ERROR: probe report claims unexecuted domain coverage" >&2
  exit 1
fi

echo "PASS run_tests local entrypoint contract"
