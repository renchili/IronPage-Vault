#!/usr/bin/env bash
set -euo pipefail

OUT_DIR=${1:-artifacts/regression}
LOG_DIR="$OUT_DIR/logs"
RESULTS="$OUT_DIR/results.tsv"
mkdir -p "$LOG_DIR"
printf 'stage\tstatus\tduration_seconds\tlog\n' > "$RESULTS"

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

write_summary() {
  python3 - "$OUT_DIR" "$RESULTS" <<'PY'
import csv, datetime, json, os, subprocess, sys
out, results_path = sys.argv[1:]
rows = []
with open(results_path, newline='', encoding='utf-8') as f:
    for row in csv.DictReader(f, delimiter='\t'):
        row['status'] = int(row['status'])
        row['duration_seconds'] = int(row['duration_seconds'])
        rows.append(row)
passed = bool(rows) and all(row['status'] == 0 for row in rows)
payload = {
    'generated_at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip(),
    'overall_status': 'passed' if passed else 'failed',
    'total_stages': len(rows),
    'stages': rows,
}
with open(os.path.join(out, 'summary.json'), 'w', encoding='utf-8') as f:
    json.dump(payload, f, indent=2)
with open(os.path.join(out, 'summary.md'), 'w', encoding='utf-8') as f:
    f.write('# Full Regression Result\n\n')
    f.write(f"Commit: `{payload['commit']}`\n\n")
    f.write(f"Generated: `{payload['generated_at']}`\n\n")
    f.write(f"Overall: **{'PASSED' if passed else 'FAILED'}**\n\n")
    f.write('| Stage | Result | Duration | Log |\n|---|---:|---:|---|\n')
    for row in rows:
        label = 'PASS' if row['status'] == 0 else 'FAIL'
        f.write(f"| {row['stage']} | {label} | {row['duration_seconds']}s | `{row['log']}` |\n")
PY
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
  set -e
  ended=$(date +%s)
  elapsed=$((ended-started))
  printf '%s\t%s\t%s\t%s\n' "$stage" "$status" "$elapsed" "$log" >> "$RESULTS"
  if [ "$status" -eq 0 ]; then
    echo "PASS $stage (${elapsed}s)"
    return 0
  fi
  echo "FAIL $stage (${elapsed}s); see artifact log $log"
  print_failure_log_tail "$stage" "$log"
  write_summary
  exit "$status"
}

if [ "${IRONPAGE_REGRESSION_CONTRACT_PROBE:-}" = "1" ]; then
  run_stage contract_pass bash -lc 'true'
  run_stage contract_fail bash -lc 'echo IRONPAGE_REGRESSION_CONTRACT_FAIL_SENTINEL >&2; false'
fi

# Every stage is sequential. A failed stage records its result, writes the
# summary, and exits before any later build, test, or upload can start.
run_stage source_inventory python3 ci/source_inventory.py "$OUT_DIR/source-inventory.json"
run_stage documentation_consistency bash ci/docs_consistency_check.sh
run_stage regression_failure_contract bash ci/regression_contract_check.sh
run_stage local_entrypoint_contract bash ci/run_tests_contract_check.sh
run_stage prepare_workspace bash -lc '
  set -euo pipefail
  mkdir -p docs/swagger
  printf "package swagger\n" > docs/swagger/docs.go
  go mod tidy
  go install github.com/swaggo/swag/cmd/swag@v1.16.4
  SWAG_BIN="$(go env GOPATH)/bin/swag" bash scripts/generate_swagger.sh
  test -f docs/swagger/docs.go
  test -f docs/swagger/swagger.yaml
'
run_stage swagger_contract bash ci/swagger_contract_check.sh
run_stage swagger_route_coverage bash tests/contracts/swagger_route_coverage.sh
run_stage repository_contracts bash tests/contracts/repository_rules.sh
run_stage structure_contracts bash tests/contracts/structure_rules.sh
run_stage scheduled_backup_contract bash ci/scheduled_backup_contract_check.sh
run_stage metadata_storage_contract bash ci/metadata_storage_check.sh
run_stage gofmt bash -lc 'find cmd internal -name "*.go" -print0 | xargs -0 gofmt -l | tee /tmp/ironpage_gofmt.out; test ! -s /tmp/ironpage_gofmt.out'
run_stage go_vet bash -lc 'go vet ./...'
run_stage go_test_race bash -lc 'go test -mod=mod -race ./...'
run_stage shell_syntax bash ci/shell_syntax_check.sh
run_stage docker_build bash -lc '
  set -euo pipefail
  env_file=$(mktemp)
  trap "rm -f \"$env_file\"" EXIT
  IRONPAGE_ENV_FILE="$env_file" IRONPAGE_DEPLOY_DRY_RUN=true bash scripts/deploy.sh
  docker compose --env-file "$env_file" build ironpage
'
run_stage docker_acceptance env IRONPAGE_UI_EVIDENCE_DIR="$OUT_DIR/ui-interaction" bash ci/docker_acceptance.sh

write_summary
