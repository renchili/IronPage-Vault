#!/usr/bin/env bash
set -euo pipefail

# Pre-merge CI control-plane contract check.
# This executes the regression runner in a deliberate failing mode and verifies
# the CI contract: a failed stage must still produce results.tsv, summary.json,
# and summary.md, while the runner exits non-zero. It also verifies that failed
# stage log content is printed to stdout.

probe_dir="$(mktemp -d)"
deploy_probe_dir=""
cleanup() {
  rm -rf "$probe_dir"
  if [ -n "$deploy_probe_dir" ]; then
    rm -rf "$deploy_probe_dir"
  fi
}
trap cleanup EXIT

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

# A fresh checkout must have a real one-command normal deployment path. The
# deployer generates random external runtime values, stores them outside the
# image, reuses them on repeated runs, and leaves application fallback checks
# intact.
test -f scripts/deploy.sh
bash -n scripts/deploy.sh
grep -Fxq '.env' .gitignore
grep -Fxq '.env' .dockerignore
grep -q 'scripts/deploy.sh' API_tests/test_bootstrap_restart_docker.sh

deploy_probe_dir="$(mktemp -d)"
deploy_env="$deploy_probe_dir/runtime.env"
IRONPAGE_ENV_FILE="$deploy_env" IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh >"$deploy_probe_dir/first.log"
test -s "$deploy_env"
cp "$deploy_env" "$deploy_probe_dir/runtime.before"
IRONPAGE_ENV_FILE="$deploy_env" IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh >"$deploy_probe_dir/second.log"
cmp "$deploy_probe_dir/runtime.before" "$deploy_env"

docker compose --env-file "$deploy_env" config >/dev/null

python3 - "$deploy_env" <<'PY'
from pathlib import Path
import stat
import sys

path = Path(sys.argv[1])
values = {}
for line in path.read_text(encoding="utf-8").splitlines():
    if not line or line.startswith("#"):
        continue
    key, separator, value = line.partition("=")
    if not separator:
        raise SystemExit(f"malformed runtime line: {line!r}")
    values[key] = value

required = {
    "DB_HOST",
    "DB_PORT",
    "DB_USER",
    "DB_PASSWORD",
    "DB_NAME",
    "JWT_SECRET",
    "AES_KEY",
    "ACCEPTANCE_MODE",
    "BOOTSTRAP_ADMIN_USERNAME",
    "BOOTSTRAP_ADMIN_PASSWORD",
}
missing = sorted(required - values.keys())
if missing:
    raise SystemExit(f"generated runtime configuration is missing: {missing}")
if values["DB_HOST"] != "127.0.0.1" or values["DB_PORT"] != "5432":
    raise SystemExit("generated database endpoint does not target the embedded PostgreSQL service")
if values["DB_USER"] != "ironpage" or values["DB_NAME"] != "ironpage":
    raise SystemExit("generated database identity is inconsistent")
if len(values["DB_PASSWORD"]) < 16:
    raise SystemExit("generated DB_PASSWORD is too short")
if len(values["JWT_SECRET"]) < 32 or len(values["AES_KEY"]) < 32:
    raise SystemExit("generated cryptographic runtime values are too short")
if values["ACCEPTANCE_MODE"] != "false":
    raise SystemExit("one-command deployment must default to normal mode")
bootstrap_password_length = len(values["BOOTSTRAP_ADMIN_PASSWORD"].encode("utf-8"))
if not values["BOOTSTRAP_ADMIN_USERNAME"] or bootstrap_password_length < 16:
    raise SystemExit("generated initial Admin configuration is incomplete")
if bootstrap_password_length > 72:
    raise SystemExit("generated initial Admin password exceeds bcrypt's 72-byte limit")
mode = stat.S_IMODE(path.stat().st_mode)
if mode != 0o600:
    raise SystemExit(f"runtime configuration mode must be 0600, got {mode:o}")
PY

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
