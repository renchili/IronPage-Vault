#!/usr/bin/env bash
set -euo pipefail

probe_dir=$(mktemp -d)
deploy_probe_dir=$(mktemp -d)
cleanup() {
  rm -rf "$probe_dir" "$deploy_probe_dir"
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
grep -q -- '---- contract_fail failure log:' "$probe_dir/probe.log"

python3 - "$probe_dir/summary.json" <<'PY'
import json, sys
with open(sys.argv[1], encoding='utf-8') as f:
    summary = json.load(f)
if summary.get('overall_status') != 'failed':
    raise SystemExit('failing probe did not produce overall_status=failed')
stages = summary.get('stages', [])
if [stage.get('stage') for stage in stages] != ['contract_pass', 'contract_fail']:
    raise SystemExit(f'first-error stop failed; unexpected stages: {stages!r}')
if stages[0].get('status') != 0 or stages[1].get('status') == 0:
    raise SystemExit('contract probe stage statuses are invalid')
PY

grep -q 'Overall: \*\*FAILED\*\*' "$probe_dir/summary.md"

# The repository has one executable workflow. Concurrency, cooldown, failure
# latching, and all validation therefore share a single control plane.
mapfile -t workflows < <(find .github/workflows -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \))
test "${#workflows[@]}" -eq 1
test "${workflows[0]}" = ".github/workflows/ci.yml"
grep -q 'cancel-in-progress: false' .github/workflows/ci.yml
grep -q 'IRONPAGE_CI_TARGET' .github/workflows/ci.yml
grep -q 'ci/ci_execution_guard.py' .github/workflows/ci.yml
grep -q 'ci/run_full_regression.sh' .github/workflows/ci.yml
grep -q 'COOLDOWN_SECONDS = 10 \* 60' ci/ci_execution_guard.py
grep -q 'failed_same_revision' ci/ci_execution_guard.py
! grep -RIn 'if: always()' .github/workflows

# A fresh checkout generates every local runtime value without retaining fixed
# identities, ports, paths, credentials, or container names in Compose/code.
test -f scripts/deploy.sh
bash -n scripts/deploy.sh
grep -Fxq '.env' .gitignore
grep -Fxq '.env' .dockerignore
grep -Fxq 'reports/' .gitignore
grep -Fxq 'artifacts/' .gitignore
grep -q 'scripts/deploy.sh' tests/api/test_bootstrap_restart_docker.sh

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
for line in path.read_text(encoding='utf-8').splitlines():
    if not line or line.startswith('#'):
        continue
    key, separator, value = line.partition('=')
    if not separator:
        raise SystemExit(f'malformed runtime line: {line!r}')
    values[key] = value

required = {
    'HOST_BIND_ADDRESS', 'HOST_PORT', 'HTTP_PORT', 'HTTP_ADDR',
    'DB_PORT', 'DB_USER', 'DB_PASSWORD', 'DB_NAME',
    'JWT_SECRET', 'AES_KEY', 'ACCEPTANCE_MODE',
    'BOOTSTRAP_ADMIN_USERNAME', 'BOOTSTRAP_ADMIN_PASSWORD',
    'IRONPAGE_APP_ROOT', 'MIGRATIONS_DIR', 'PUBLIC_DIR',
    'POSTGRES_VOLUME_ROOT', 'PGDATA', 'IRONPAGE_VOLUME_ROOT',
    'STORAGE_DIR', 'BACKUP_DIR',
}
missing = sorted(key for key in required if not values.get(key))
if missing:
    raise SystemExit(f'generated runtime configuration is incomplete: {missing}')
for key in ('HOST_PORT', 'HTTP_PORT', 'DB_PORT'):
    port = int(values[key])
    if not 1024 <= port <= 65535:
        raise SystemExit(f'{key} is outside the unprivileged port range')
if len({values['HOST_PORT'], values['HTTP_PORT'], values['DB_PORT']}) != 3:
    raise SystemExit('generated ports are not independently configurable')
if not values['DB_USER'].startswith('ironpage_') or not values['DB_NAME'].startswith('ironpage_'):
    raise SystemExit('database identity is not installation-specific')
if len(values['DB_PASSWORD']) < 16 or len(values['JWT_SECRET']) < 32 or len(values['AES_KEY']) < 32:
    raise SystemExit('generated secrets are too short')
password_bytes = len(values['BOOTSTRAP_ADMIN_PASSWORD'].encode('utf-8'))
if not 16 <= password_bytes <= 72:
    raise SystemExit('generated bootstrap password is outside bcrypt limits')
if values['ACCEPTANCE_MODE'] != 'false':
    raise SystemExit('one-command deployment must default to normal mode')
if stat.S_IMODE(path.stat().st_mode) != 0o600:
    raise SystemExit('runtime environment file must use mode 0600')
PY

! grep -q 'container_name:' docker-compose.yml
! grep -Eq '\$\{(HOST_BIND_ADDRESS|DB_USER|DB_NAME|DB_PORT|HOST_PORT|HTTP_PORT|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR):-' docker-compose.yml
! grep -Eq 'env\("(HTTP_ADDR|DB_PORT|DB_USER|DB_NAME|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR)", "[^"]+"\)' internal/app/config.go
! grep -q 'DBHost' internal/app/config.go

# Canonical source and evidence layout must be unambiguous and clean.
test -d tests/api
test -d tests/contracts
test ! -e API_tests
test ! -e unit_tests
test -f public/index.html
test ! -e public/manual-test.html
test ! -e deploy/aws
test ! -e reports/regression
test ! -e PLAN.md
test ! -d docs/review-fixes || test -z "$(find docs/review-fixes -type f -print -quit)"
test -f ci/source_inventory.py

# Generated local reports may describe only their actual stage rows.
! grep -q 'This local report covers Swagger generation' run_tests.sh
grep -q 'Executed stages' run_tests.sh
grep -q 'executed_stages' run_tests.sh

echo "PASS regression flow contract"
