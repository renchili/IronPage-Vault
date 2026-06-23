#!/usr/bin/env bash
set -u -o pipefail

OUT_DIR=${IRONPAGE_ACCEPTANCE_REPORT_DIR:-artifacts/local-acceptance}
LOG_DIR="$OUT_DIR/logs"
RESULTS="$OUT_DIR/results.tsv"
mkdir -p "$LOG_DIR"
printf 'stage	result	status	duration_seconds	log
' > "$RESULTS"
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
  local stage="$1"
  local result="$2"
  local status="$3"
  local elapsed="$4"
  local log="$5"
  printf '%s	%s	%s	%s	%s
' "$stage" "$result" "$status" "$elapsed" "$log" >> "$RESULTS"
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
    result="PASS"
    echo "PASS $stage (${elapsed}s)"
  else
    result="FAIL"
    OVERALL_STATUS=1
    echo "FAIL $stage (${elapsed}s); see artifact log $log"
    print_failure_log_tail "$stage" "$log"
  fi
  record_stage "$stage" "$result" "$status" "$elapsed" "$log"
  return 0
}

skip_stage() {
  local stage="$1"
  local reason="$2"
  local log="$LOG_DIR/${stage}.log"
  printf '%s
' "$reason" > "$log"
  echo "SKIP $stage: $reason"
  record_stage "$stage" "SKIP" "0" "0" "$log"
}

run_script() {
  local stage="$1"
  local script="$2"
  if [ -f "$script" ]; then
    run_stage "$stage" bash "$script"
  else
    skip_stage "$stage" "script not found: $script"
  fi
}

load_api_tokens() {
  if [ -s /tmp/ironpage_admin_token.out ]; then
    export ADMIN_TOKEN="$(cat /tmp/ironpage_admin_token.out)"
  fi
  if [ -s /tmp/ironpage_editor_token.out ]; then
    export EDITOR_TOKEN="$(cat /tmp/ironpage_editor_token.out)"
  fi
  if [ -s /tmp/ironpage_reviewer_token.out ]; then
    export REVIEWER_TOKEN="$(cat /tmp/ironpage_reviewer_token.out)"
  fi
}

