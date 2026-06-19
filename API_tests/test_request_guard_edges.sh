#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh

FAIL=0

check_error() {
  expect_json_field "$1 code" error.code "$2" || return 1
  expect_json_nonempty "$1 request id" error.request_id || return 1
  expect_json_nonempty "$1 timestamp" error.timestamp
}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)")
expect_code "missing timestamp" 400 "$code" || FAIL=$((FAIL+1))
check_error "missing timestamp" REQUEST_TIMESTAMP_REQUIRED || FAIL=$((FAIL+1))

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: malformed")
expect_code "malformed timestamp" 400 "$code" || FAIL=$((FAIL+1))
check_error "malformed timestamp" REQUEST_TIMESTAMP_INVALID || FAIL=$((FAIL+1))

exit "$FAIL"
