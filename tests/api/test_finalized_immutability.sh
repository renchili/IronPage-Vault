#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh
FAIL=0
: "${ADMIN_TOKEN:?set ADMIN_TOKEN}"
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"
SAMPLE_PDF=${SAMPLE_PDF:-testdata/pdfs/sample_contract.pdf}

json_data_count() {
  python3 - "$BODY" <<'PY'
import json, sys
with open(sys.argv[1]) as f:
    body = json.load(f)
print(len(body.get("data", [])))
PY
}

read_count() {
  local token="$1" path="$2" name="$3"
  local code
  code=$(auth_get "$token" "$path")
  expect_code "$name" 200 "$code" || return 1
  json_data_count
}

expect_finalized_denial() {
  local name="$1" code="$2"
  expect_code "$name" 409 "$code" || return 1
  expect_json_field "$name error" error.code DOCUMENT_FINALIZED
}

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -X POST -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)" -F "title=Finalized Immutability Probe" -F "file=@$SAMPLE_PDF")
expect_code "upload finalized immutability probe" 201 "$code" || FAIL=$((FAIL+1))
DOC_ID="$(json_field data.id)"
if [ -z "$DOC_ID" ]; then echo "FAIL api: finalized probe doc id missing"; exit 1; fi

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions" '{"page":1,"x":10,"y":10,"width":20,"height":20,"reason":"staged before finalization"}')
expect_code "stage redaction before finalization" 201 "$code" || FAIL=$((FAIL+1))
REDACTION_ID="$(json_field id)"

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Under Review"}')
expect_code "probe draft to under review" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/annotations" '{"type":"Comment","page":1,"comment":"annotation before finalization","disposition":"Needs Discussion"}')
expect_code "create annotation before finalization" 201 "$code" || FAIL=$((FAIL+1))
ANNOTATION_ID="$(json_field id)"

code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Redaction Pending"}')
expect_code "probe under review to redaction pending" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Approved"}')
expect_code "probe redaction pending to approved" 200 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/finalize" '{}')
expect_code "finalize approved probe" 200 "$code" || FAIL=$((FAIL+1))

BASE_VERSIONS=$(read_count "$EDITOR_TOKEN" "/api/documents/$DOC_ID/versions?page_size=100" "read finalized version baseline") || FAIL=$((FAIL+1))
BASE_REDACTIONS=$(read_count "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions?page_size=100" "read finalized redaction baseline") || FAIL=$((FAIL+1))
BASE_ANNOTATIONS=$(read_count "$EDITOR_TOKEN" "/api/documents/$DOC_ID/annotations?page_size=100" "read finalized annotation baseline") || FAIL=$((FAIL+1))
BASE_AUDITS=$(read_count "$ADMIN_TOKEN" "/api/audit-logs?document_id=$DOC_ID&page_size=100" "read finalized audit baseline") || FAIL=$((FAIL+1))
BASE_NOTIFICATIONS=$(read_count "$EDITOR_TOKEN" "/api/notifications?page_size=100" "read finalized notification baseline") || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/rollback" '{"version":1}')
expect_finalized_denial "finalized document rejects rollback" "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions" '{"page":1,"x":10,"y":10,"width":20,"height":20,"reason":"must be rejected"}')
expect_finalized_denial "finalized document rejects redaction proposal" "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions/$REDACTION_ID/confirm" '{}')
expect_finalized_denial "finalized document rejects redaction confirmation" "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/annotations" '{"type":"Comment","page":1,"comment":"must be rejected","disposition":"Needs Discussion"}')
expect_finalized_denial "finalized document rejects annotation creation" "$code" || FAIL=$((FAIL+1))

code=$(auth_patch_json "$REVIEWER_TOKEN" "/api/annotations/$ANNOTATION_ID/disposition" '{"disposition":"Approved"}')
expect_finalized_denial "finalized document rejects annotation disposition" "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"X","start":1}')
expect_finalized_denial "finalized document rejects Bates mutation" "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Finalized"}')
expect_finalized_denial "finalized document rejects workflow transition" "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/finalize" '{}')
expect_finalized_denial "finalized document rejects repeated finalization" "$code" || FAIL=$((FAIL+1))

AFTER_VERSIONS=$(read_count "$EDITOR_TOKEN" "/api/documents/$DOC_ID/versions?page_size=100" "read version count after denials") || FAIL=$((FAIL+1))
AFTER_REDACTIONS=$(read_count "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions?page_size=100" "read redaction count after denials") || FAIL=$((FAIL+1))
AFTER_ANNOTATIONS=$(read_count "$EDITOR_TOKEN" "/api/documents/$DOC_ID/annotations?page_size=100" "read annotation count after denials") || FAIL=$((FAIL+1))
AFTER_AUDITS=$(read_count "$ADMIN_TOKEN" "/api/audit-logs?document_id=$DOC_ID&page_size=100" "read audit count after denials") || FAIL=$((FAIL+1))
AFTER_NOTIFICATIONS=$(read_count "$EDITOR_TOKEN" "/api/notifications?page_size=100" "read notification count after denials") || FAIL=$((FAIL+1))

for item in \
  "versions:$BASE_VERSIONS:$AFTER_VERSIONS" \
  "redactions:$BASE_REDACTIONS:$AFTER_REDACTIONS" \
  "annotations:$BASE_ANNOTATIONS:$AFTER_ANNOTATIONS" \
  "audits:$BASE_AUDITS:$AFTER_AUDITS" \
  "notifications:$BASE_NOTIFICATIONS:$AFTER_NOTIFICATIONS"; do
  name=${item%%:*}
  rest=${item#*:}
  before=${rest%%:*}
  after=${rest##*:}
  if [ "$before" != "$after" ]; then
    echo "FAIL api: finalized denial changed $name count before=$before after=$after"
    FAIL=$((FAIL+1))
  else
    echo "PASS api: finalized denials leave $name unchanged"
  fi
done

exit "$FAIL"
