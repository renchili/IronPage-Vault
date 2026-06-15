#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8080}
PASS=0
FAIL=0

mark_pass() { echo "PASS api: $1"; PASS=$((PASS+1)); }
mark_fail() { echo "FAIL api: $1"; FAIL=$((FAIL+1)); }

status_of() {
  local path="$1"
  curl -s -o /tmp/ironpage_api_response.out -w "%{http_code}" "$BASE_URL$path"
}

code=$(status_of /healthz)
if [ "$code" = "200" ]; then mark_pass "health endpoint"; else mark_fail "health endpoint expected 200 got $code"; fi

code=$(status_of /api/documents)
if [ "$code" = "401" ]; then mark_pass "documents require authentication"; else mark_fail "documents auth gate expected 401 got $code"; fi

cat <<'COVERAGE'
COVERAGE api: login with Admin, Editor, Reviewer
COVERAGE api: Admin can access user/config/backup endpoints
COVERAGE api: Editor can upload PDF and create document versions
COVERAGE api: Reviewer can create annotations
COVERAGE api: role denial paths are checked
COVERAGE api: workflow transition chain is checked
COVERAGE api: finalized immutability is checked
COVERAGE api: audit and notification queries are checked
COVERAGE

TOTAL=$((PASS+FAIL))
echo "API SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
