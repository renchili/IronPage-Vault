#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"

DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
if [ ! -s "$DOC_ID_FILE" ]; then
  echo "FAIL api: compare self contained requires uploaded document id"
  exit 1
fi
DOC_ID="$(cat "$DOC_ID_FILE")"

code=$(auth_get "$EDITOR_TOKEN" "/api/documents/$DOC_ID/versions")
expect_code "list versions for compare setup" 200 "$code" || FAIL=$((FAIL+1))
LEFT_VERSION_ID="$(json_field data[0].id)"
if [ -z "$LEFT_VERSION_ID" ]; then echo "FAIL api: left version id missing"; exit 1; fi

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"CMP-","suffix":"","zero_padding":3,"start":0}')
expect_code "create second version for compare" 201 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" "/api/documents/$DOC_ID/versions")
expect_code "list versions after compare setup" 200 "$code" || FAIL=$((FAIL+1))
RIGHT_VERSION_ID="$(json_field data[0].id)"
if [ -z "$RIGHT_VERSION_ID" ]; then echo "FAIL api: right version id missing"; exit 1; fi

body="{\"left_version_id\":\"$LEFT_VERSION_ID\",\"right_version_id\":\"$RIGHT_VERSION_ID\"}"
code=$(auth_post_json "$EDITOR_TOKEN" /api/documents/compare "$body")
expect_code "compare returns text bbox result" 200 "$code" || FAIL=$((FAIL+1))
expect_json_field "compare kind" data.comparison_kind text_bbox || FAIL=$((FAIL+1))
expect_json_field "compare text supported" data.text_diff_supported true || FAIL=$((FAIL+1))
expect_json_field "compare bbox supported" data.bbox_supported true || FAIL=$((FAIL+1))

exit "$FAIL"
