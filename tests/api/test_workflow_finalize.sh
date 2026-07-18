#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"
DOC_ID=$(cat "${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}")

code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Approved"}')
expect_code "invalid workflow transition rejected" 400 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Under Review"}')
expect_code "editor moves draft to under review" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Redaction Pending"}')
expect_code "reviewer moves to redaction pending" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Approved"}')
expect_code "reviewer moves to approved" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/finalize" '{}')
expect_code "editor finalizes approved document" 200 "$code" || FAIL=$((FAIL+1))

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Finalized"}')
expect_code "finalized document rejects transition" 403 "$code" || FAIL=$((FAIL+1))

exit "$FAIL"
