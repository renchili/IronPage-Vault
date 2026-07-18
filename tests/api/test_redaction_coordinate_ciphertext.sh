#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"

DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
if [ ! -s "$DOC_ID_FILE" ]; then
  echo "FAIL api: redaction coordinate ciphertext test requires uploaded document id"
  exit 1
fi
DOC_ID="$(cat "$DOC_ID_FILE")"

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions" '{"page":1,"x":11,"y":12,"width":13,"height":14,"reason":"ciphertext-check"}')
expect_code "create encrypted-coordinate redaction" 201 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions")
expect_code "list redactions hides plaintext coordinates" 200 "$code" || FAIL=$((FAIL+1))
if grep -q '"x":\|"y":\|"width":\|"height":\|"reason":' "$BODY"; then
  echo "FAIL api: redaction list exposed sensitive coordinate/reason fields"
  cat "$BODY"
  FAIL=$((FAIL+1))
else
  echo "PASS api: redaction list hides sensitive coordinate/reason fields"
fi

exit "$FAIL"
