#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
SAMPLE_PDF=${SAMPLE_PDF:-testdata/pdfs/sample_contract.pdf}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents/batch" \
  -X POST \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "X-Request-ID: $(reqid)" \
  -H "X-Request-Timestamp: $(ts)" \
  -F "files=@$SAMPLE_PDF" \
  -F "files=@$SAMPLE_PDF")
expect_code "editor batch upload" 201 "$code" || FAIL=$((FAIL+1))

COUNT=$(json_field count)
[ "$COUNT" = "2" ] || { echo "FAIL api: expected batch count 2 got=$COUNT"; FAIL=$((FAIL+1)); }

exit "$FAIL"