prepare_swagger() {
  mkdir -p docs/swagger
  printf 'package swagger
' > docs/swagger/docs.go

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

out_dir, results_path, overall_status = sys.argv[1], sys.argv[2], int(sys.argv[3])
rows = []
with open(results_path, newline='', encoding='utf-8') as f:
    for row in csv.DictReader(f, delimiter='\t'):
        row['status'] = int(row['status'])
        row['duration_seconds'] = int(row['duration_seconds'])
        rows.append(row)

counts = {
    'PASS': sum(1 for row in rows if row['result'] == 'PASS'),
    'FAIL': sum(1 for row in rows if row['result'] == 'FAIL'),
    'SKIP': sum(1 for row in rows if row['result'] == 'SKIP'),
}
passed = overall_status == 0 and counts['FAIL'] == 0
payload = {
    'generated_at': dt.datetime.now(dt.timezone.utc).isoformat(),
    'overall_status': 'passed' if passed else 'failed',
    'total_stages': len(rows),
    'passed': counts['PASS'],
    'failed': counts['FAIL'],
    'skipped': counts['SKIP'],
    'stages': rows,
}
summary_json = os.path.join(out_dir, 'summary.json')
summary_md = os.path.join(out_dir, 'summary.md')
report_html = os.path.join(out_dir, 'report.html')
with open(summary_json, 'w', encoding='utf-8') as f:
    json.dump(payload, f, indent=2)
with open(summary_md, 'w', encoding='utf-8') as f:
    f.write('# IronPage Local Acceptance Report\n\n')
    f.write(f"Generated: `{payload['generated_at']}`\n\n")
    f.write(f"Overall: **{'PASSED' if passed else 'FAILED'}**\n\n")
    f.write(f"Stages: {len(rows)} total / {counts['PASS']} passed / {counts['FAIL']} failed / {counts['SKIP']} skipped\n\n")
    f.write('| Stage | Result | Exit | Duration | Log |\n|---|---:|---:|---:|---|\n')
    for row in rows:
        f.write(f"| `{row['stage']}` | {row['result']} | {row['status']} | {row['duration_seconds']}s | `{row['log']}` |\n")

badge = 'PASSED' if passed else 'FAILED'
badge_class = 'pass' if passed else 'fail'
rows_html = []
for row in rows:
    cls = row['result'].lower()
    rel_log = os.path.relpath(row['log'], out_dir)
    rows_html.append(
        '<tr>'
        f'<td>{html.escape(row["stage"])}</td>'
        f'<td><span class="pill {cls}">{html.escape(row["result"])}</span></td>'
        f'<td>{row["status"]}</td>'
        f'<td>{row["duration_seconds"]}s</td>'
        f'<td><a href="{html.escape(rel_log)}">log</a></td>'
        '</tr>'
    )

html_doc = f'''<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>IronPage Local Acceptance Report</title>
  <style>
    :root {{ color-scheme: light dark; --ok:#1a7f37; --bad:#cf222e; --skip:#9a6700; --muted:#6e7781; --border:#d0d7de; }}
    body {{ font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; padding: 32px; background: Canvas; color: CanvasText; }}
    .wrap {{ max-width: 1120px; margin: 0 auto; }}
    h1 {{ margin-bottom: 8px; }}
    .muted {{ color: var(--muted); }}
    .cards {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px; margin: 24px 0; }}
    .card {{ border: 1px solid var(--border); border-radius: 12px; padding: 16px; }}
    .number {{ font-size: 28px; font-weight: 700; }}
    .pill {{ display:inline-block; border-radius:999px; padding: 4px 10px; font-weight:700; color:white; font-size:12px; }}
    .pass {{ background: var(--ok); }}
    .fail {{ background: var(--bad); }}
    .skip {{ background: var(--skip); }}
    table {{ width: 100%; border-collapse: collapse; border: 1px solid var(--border); border-radius: 12px; overflow: hidden; }}
    th, td {{ border-bottom: 1px solid var(--border); padding: 10px 12px; text-align: left; }}
    th {{ background: rgba(127,127,127,0.08); }}
    tr:last-child td {{ border-bottom: 0; }}
    a {{ color: LinkText; }}
    .section {{ margin-top: 28px; }}
    code {{ background: rgba(127,127,127,0.12); padding: 2px 5px; border-radius: 5px; }}
  </style>
</head>
<body>
  <main class="wrap">
    <h1>IronPage Local Acceptance Report</h1>
    <p class="muted">Generated at {html.escape(payload['generated_at'])}</p>
    <p><span class="pill {badge_class}">{badge}</span></p>
    <div class="cards">
      <div class="card"><div class="number">{len(rows)}</div><div class="muted">total stages</div></div>
      <div class="card"><div class="number">{counts['PASS']}</div><div class="muted">passed</div></div>
      <div class="card"><div class="number">{counts['FAIL']}</div><div class="muted">failed</div></div>
      <div class="card"><div class="number">{counts['SKIP']}</div><div class="muted">skipped</div></div>
    </div>
    <section class="section">
      <h2>Stage Results</h2>
      <table>
        <thead><tr><th>Stage</th><th>Result</th><th>Exit</th><th>Duration</th><th>Log</th></tr></thead>
        <tbody>{''.join(rows_html)}</tbody>
      </table>
    </section>
    <section class="section">
      <h2>Acceptance Coverage</h2>
      <p>This local report covers Swagger generation, unit/static rules, Go tests, API acceptance scripts, redaction/Bates/compare flows, notifications, and strict dependency checks. Logs are retained under <code>logs/</code>.</p>
      <p>For runtime Docker/full-regression evidence, run <code>bash ci/run_full_regression.sh artifacts/regression</code>.</p>
    </section>
  </main>
</body>
</html>
'''
with open(report_html, 'w', encoding='utf-8') as f:
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

run_script unit_rules unit_tests/test_rules.sh
run_stage go_test_all go test -mod=mod ./...

run_script api_flow API_tests/test_api_flow.sh
load_api_tokens

run_script api_contracts API_tests/test_api_contracts.sh
load_api_tokens

run_script static_review_reject_flows API_tests/test_static_review_reject_flows.sh
load_api_tokens

run_script acceptance_denials API_tests/test_acceptance_denials.sh
load_api_tokens

run_script compare_acceptance API_tests/test_compare_acceptance.sh
load_api_tokens

run_script finalized_immutability API_tests/test_finalized_immutability.sh
load_api_tokens

run_script redaction_coordinate_ciphertext API_tests/test_redaction_coordinate_ciphertext.sh
load_api_tokens

run_script pdf_content_acceptance API_tests/test_pdf_content_acceptance.sh
load_api_tokens

run_script notification_mention_side_effect API_tests/test_notification_mention_side_effect.sh
load_api_tokens

run_script structure_rules unit_tests/test_structure_rules.sh
run_script strict_dependency_failures API_tests/test_strict_dependency_failures.sh
run_script bates_sequence_multi_doc API_tests/test_bates_sequence_multi_doc.sh

write_acceptance_report
exit "$OVERALL_STATUS"
