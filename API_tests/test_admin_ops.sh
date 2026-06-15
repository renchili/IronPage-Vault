#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${ADMIN_TOKEN:?set ADMIN_TOKEN}"

code=$(auth_get "$ADMIN_TOKEN" /api/admin/config)
expect_code "admin config list" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" /api/admin/workflow-statuses)
expect_code "admin workflow statuses" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" /api/admin/notification-templates)
expect_code "admin notification templates" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" /api/audit-logs)
expect_code "audit log list" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" '/api/audit-logs?action_type=DOCUMENT_UPLOAD')
expect_code "audit log action filter" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" /api/admin/backup/jobs)
expect_code "admin backup jobs" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" /api/notifications)
expect_code "admin notifications" 200 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
