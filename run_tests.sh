#!/usr/bin/env bash
set -u -o pipefail

OUT_DIR=${IRONPAGE_ACCEPTANCE_REPORT_DIR:-artifacts/local-acceptance}
LOG_DIR="$OUT_DIR/logs"
RESULTS="$OUT_DIR/results.tsv"
mkdir -p "$LOG_DIR"
printf 'stage\tresult\tstatus\tduration_seconds\tlog\n' > "$RESULTS"
OVERALL_STATUS=0

print_failure_log_tail() {
  local stage="$1"
  local log="$2"
  echo "---- ${stage} failure log: ${log} ----"
  if [ -s "$log" ]; then
    tail -n 120 "$log"
  else
    echo "log file is missing or empty"
  fi
  echo "---- end ${stage} failure log ----"
}

record_stage() {
  local stage="$1" result="$2" status="$3" elapsed="$4" log="$5"
  printf '%s\t%s\t%s\t%s\t%s\n' "$stage" "$result" "$status" "$elapsed" "$log" >> "$RESULTS"
}

run_stage() {
  local stage="$1"
  shift
  local log="$LOG_DIR/${stage}.log"
  local started ended elapsed status result
  started=$(date +%s)
  echo "==> $stage"
  set +e
  "$@" >"$log" 2>&1
  status=$?
  set +e
  ended=$(date +%s)
  elapsed=$((ended-started))
  if [ "$status" -eq 0 ]; then
    result=PASS
    echo "PASS $stage (${elapsed}s)"
  else
    result=FAIL
    OVERALL_STATUS=1
    echo "FAIL $stage (${elapsed}s); see artifact log $log"
    print_failure_log_tail "$stage" "$log"
  fi
  record_stage "$stage" "$result" "$status" "$elapsed" "$log"
  return 0
}

skip_stage() {
  local stage="$1" reason="$2"
  local log="$LOG_DIR/${stage}.log"
  printf '%s\n' "$reason" > "$log"
  echo "SKIP $stage: $reason"
  record_stage "$stage" SKIP 0 0 "$log"
  if [ "$OVERALL_STATUS" -eq 0 ]; then
    OVERALL_STATUS=2
  fi
}

run_script() {
  local stage="$1" script="$2"
  if [ -f "$script" ]; then
    run_stage "$stage" bash "$script"
  else
    skip_stage "$stage" "script not found: $script"
  fi
}

load_api_tokens() {
  if [ -s /tmp/ironpage_admin_token.out ]; then export ADMIN_TOKEN="$(cat /tmp/ironpage_admin_token.out)"; fi
  if [ -s /tmp/ironpage_editor_token.out ]; then export EDITOR_TOKEN="$(cat /tmp/ironpage_editor_token.out)"; fi
  if [ -s /tmp/ironpage_reviewer_token.out ]; then export REVIEWER_TOKEN="$(cat /tmp/ironpage_reviewer_token.out)"; fi
}

prepare_swagger() {
  mkdir -p docs/swagger
  printf 'package swagger\n' > docs/swagger/docs.go
  if command -v swag >/dev/null 2>&1; then
    SWAG_BIN="$(command -v swag)" bash scripts/generate_swagger.sh
    return
  fi
  go install github.com/swaggo/swag/cmd/swag@v1.16.4
  SWAG_BIN="$(go env GOPATH)/bin/swag" bash scripts/generate_swagger.sh
}

