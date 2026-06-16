#!/usr/bin/env bash
set -u -o pipefail
. API_tests/lib.sh
FAIL=0
: "${ADMIN_TOKEN:?set ADMIN_TOKEN}"
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"
: "${REVIEWER_TOKEN:?set REVIEWER_TOKEN}"

code=$(auth_get "$EDITOR_TOKEN" /api/audit-logs)
expect_code "editor cannot list audit logs" 403 "$code" || FAIL=$((FAIL+1))

DOC_ID_FILE=${DOC_ID_FILE:-/tmp/ironpage_doc_id.out}
if [ -s "$DOC_ID_FILE" ]; then
  DOC_ID="$(cat "$DOC_ID_FILE")"
  code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/workflow/transition" '{"status":"Under Review"}')
  expect_code "workflow transition covered" 200 "$code" || FAIL=$((FAIL+1))
  code=$(auth_post_json "$REVIEWER_TOKEN" "/api/documents/$DOC_ID/redactions" '{"page":1,"x":1,"y":1,"width":10,"height":10,"reason":"blocked"}')
  expect_code "reviewer cannot create redaction" 403 "$code" || FAIL=$((FAIL+1))
  code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"IP-","suffix":"","zero_padding":4,"start":0}')
  expect_code "bates creates visible version" 201 "$code" || FAIL=$((FAIL+1))
fi

code=$(auth_post_json "$ADMIN_TOKEN" /api/admin/backup/run '{}')
expect_code "admin backup run" 201 "$code" || FAIL=$((FAIL+1))
code=$(auth_post_json "$EDITOR_TOKEN" /api/notifications/not-found/read '{}')
expect_code "missing notification read is 404" 404 "$code" || FAIL=$((FAIL+1))
exit "$FAIL"
