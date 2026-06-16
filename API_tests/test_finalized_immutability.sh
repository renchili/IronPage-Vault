#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
if [ ! -s "$DOC_ID_FILE" ]; then
  echo "SKIP finalized immutability requires uploaded doc id"
  exit 0
fi
DOC_ID="$(cat "$DOC_ID_FILE")"
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/finalize" '{}')
expect_code "finalize route reachable" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"X","start":1}')
expect_code "finalized document rejects bates mutation" 409 "$code" || FAIL=$((FAIL+1))
exit "$FAIL"
