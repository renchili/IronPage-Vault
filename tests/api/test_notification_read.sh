#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
code=$(auth_post_json "$EDITOR_TOKEN" "/api/notifications/not_missing/read" '{}')
expect_code "missing notification read returns 404" 404 "$code" || FAIL=$((FAIL+1))
exit "$FAIL"
