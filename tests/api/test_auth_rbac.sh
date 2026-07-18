#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${ADMIN_TOKEN:?set ADMIN_TOKEN}"
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/healthz")
expect_code "health" 200 "$code" || FAIL=$((FAIL+1))

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents")
expect_code "anonymous documents request" 401 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/auth/me)
expect_code "me endpoint" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$ADMIN_TOKEN" /api/admin/users)
expect_code "admin user list" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" /api/admin/users)
expect_code "editor user list denial" 403 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$REVIEWER_TOKEN" /api/admin/config)
expect_code "reviewer config denial" 403 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
