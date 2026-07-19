#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"
SAMPLE_PDF=${SAMPLE_PDF:-testdata/pdfs/sample_contract.pdf}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -X POST -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)" -F "title=Finalized Immutability Probe" -F "file=@$SAMPLE_PDF")
expect_code "upload finalized immutability probe" 201 "$code" || FAIL=$((FAIL+1))
DOC_ID="$(json_field data.id)"
if [ -z "$DOC_ID" ]; then echo "FAIL api: finalized probe doc id missing"; exit 1; fi

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Under Review"}')
expect_code "probe draft to under review" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Redaction Pending"}')
expect_code "probe under review to redaction pending" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Approved"}')
expect_code "probe redaction pending to approved" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/finalize" '{}')
expect_code "finalize approved probe" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"X","start":1}')
expect_code "finalized document rejects bates mutation" 409 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
