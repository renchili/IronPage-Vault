#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh
FAIL=0
: "${ADMIN_TOKEN:?set ADMIN_TOKEN}"

assert_audit_action_filter() {
  python3 - "$BODY" <<'PY'
import json, sys
with open(sys.argv[1]) as f:
    body = json.load(f)
rows = body.get("data", [])
if not rows:
    print("FAIL api: audit action filter returned no DOCUMENT_UPLOAD rows")
    sys.exit(1)
bad = [r for r in rows if r.get("action_type") != "DOCUMENT_UPLOAD"]
if bad:
    print("FAIL api: audit action filter returned non DOCUMENT_UPLOAD rows")
    print(json.dumps(bad[:3], indent=2))
    sys.exit(1)
for row in rows:
    if not row.get("source_ip"):
        print("FAIL api: audit row did not open protected source_ip")
        sys.exit(1)
    metadata = row.get("metadata")
    if not isinstance(metadata, dict) or metadata.get("version") != 1:
        print("FAIL api: audit row did not open structured upload metadata")
        print(json.dumps(row, indent=2))
        sys.exit(1)
print("PASS api: audit action filter rows and protected fields match")
PY
}

code=$(auth_get "$ADMIN_TOKEN" /api/admin/config)
expect_code "admin config list" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/backup.local_volume '{"value":"/tmp/not-runtime-owned"}')
expect_code "deployment-owned config is read-only" 409 "$code" || FAIL=$((FAIL+1))
expect_json_field "deployment-owned config error" error.code CONFIG_KEY_READ_ONLY || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/unknown.key '{"value":"1"}')
expect_code "unknown config key is rejected" 400 "$code" || FAIL=$((FAIL+1))
expect_json_field "unknown config error" error.code CONFIG_KEY_NOT_MANAGED || FAIL=$((FAIL+1))

code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/backup.schedule_enabled '{"value":"enabled"}')
expect_code "backup schedule enabled rejects invalid boolean" 400 "$code" || FAIL=$((FAIL+1))
expect_json_field "backup schedule enabled error" error.code INVALID_CONFIG_VALUE || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/backup.interval '{"value":"59s"}')
expect_code "backup interval rejects below minimum" 400 "$code" || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/backup.interval '{"value":"169h"}')
expect_code "backup interval rejects above maximum" 400 "$code" || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/backup.interval '{"value":"30m"}')
expect_code "backup interval accepts managed duration" 200 "$code" || FAIL=$((FAIL+1))
expect_json_field "backup interval normalized value" value 30m0s || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/backup.schedule_enabled '{"value":"true"}')
expect_code "Admin enables scheduled backup" 200 "$code" || FAIL=$((FAIL+1))
expect_json_field "scheduled backup enabled value" value true || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/backup.schedule_enabled '{"value":"false"}')
expect_code "Admin disables scheduled backup" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/pagination.default_page_size '{"value":"0"}')
expect_code "pagination default rejects zero" 400 "$code" || FAIL=$((FAIL+1))
expect_json_field "pagination default error" error.code INVALID_CONFIG_VALUE || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/pagination.max_page_size '{"value":"101"}')
expect_code "pagination max rejects above absolute maximum" 400 "$code" || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/pagination.max_page_size '{"value":"20"}')
expect_code "pagination max rejects below current default" 400 "$code" || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/pagination.max_page_size '{"value":"100"}')
expect_code "pagination max accepts 100" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/config/pagination.default_page_size '{"value":"25"}')
expect_code "pagination default accepts 25" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" /api/admin/workflow-statuses)
expect_code "admin workflow statuses" 200 "$code" || FAIL=$((FAIL+1))
workflow_body='{"statuses":[{"name":"Draft","mutable":true},{"name":"Under Review","mutable":true},{"name":"Redaction Pending","mutable":true},{"name":"Approved","mutable":true},{"name":"Finalized","mutable":false}]}'
code=$(auth_put_json "$ADMIN_TOKEN" /api/admin/workflow-statuses "$workflow_body")
expect_code "admin replaces workflow statuses" 200 "$code" || FAIL=$((FAIL+1))
expect_json_field "workflow final status" data[4].name Finalized || FAIL=$((FAIL+1))
expect_json_field "workflow final status immutable" data[4].mutable false || FAIL=$((FAIL+1))
code=$(auth_put_json "$ADMIN_TOKEN" /api/admin/workflow-statuses '{"statuses":[{"name":"Draft","mutable":true},{"name":"Finalized","mutable":true}]}')
expect_code "workflow rejects mutable Finalized" 400 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$ADMIN_TOKEN" /api/admin/notification-templates)
expect_code "admin notification templates" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_patch_json "$ADMIN_TOKEN" /api/admin/notification-templates/workflow.transition '{"subject":"Workflow transition","body":"Document status changed"}')
expect_code "admin updates notification template" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$ADMIN_TOKEN" /api/audit-logs)
expect_code "audit log list" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$ADMIN_TOKEN" '/api/audit-logs?action_type=DOCUMENT_UPLOAD')
expect_code "audit log action filter" 200 "$code" || FAIL=$((FAIL+1))
assert_audit_action_filter || FAIL=$((FAIL+1))

