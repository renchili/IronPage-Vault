#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
if [ ! -s "$DOC_ID_FILE" ]; then echo "FAIL api: missing document id file $DOC_ID_FILE"; exit 1; fi
DOC_ID="$(cat "$DOC_ID_FILE")"
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"IP-","suffix":"","zero_padding":4,"start_number":1}')
expect_code "editor applies Bates version" 201 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions" '{"page":1,"x":10,"y":10,"width":40,"height":12,"reason":"test"}')
expect_code "editor proposes redaction" 201 "$code" || FAIL=$((FAIL+1))
exit "$FAIL"
