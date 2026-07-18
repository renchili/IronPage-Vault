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
bad = [r for r in rows if r.get("action_type") != "DOCUMENT_UPLOAD"]
if bad:
    print("FAIL api: audit action filter returned non DOCUMENT_UPLOAD rows")
    print(json.dumps(bad[:3], indent=2))
    sys.exit(1)
print("PASS api: audit action filter rows match")
PY
}

code=$(auth_get "$ADMIN_TOKEN" /api/admin/config)
expect_code "admin config list" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$ADMIN_TOKEN" /api/admin/workflow-statuses)
expect_code "admin workflow statuses" 200 "$code" || FAIL=$((FAIL+1))
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
# Artifact paths are container-internal during Docker acceptance; their
# existence is verified by the restore request below rather than host-side test.

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

if [ -n "$DB_DUMP_PATH" ] && [ -n "$FILE_SNAPSHOT_PATH" ]; then
  body="{\"database_dump_path\":\"$DB_DUMP_PATH\",\"file_snapshot_path\":\"$FILE_SNAPSHOT_PATH\"}"
  code=$(auth_post_json "$ADMIN_TOKEN" /api/admin/backup/restore "$body")
  expect_code "admin restore accepts real artifacts" 200 "$code" || FAIL=$((FAIL+1))
  expect_json_field "admin restore status" status Restored || FAIL=$((FAIL+1))
fi

code=$(auth_get "$ADMIN_TOKEN" /api/notifications)
expect_code "admin notifications" 200 "$code" || FAIL=$((FAIL+1))

: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
code=$(auth_get "$EDITOR_TOKEN" /api/admin/config)
expect_code "editor cannot list admin config" 403 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$EDITOR_TOKEN" /api/audit-logs)
expect_code "editor cannot list audit logs" 403 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$EDITOR_TOKEN" /api/admin/backup/jobs)
expect_code "editor cannot list backup jobs" 403 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
