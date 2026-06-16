#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${LEFT_VERSION_ID:?set LEFT_VERSION_ID}"
: "${RIGHT_VERSION_ID:?set RIGHT_VERSION_ID}"
body="{\"left_version_id\":\"$LEFT_VERSION_ID\",\"right_version_id\":\"$RIGHT_VERSION_ID\"}"
code=$(auth_post_json "$EDITOR_TOKEN" /api/documents/compare "$body")
expect_code "compare returns text bbox result" 200 "$code" || FAIL=$((FAIL+1))
exit "$FAIL"
