#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh

FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"

DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
if [ ! -s "$DOC_ID_FILE" ]; then
  echo "FAIL api: missing document id file $DOC_ID_FILE"
  exit 1
fi

DOC_ID="$(cat "$DOC_ID_FILE")"

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Under Review"}')
expect_code "editor transitions draft to under review" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$REVIEWER_TOKEN" "/api/documents/$DOC_ID")
expect_code "reviewer reads document after under review" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Finalized"}')
expect_code "reviewer cannot skip workflow to finalized" 400 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
