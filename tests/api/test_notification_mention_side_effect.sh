#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"
SAMPLE_PDF=${SAMPLE_PDF:-testdata/pdfs/sample_contract.pdf}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -X POST -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)" -F "title=Mention Side Effect Probe" -F "file=@$SAMPLE_PDF")
expect_code "upload mention probe" 201 "$code" || FAIL=$((FAIL+1))
DOC_ID="$(json_field data.id)"
if [ -z "$DOC_ID" ]; then echo "FAIL api: mention probe doc id missing"; exit 1; fi

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Under Review"}')
expect_code "mention probe to under review" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/annotations" '{"type":"Sticky note","page":1,"x":10,"y":10,"width":10,"height":10,"comment":"Please check @editor","disposition":"Needs Discussion"}')
expect_code "create mention annotation" 201 "$code" || FAIL=$((FAIL+1))
code=$(auth_get "$EDITOR_TOKEN" /api/notifications)
expect_code "editor notifications after mention" 200 "$code" || FAIL=$((FAIL+1))
if grep -q "mention" "$BODY" || grep -q "Please check" "$BODY" || grep -q "$DOC_ID" "$BODY"; then
  echo "PASS api: mention notification side effect visible"
else
  echo "FAIL api: mention notification side effect missing"
  cat "$BODY"
  FAIL=$((FAIL+1))
fi

exit "$FAIL"
