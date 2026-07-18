#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh
FAIL=0

: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"

DOC_ID=$(cat "${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}")

code=$(auth_get "$EDITOR_TOKEN" "/api/documents/$DOC_ID/versions")
expect_code "list versions before compare" 200 "$code" || FAIL=$((FAIL+1))

LEFT=$(json_field data[0].id)
RIGHT="$LEFT"

code=$(auth_post_json "$EDITOR_TOKEN" /api/documents/compare "{\"left_version_id\":\"$LEFT\",\"right_version_id\":\"$RIGHT\"}")
expect_code "compare identical versions" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" /api/documents/compare '{"left_version_id":"missing-left","right_version_id":"missing-right"}')
expect_code "compare missing versions rejected" 404 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