code=$(auth_post_json "$ADMIN_TOKEN" /api/admin/backup/run '{}')
expect_code "admin backup run" 201 "$code" || FAIL=$((FAIL+1))
expect_json_field "admin backup status" status Completed || FAIL=$((FAIL+1))
expect_json_field "admin backup kind" kind full_backup || FAIL=$((FAIL+1))
expect_json_field "admin backup dump mode" artifacts.database_dump_mode pg_dump_custom || FAIL=$((FAIL+1))
expect_json_field "admin backup file mode" artifacts.file_snapshot_mode tar || FAIL=$((FAIL+1))
expect_json_field "admin backup restore supported" restore_supported true || FAIL=$((FAIL+1))
expect_json_nonempty "admin backup dump path" artifacts.database_dump_path || FAIL=$((FAIL+1))
expect_json_nonempty "admin backup file path" artifacts.file_snapshot_path || FAIL=$((FAIL+1))
DB_DUMP_PATH="$(json_field artifacts.database_dump_path)"
FILE_SNAPSHOT_PATH="$(json_field artifacts.file_snapshot_path)"

code=$(auth_get "$ADMIN_TOKEN" /api/admin/backup/jobs)
expect_code "admin backup jobs" 200 "$code" || FAIL=$((FAIL+1))
if ! grep -q "full_backup" "$BODY"; then
  echo "FAIL api: backup jobs response does not include full_backup"
  cat "$BODY"
  FAIL=$((FAIL+1))
else
  echo "PASS api: backup jobs include full_backup"
fi

code=$(auth_post_json "$ADMIN_TOKEN" /api/admin/backup/restore '{}')
expect_code "admin restore rejects empty request" 400 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$ADMIN_TOKEN" /api/admin/backup/restore/rst_missing/resolve '{"status":"Completed","verification_note":"verified against restored files and database"}')
expect_code "missing interrupted restore cannot be resolved" 404 "$code" || FAIL=$((FAIL+1))
expect_json_field "missing restore resolution error" error.code RESTORE_RECONCILIATION_NOT_FOUND || FAIL=$((FAIL+1))
if [ -n "$DB_DUMP_PATH" ] && [ -n "$FILE_SNAPSHOT_PATH" ]; then
  body="{\"database_dump_path\":\"$DB_DUMP_PATH\",\"file_snapshot_path\":\"$FILE_SNAPSHOT_PATH\"}"
  code=$(auth_post_json "$ADMIN_TOKEN" /api/admin/backup/restore "$body")
  expect_code "admin restore accepts real artifacts" 200 "$code" || FAIL=$((FAIL+1))
  expect_json_field "admin restore status" status Restored || FAIL=$((FAIL+1))
  expect_json_nonempty "admin restore id" id || FAIL=$((FAIL+1))
fi

code=$(auth_get "$ADMIN_TOKEN" /api/notifications)
expect_code "admin notifications" 200 "$code" || FAIL=$((FAIL+1))

: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
code=$(auth_get "$EDITOR_TOKEN" /api/admin/config)
expect_code "editor cannot list admin config" 403 "$code" || FAIL=$((FAIL+1))
code=$(auth_patch_json "$EDITOR_TOKEN" /api/admin/config/backup.schedule_enabled '{"value":"true"}')
expect_code "editor cannot manage backup schedule" 403 "$code" || FAIL=$((FAIL+1))
code=$(auth_put_json "$EDITOR_TOKEN" /api/admin/workflow-statuses "$workflow_body")
expect_code "editor cannot replace workflow statuses" 403 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$EDITOR_TOKEN" /api/audit-logs)
expect_code "editor cannot list audit logs" 403 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$EDITOR_TOKEN" /api/admin/backup/jobs)
expect_code "editor cannot list backup jobs" 403 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
