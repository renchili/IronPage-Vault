#!/usr/bin/env bash
set -u -o pipefail

OUT_DIR=${1:-artifacts/regression}
LOG_DIR="$OUT_DIR/logs"
RESULTS="$OUT_DIR/results.tsv"
mkdir -p "$LOG_DIR"
printf 'stage\tstatus\tduration_seconds\tlog\n' > "$RESULTS"

failed=0
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
  if [ "$status" -eq 0 ]; then
    echo "PASS $stage (${elapsed}s)"
  else
    echo "FAIL $stage (${elapsed}s); see $log"
    failed=1
  fi
  printf '%s\t%s\t%s\t%s\n' "$stage" "$status" "$elapsed" "$log" >> "$RESULTS"
}

run_stage swagger bash -lc 'bash scripts/generate_swagger.sh && test -f docs/swagger/docs.go && test -f docs/swagger/swagger.yaml && bash unit_tests/test_swagger_route_coverage.sh'
run_stage gofmt bash -lc "test -z \"\$(find cmd internal -name '*.go' -print0 | xargs -0 gofmt -l)\""
run_stage go_vet bash -lc 'go vet ./...'
run_stage go_test_race bash -lc 'go test -mod=mod -race ./...'
run_stage static_rules bash -lc 'bash unit_tests/test_rules.sh && bash unit_tests/test_structure_rules.sh'
run_stage docker_build bash -lc 'docker build --progress=plain -t ironpage-vault:regression .'
run_stage docker_acceptance bash -lc 'bash scripts/docker_acceptance.sh'

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
