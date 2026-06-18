#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh

FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"

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

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -H "X-Request-ID: $(reqid)")
expect_code "documents requires auth" 401 "$code" || FAIL=$((FAIL+1))
check_error "documents auth envelope" AUTH_REQUIRED || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/admin/config)
expect_code "admin route denied" 403 "$code" || FAIL=$((FAIL+1))
check_error "admin denial envelope" FORBIDDEN || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/documents/does-not-exist)
expect_code "missing document" 404 "$code" || FAIL=$((FAIL+1))
check_error "missing document envelope" DOCUMENT_NOT_FOUND || FAIL=$((FAIL+1))

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/not-a-real-route" -H "X-Request-ID: $(reqid)")
expect_code "unknown route" 404 "$code" || FAIL=$((FAIL+1))
check_error "unknown route envelope" NOT_FOUND || FAIL=$((FAIL+1))

code=$(curl -s -o "$BODY" -w "%{http_code}" -X POST "$BASE_URL/healthz" -H "X-Request-ID: $(reqid)")
expect_code "invalid method" 405 "$code" || FAIL=$((FAIL+1))
check_error "invalid method envelope" METHOD_NOT_ALLOWED || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/documents)
expect_code "default pagination" 200 "$code" || FAIL=$((FAIL+1))
check_page "default pagination" 1 25 || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" '/api/documents?page=-1&page_size=0')
expect_code "minimum pagination" 200 "$code" || FAIL=$((FAIL+1))
check_page "minimum pagination" 1 25 || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" '/api/documents?page=2&page_size=101')
expect_code "maximum pagination" 200 "$code" || FAIL=$((FAIL+1))
check_page "maximum pagination" 2 100 || FAIL=$((FAIL+1))

exit "$FAIL"
