#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${ADMIN_TOKEN:?set ADMIN_TOKEN}"
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"
DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
if [ -s "$DOC_ID_FILE" ]; then
  DOC_ID="$(cat "$DOC_ID_FILE")"
  code=$(auth_post_json "$ADMIN_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"X","start":1}')
  expect_code "admin is not editor for Bates mutation" 403 "$code" || FAIL=$((FAIL+1))
  code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/redactions" '{"page":1,"x":1,"y":1,"width":10,"height":10,"reason":"x"}')
  expect_code "reviewer cannot propose redaction" 403 "$code" || FAIL=$((FAIL+1))
fi
exit "$FAIL"
