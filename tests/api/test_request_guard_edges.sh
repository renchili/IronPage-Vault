#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh

FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"

check_error() {
  expect_json_field "$1 code" error.code "$2" || return 1
  expect_json_nonempty "$1 request id" error.request_id || return 1
  expect_json_nonempty "$1 timestamp" error.timestamp
}

request_with_guard() {
  local token="$1" request_id="$2" timestamp="$3"
  curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" \
    -H "Authorization: Bearer $token" \
    -H "X-Request-ID: $request_id" \
    -H "X-Request-Timestamp: $timestamp"
}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)")
expect_code "missing timestamp" 400 "$code" || FAIL=$((FAIL+1))
check_error "missing timestamp" REQUEST_TIMESTAMP_REQUIRED || FAIL=$((FAIL+1))

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: malformed")
expect_code "malformed timestamp" 400 "$code" || FAIL=$((FAIL+1))
check_error "malformed timestamp" REQUEST_TIMESTAMP_INVALID || FAIL=$((FAIL+1))

fresh_timestamp=$(date -u -d '59 seconds ago' +%Y-%m-%dT%H:%M:%SZ)
code=$(request_with_guard "$EDITOR_TOKEN" "$(reqid)" "$fresh_timestamp")
expect_code "59-second-old timestamp is accepted" 200 "$code" || FAIL=$((FAIL+1))

expired_timestamp=$(date -u -d '61 seconds ago' +%Y-%m-%dT%H:%M:%SZ)
code=$(request_with_guard "$EDITOR_TOKEN" "$(reqid)" "$expired_timestamp")
expect_code "61-second-old timestamp is rejected" 401 "$code" || FAIL=$((FAIL+1))
check_error "expired timestamp" REQUEST_EXPIRED || FAIL=$((FAIL+1))

future_timestamp=$(date -u -d '61 seconds' +%Y-%m-%dT%H:%M:%SZ)
code=$(request_with_guard "$EDITOR_TOKEN" "$(reqid)" "$future_timestamp")
expect_code "61-second-future timestamp is rejected" 401 "$code" || FAIL=$((FAIL+1))
check_error "future timestamp" REQUEST_EXPIRED || FAIL=$((FAIL+1))

replay_id=$(reqid)
code=$(request_with_guard "$EDITOR_TOKEN" "$replay_id" "$(ts)")
expect_code "first request id use succeeds" 200 "$code" || FAIL=$((FAIL+1))
code=$(request_with_guard "$EDITOR_TOKEN" "$replay_id" "$(ts)")
expect_code "same JWT rejects duplicate request id" 409 "$code" || FAIL=$((FAIL+1))
check_error "duplicate request id" REPLAY_DETECTED || FAIL=$((FAIL+1))

code=$(request_with_guard "$REVIEWER_TOKEN" "$replay_id" "$(ts)")
expect_code "request id scope permits a different JWT" 200 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
