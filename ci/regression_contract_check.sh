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

# The Docker acceptance runner must generate execution-scoped runtime values,
# explicitly enable acceptance mode, and pass fixture values into the test
# container instead of depending on application defaults.
grep -q 'random_hex' ci/docker_acceptance.sh
grep -q 'export ACCEPTANCE_MODE=true' ci/docker_acceptance.sh
grep -q 'export DB_PASSWORD=' ci/docker_acceptance.sh
grep -q 'export JWT_SECRET=' ci/docker_acceptance.sh
grep -q 'export AES_KEY=' ci/docker_acceptance.sh
grep -q -- '-e SEED_ADMIN_PASSWORD=' ci/docker_acceptance.sh
grep -q -- '-e SEED_EDITOR_PASSWORD=' ci/docker_acceptance.sh
grep -q -- '-e SEED_REVIEWER_PASSWORD=' ci/docker_acceptance.sh

# Authentication acceptance must include a persisted rolling window, real
# normal-mode restart evidence, fail-closed fault injection, and browser flow.
test -f migrations/002_login_attempt_window.sql
test -f API_tests/test_auth_lockout_docker.sh
test -f API_tests/test_bootstrap_restart_docker.sh
test -f API_tests/test_ui_interaction_acceptance.sh
test -f API_tests/ui_interaction_cdp.py
grep -q 'loginAttemptWindow = 15 \* time.Minute' internal/app/auth.go
grep -q 'LOGIN_ATTEMPT_WRITE_ERROR' internal/app/auth.go
grep -q 'LOGIN_STATE_WRITE_ERROR' internal/app/auth.go
grep -q 'AUTH_STATE_READ_ERROR' internal/app/auth.go
grep -q 'REPLAY_GUARD_ERROR' internal/app/auth.go
grep -q 'SESSION_UPDATE_ERROR' internal/app/auth.go
grep -q 'LOGOUT_WRITE_ERROR' internal/app/auth.go
grep -q 'bash API_tests/test_bootstrap_restart_docker.sh' ci/docker_acceptance.sh
grep -q 'bash API_tests/test_auth_lockout_docker.sh' ci/docker_acceptance.sh
grep -q 'bash API_tests/test_ui_interaction_acceptance.sh' ci/docker_acceptance.sh
grep -q 'IRONPAGE_UI_EVIDENCE_DIR=' ci/run_full_regression.sh
bash -n API_tests/test_auth_lockout_docker.sh
bash -n API_tests/test_bootstrap_restart_docker.sh
bash -n API_tests/test_ui_interaction_acceptance.sh
python3 -m py_compile API_tests/ui_interaction_cdp.py

# Markdown-only changes must have a dedicated consistency workflow and the
# current clarification contract must be executable in PR CI.
test -f ci/docs_consistency_check.sh
test -f .github/workflows/documentation-consistency.yml
grep -q 'bash ci/docs_consistency_check.sh' .github/workflows/documentation-consistency.yml
bash -n ci/docs_consistency_check.sh

# The acceptance image must remain capable of running the strict dependency
# contract without relying on the product runtime image.
docker build -f ci/Dockerfile.acceptance -t ironpage-vault-ci-acceptance-contract .
docker run --rm --entrypoint bash ironpage-vault-ci-acceptance-contract -lc '
  test -f internal/service/pdf.go
  test -f internal/platform/pdf_strict.go
  test -f internal/platform/backup_strict.go
  bash API_tests/test_strict_dependency_failures.sh
'

echo "PASS regression flow contract"
