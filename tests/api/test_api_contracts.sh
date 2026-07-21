#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh

FAIL=0
: "${ADMIN_TOKEN:?set ADMIN_TOKEN}"
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
DOC_ID=$(cat "$DOC_ID_FILE" 2>/dev/null || true)
if [ -z "$DOC_ID" ]; then
  echo "FAIL api: pagination contract document id missing"
  exit 1
fi

check_error() {
  local name="$1" expected="$2"
  expect_json_field "$name code" error.code "$expected" || return 1
  expect_json_nonempty "$name request id" error.request_id || return 1
  expect_json_nonempty "$name timestamp" error.timestamp
}

check_page() {
  local name="$1" page="$2" size="$3"
  expect_json_field "$name page" page "$page" || return 1
  expect_json_field "$name size" page_size "$size"
}

check_collection_request() {
  local name="$1" token="$2" path="$3" page="$4" size="$5"
  local code
  code=$(auth_get "$token" "$path")
  expect_code "$name" 200 "$code" || return 1
  check_page "$name" "$page" "$size"
}

check_collection_pagination() {
  local name="$1" token="$2" path="$3"
  check_collection_request "$name default" "$token" "$path" 1 25 || FAIL=$((FAIL+1))
  check_collection_request "$name minimum" "$token" "$path?page=-1&page_size=0" 1 25 || FAIL=$((FAIL+1))
  check_collection_request "$name boundary" "$token" "$path?page=2&page_size=1" 2 1 || FAIL=$((FAIL+1))
  check_collection_request "$name maximum" "$token" "$path?page=2&page_size=101" 2 100 || FAIL=$((FAIL+1))
}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -H "X-Request-ID: $(reqid)")
expect_code "documents requires auth" 401 "$code" || FAIL=$((FAIL+1))
check_error "documents auth envelope" AUTH_REQUIRED || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/admin/config)
expect_code "admin route denied" 403 "$code" || FAIL=$((FAIL+1))
check_error "admin denial envelope" FORBIDDEN || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/documents/does-not-exist)
expect_code "missing document" 404 "$code" || FAIL=$((FAIL+1))
check_error "missing document envelope" DOCUMENT_NOT_FOUND || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/not-a-real-route)
expect_code "unknown route" 404 "$code" || FAIL=$((FAIL+1))
check_error "unknown route envelope" NOT_FOUND || FAIL=$((FAIL+1))

code=$(curl -s -o "$BODY" -w "%{http_code}" -X POST "$BASE_URL/healthz" -H "X-Request-ID: $(reqid)")
expect_code "invalid method" 405 "$code" || FAIL=$((FAIL+1))
check_error "invalid method envelope" METHOD_NOT_ALLOWED || FAIL=$((FAIL+1))

check_collection_pagination "documents" "$EDITOR_TOKEN" /api/documents
check_collection_pagination "admin users" "$ADMIN_TOKEN" /api/admin/users
check_collection_pagination "admin config" "$ADMIN_TOKEN" /api/admin/config
check_collection_pagination "workflow statuses" "$ADMIN_TOKEN" /api/admin/workflow-statuses
check_collection_pagination "notification templates" "$ADMIN_TOKEN" /api/admin/notification-templates
check_collection_pagination "backup jobs" "$ADMIN_TOKEN" /api/admin/backup/jobs
check_collection_pagination "document versions" "$EDITOR_TOKEN" "/api/documents/$DOC_ID/versions"
check_collection_pagination "document redactions" "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions"
check_collection_pagination "document annotations" "$EDITOR_TOKEN" "/api/documents/$DOC_ID/annotations"
check_collection_pagination "notifications" "$EDITOR_TOKEN" /api/notifications
check_collection_pagination "audit logs" "$ADMIN_TOKEN" /api/audit-logs

exit "$FAIL"