write_acceptance_report() {
  python3 - "$OUT_DIR" "$RESULTS" "$OVERALL_STATUS" <<'PY'
import csv
import datetime as dt
import html
import json
import os
import sys

out_dir, results_path, exit_status = sys.argv[1], sys.argv[2], int(sys.argv[3])
rows = []
with open(results_path, newline='', encoding='utf-8') as f:
    for row in csv.DictReader(f, delimiter='\t'):
        row['status'] = int(row['status'])
        row['duration_seconds'] = int(row['duration_seconds'])
        rows.append(row)

counts = {name: sum(1 for row in rows if row['result'] == name) for name in ('PASS', 'FAIL', 'SKIP')}
if exit_status == 1 or counts['FAIL']:
    state = 'failed'
elif exit_status == 2 or counts['SKIP']:
    state = 'incomplete'
elif rows:
    state = 'passed'
else:
    state = 'incomplete'
executed_stages = [row['stage'] for row in rows]
ui_manifest_path = os.path.join(out_dir, 'ui', 'manifest.json')
ui_manifest = None
if os.path.exists(ui_manifest_path):
    with open(ui_manifest_path, encoding='utf-8') as f:
        ui_manifest = json.load(f)

payload = {
    'generated_at': dt.datetime.now(dt.timezone.utc).isoformat(),
    'overall_status': state,
    'total_stages': len(rows),
    'passed': counts['PASS'],
    'failed': counts['FAIL'],
    'skipped': counts['SKIP'],
    'executed_stages': executed_stages,
    'stages': rows,
    'ui_screenshot': ui_manifest,
}

with open(os.path.join(out_dir, 'summary.json'), 'w', encoding='utf-8') as f:
    json.dump(payload, f, indent=2)
with open(os.path.join(out_dir, 'summary.md'), 'w', encoding='utf-8') as f:
    f.write('# IronPage Local Acceptance Report\n\n')
    f.write(f"Generated: `{payload['generated_at']}`\n\n")
    f.write(f"Overall: **{state.upper()}**\n\n")
    f.write(f"Stages: {len(rows)} total / {counts['PASS']} passed / {counts['FAIL']} failed / {counts['SKIP']} skipped\n\n")
    f.write('## Executed stages\n\n')
    for stage in executed_stages:
        f.write(f'- `{stage}`\n')
    f.write('\n| Stage | Result | Exit | Duration | Log |\n|---|---:|---:|---:|---|\n')
    for row in rows:
        f.write(f"| `{row['stage']}` | {row['result']} | {row['status']} | {row['duration_seconds']}s | `{row['log']}` |\n")
    if ui_manifest:
        f.write('\n## UI screenshot evidence\n\n')
        f.write(f"- Page: `{ui_manifest['page']}`\n")
        f.write(f"- Browser: `{ui_manifest['browser']}`\n")
        f.write(f"- Screenshot: `ui/{ui_manifest['screenshot']}`\n")

rows_html = []
for row in rows:
    rel_log = os.path.relpath(row['log'], out_dir)
    rows_html.append(
        '<tr>'
        f'<td>{html.escape(row["stage"])}</td>'
        f'<td>{html.escape(row["result"])}</td>'
        f'<td>{row["status"]}</td>'
        f'<td>{row["duration_seconds"]}s</td>'
        f'<td><a href="{html.escape(rel_log)}">log</a></td>'
        '</tr>'
    )
stage_items = ''.join(f'<li><code>{html.escape(stage)}</code></li>' for stage in executed_stages)
ui_section = ''
if ui_manifest:
    screenshot_src = 'ui/' + ui_manifest['screenshot']
    ui_section = (
        '<section><h2>UI screenshot evidence</h2>'
        f'<p>Browser: {html.escape(ui_manifest["browser"])}</p>'
        f'<img src="{html.escape(screenshot_src)}" alt="IronPage acceptance UI screenshot">'
        '</section>'
    )

html_doc = f'''<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>IronPage Local Acceptance Report</title>
<style>
body{{font-family:system-ui,sans-serif;max-width:1120px;margin:auto;padding:32px}}
table{{width:100%;border-collapse:collapse}}th,td{{border:1px solid #aaa;padding:8px;text-align:left}}
code{{background:#eee;padding:2px 5px}}img{{max-width:100%}}
</style>
</head>
<body>
<h1>IronPage Local Acceptance Report</h1>
<p>Generated at {html.escape(payload['generated_at'])}</p>
<p><strong>{state.upper()}</strong></p>
<h2>Executed stages</h2>
<p>This report claims coverage only for the stage rows recorded below. Any skipped row makes the report incomplete.</p>
<ul>{stage_items}</ul>
<table>
<thead><tr><th>Stage</th><th>Result</th><th>Exit</th><th>Duration</th><th>Log</th></tr></thead>
<tbody>{''.join(rows_html)}</tbody>
</table>
{ui_section}
<p>Docker and full-regression evidence is separate and must be generated by <code>bash ci/run_full_regression.sh artifacts/regression</code>.</p>
</body>
</html>
'''
with open(os.path.join(out_dir, 'report.html'), 'w', encoding='utf-8') as f:
    f.write(html_doc)
PY
  echo "Local acceptance report: $OUT_DIR/report.html"
  echo "Local acceptance summary: $OUT_DIR/summary.md"
}

run_stage prepare_swagger prepare_swagger

if [ "${IRONPAGE_RUN_TESTS_CONTRACT_PROBE:-}" = "1" ]; then
  run_stage local_entrypoint_contract go test -mod=mod ./internal/core
  write_acceptance_report
  test -s docs/swagger/docs.go
  test -s docs/swagger/swagger.yaml
  echo "PASS run_tests local entrypoint contract"
  exit "$OVERALL_STATUS"
fi

run_script repository_contracts tests/contracts/repository_rules.sh
run_stage go_test_all go test -mod=mod ./...
run_script structure_contracts tests/contracts/structure_rules.sh
run_script strict_dependency_failures tests/api/test_strict_dependency_failures.sh

api_reason=""
for name in BASE_URL SEED_ADMIN_PASSWORD SEED_EDITOR_PASSWORD SEED_REVIEWER_PASSWORD; do
  if [ -z "${!name:-}" ]; then
    api_reason="stateful acceptance requires BASE_URL and all SEED_*_PASSWORD values"
    break
  fi
done

api_stages=(
  api_flow
  api_contracts
  static_review_reject_flows
  acceptance_denials
  request_guard_edges
  compare_acceptance
  finalized_immutability
  pdf_content_acceptance
  notification_mention_side_effect
  ui_screenshot_acceptance
  bates_sequence_multi_doc
)

if [ -n "$api_reason" ]; then
  for stage in "${api_stages[@]}"; do
    skip_stage "$stage" "$api_reason"
  done
else
  rm -f /tmp/ironpage_admin_token.out /tmp/ironpage_editor_token.out /tmp/ironpage_reviewer_token.out
  run_script api_flow tests/api/test_api_flow.sh
  load_api_tokens
  run_script api_contracts tests/api/test_api_contracts.sh
  load_api_tokens
  run_script static_review_reject_flows tests/api/test_static_review_reject_flows.sh
  load_api_tokens
  run_script acceptance_denials tests/api/test_acceptance_denials.sh
  load_api_tokens
  run_script request_guard_edges tests/api/test_request_guard_edges.sh
  load_api_tokens
  run_script compare_acceptance tests/api/test_compare_acceptance.sh
  load_api_tokens
  run_script finalized_immutability tests/api/test_finalized_immutability.sh
  load_api_tokens
  run_script pdf_content_acceptance tests/api/test_pdf_content_acceptance.sh
  load_api_tokens
  run_script notification_mention_side_effect tests/api/test_notification_mention_side_effect.sh
  load_api_tokens
  run_script ui_screenshot_acceptance tests/api/test_ui_screenshot_acceptance.sh
  load_api_tokens
  run_script bates_sequence_multi_doc tests/api/test_bates_sequence_multi_doc.sh
fi

write_acceptance_report
exit "$OVERALL_STATUS"
