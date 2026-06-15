#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"
SAMPLE_PDF=${SAMPLE_PDF:-testdata/pdfs/sample_contract.pdf}
OUT=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -X POST -H "Authorization: Bearer $REVIEWER_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)")
expect_code "reviewer upload role check" 403 "$code" || FAIL=$((FAIL+1))

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -X POST -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)" -F "title=API Contract" -F "file=@$SAMPLE_PDF")
expect_code "editor upload" 201 "$code" || FAIL=$((FAIL+1))
DOC_ID=$(json_field data.id)
[ -n "$DOC_ID" ] && echo "$DOC_ID" > "$OUT" || { echo "FAIL api: document id missing"; FAIL=$((FAIL+1)); }

code=$(auth_get "$REVIEWER_TOKEN" /api/documents/$DOC_ID)
expect_code "reviewer read document" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_get "$REVIEWER_TOKEN" /api/documents/$DOC_ID/versions)
expect_code "reviewer list versions" 200 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
