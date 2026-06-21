#!/usr/bin/env bash
set -euo pipefail

# Pre-merge CI control-plane contract check.
# This executes the regression runner in a deliberate failing mode and verifies
# the CI contract: a failed stage must still produce results.tsv, summary.json,
# and summary.md, while the runner exits non-zero. It also verifies that failed
# stage log content is printed to stdout.

probe_dir="$(mktemp -d)"
trap 'rm -rf "$probe_dir"' EXIT

set +e
IRONPAGE_REGRESSION_CONTRACT_PROBE=1 bash ci/run_full_regression.sh "$probe_dir" >"$probe_dir/probe.log" 2>&1
status=$?
set -e

if [ "$status" -eq 0 ]; then
  echo "FAIL regression contract: failing probe unexpectedly exited 0"
  cat "$probe_dir/probe.log"
  exit 1
fi

test -s "$probe_dir/results.tsv"
test -s "$probe_dir/summary.json"
test -s "$probe_dir/summary.md"
test -s "$probe_dir/logs/contract_fail.log"

grep -q 'IRONPAGE_REGRESSION_CONTRACT_FAIL_SENTINEL' "$probe_dir/logs/contract_fail.log"
grep -q 'FAIL contract_fail' "$probe_dir/probe.log"
grep -q -- '---- contract_fail failure log:' "$probe_dir/probe.log"
grep -q 'IRONPAGE_REGRESSION_CONTRACT_FAIL_SENTINEL' "$probe_dir/probe.log"
grep -q -- '---- end contract_fail failure log ----' "$probe_dir/probe.log"

python3 - "$probe_dir/summary.json" <<'PY'
import json, sys
summary_path = sys.argv[1]
with open(summary_path, encoding='utf-8') as f:
    summary = json.load(f)
if summary.get('overall_status') != 'failed':
    raise SystemExit(f"expected overall_status=failed, got {summary.get('overall_status')!r}")
stages = {stage['stage']: stage for stage in summary.get('stages', [])}
if stages.get('contract_pass', {}).get('status') != 0:
    raise SystemExit('contract_pass stage missing or not successful')
if stages.get('contract_fail', {}).get('status') == 0:
    raise SystemExit('contract_fail stage missing or unexpectedly successful')
PY

grep -q 'Overall: \*\*FAILED\*\*' "$probe_dir/summary.md"

docker build -f ci/Dockerfile.acceptance -t ironpage-vault-ci-acceptance-contract .
docker run --rm --entrypoint bash ironpage-vault-ci-acceptance-contract -lc '
  test -f internal/service/pdf.go
  test -f internal/platform/pdf_strict.go
  test -f internal/platform/backup_strict.go
  bash API_tests/test_strict_dependency_failures.sh
'

echo "PASS regression flow contract"
