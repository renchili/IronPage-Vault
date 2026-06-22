#!/usr/bin/env bash
set -u -o pipefail

OUT_DIR=${1:-artifacts/regression}
LOG_DIR="$OUT_DIR/logs"
RESULTS="$OUT_DIR/results.tsv"
mkdir -p "$LOG_DIR"
printf 'stage	status	duration_seconds	log
' > "$RESULTS"

print_failure_log_tail() {
  local stage="$1"
  local log="$2"
  echo "---- ${stage} failure log: ${log} ----"
  if [ -s "$log" ]; then
    tail -n 200 "$log"
  else
    echo "log file is missing or empty"
  fi
  echo "---- end ${stage} failure log ----"
}

run_stage() {
  local stage="$1"
  shift
  local log="$LOG_DIR/${stage}.log"
  local started ended elapsed status
  started=$(date +%s)
  echo "==> $stage"
  set +e
  "$@" >"$log" 2>&1
  status=$?
  set +e
  ended=$(date +%s)
  elapsed=$((ended-started))
  if [ "$status" -eq 0 ]; then
    echo "PASS $stage (${elapsed}s)"
  else
    echo "FAIL $stage (${elapsed}s); see artifact log $log"
    print_failure_log_tail "$stage" "$log"
  fi
  printf '%s	%s	%s	%s
' "$stage" "$status" "$elapsed" "$log" >> "$RESULTS"
  return "$status"
}

write_summary() {
  python3 - "$OUT_DIR" "$RESULTS" <<'PY'
import csv, datetime, json, os, sys
out, results_path = sys.argv[1:]
rows = []
with open(results_path, newline='', encoding='utf-8') as f:
    for row in csv.DictReader(f, delimiter='\t'):
        row['status'] = int(row['status'])
        row['duration_seconds'] = int(row['duration_seconds'])
        rows.append(row)
passed = all(row['status'] == 0 for row in rows)
payload = {
    'generated_at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'overall_status': 'passed' if passed else 'failed',
    'stages': rows,
}
with open(os.path.join(out, 'summary.json'), 'w', encoding='utf-8') as f:
    json.dump(payload, f, indent=2)
with open(os.path.join(out, 'summary.md'), 'w', encoding='utf-8') as f:
    f.write('# Full Regression Result\n\n')
    f.write(f"Overall: **{'PASSED' if passed else 'FAILED'}**\n\n")
    f.write('| Stage | Result | Duration | Log |\n|---|---:|---:|---|\n')
    for row in rows:
        label = 'PASS' if row['status'] == 0 else 'FAIL'
        f.write(f"| {row['stage']} | {label} | {row['duration_seconds']}s | `{row['log']}` |\n")
sys.exit(0 if passed else 1)
PY
}

if [ "${IRONPAGE_REGRESSION_CONTRACT_PROBE:-}" = "1" ]; then
  run_stage contract_pass bash -lc 'true'
  run_stage contract_fail bash -lc 'echo IRONPAGE_REGRESSION_CONTRACT_FAIL_SENTINEL >&2; false'
  write_summary
  exit $?
fi

# CI control-plane rule: this script may execute CI-owned scripts under ci/**
# and standard tool commands. It must not call run_tests.sh or project-owned
# API_tests/unit_tests shell scripts as the source of a pre-merge conclusion.

if ! run_stage prepare_workspace bash -lc '
  set -euo pipefail
  mkdir -p docs/swagger
  printf "package swagger\n" > docs/swagger/docs.go
  go mod tidy
  go install github.com/swaggo/swag/cmd/swag@v1.16.4
  SWAG_BIN="$(go env GOPATH)/bin/swag" bash scripts/generate_swagger.sh
  test -f docs/swagger/docs.go
  test -f docs/swagger/swagger.yaml
'; then
  write_summary
  exit 1
fi

run_stage swagger_contract bash -lc 'test -s docs/swagger/docs.go && test -s docs/swagger/swagger.yaml && grep -q "/api/auth/login" docs/swagger/swagger.yaml && grep -q "/api/documents" docs/swagger/swagger.yaml'
run_stage swagger_route_coverage bash unit_tests/test_swagger_route_coverage.sh
run_stage scheduled_backup_contract bash ci/scheduled_backup_contract_check.sh
run_stage metadata_storage_contract bash ci/metadata_storage_check.sh
run_stage gofmt bash -lc 'find cmd internal -name "*.go" -print0 | xargs -0 gofmt -l | tee /tmp/ironpage_gofmt.out; test ! -s /tmp/ironpage_gofmt.out'
run_stage go_vet bash -lc 'go vet ./...'
run_stage go_test_race bash -lc 'go test -mod=mod -race ./...'
run_stage shell_syntax bash ci/shell_syntax_check.sh
run_stage docker_build bash -lc 'docker build --progress=plain -t ironpage-vault:regression .'
run_stage docker_acceptance bash ci/docker_acceptance.sh

write_summary
