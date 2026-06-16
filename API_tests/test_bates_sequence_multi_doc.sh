#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
SAMPLE_PDF=${SAMPLE_PDF:-testdata/pdfs/sample_contract.pdf}

upload_doc() {
  local title="$1"
  local code
  code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" \
    -X POST \
    -H "Authorization: Bearer $EDITOR_TOKEN" \
    -H "X-Request-ID: $(reqid)" \
    -H "X-Request-Timestamp: $(ts)" \
    -F "title=$title" \
    -F "file=@$SAMPLE_PDF")
  expect_code "upload $title" 201 "$code" || return 1
  json_field data.id
}

DOC1="$(upload_doc "Bates Sequence Probe 1")" || exit 1
DOC2="$(upload_doc "Bates Sequence Probe 2")" || exit 1

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC1/bates" '{"prefix":"SEQ-","suffix":"","zero_padding":3,"start":0}')
expect_code "bates first sequence allocation" 201 "$code" || FAIL=$((FAIL+1))
START1="$(json_field start_number)"
VERSION1="$(json_field version)"
[ -n "$VERSION1" ] || { echo "FAIL api: first bates version missing"; FAIL=$((FAIL+1)); }

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC2/bates" '{"prefix":"SEQ-","suffix":"","zero_padding":3,"start":0}')
expect_code "bates second sequence allocation" 201 "$code" || FAIL=$((FAIL+1))
START2="$(json_field start_number)"
VERSION2="$(json_field version)"
[ -n "$VERSION2" ] || { echo "FAIL api: second bates version missing"; FAIL=$((FAIL+1)); }

python3 - "$START1" "$START2" <<'PY'
import sys
try:
    a=int(sys.argv[1]); b=int(sys.argv[2])
except Exception:
    print("FAIL api: start_number missing or not numeric")
    sys.exit(1)
if b <= a:
    print(f"FAIL api: second Bates start not greater than first: {a} -> {b}")
    sys.exit(1)
print(f"PASS api: Bates sequence increases across documents: {a} -> {b}")
PY
if [ $? -ne 0 ]; then FAIL=$((FAIL+1)); fi

exit "$FAIL"
