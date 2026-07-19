#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

. tests/api/lib.sh

: "${SEED_ADMIN_PASSWORD:?SEED_ADMIN_PASSWORD is required for acceptance tests}"
: "${SEED_EDITOR_PASSWORD:?SEED_EDITOR_PASSWORD is required for acceptance tests}"
: "${SEED_REVIEWER_PASSWORD:?SEED_REVIEWER_PASSWORD is required for acceptance tests}"

login_and_store() {
  local user="$1"
  local secret="$2"
  local out="$3"

  code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/auth/login" \
    -H 'Content-Type: application/json' \
    -H "X-Request-ID: $(reqid)" \
    -d "{\"username\":\"$user\",\"password\":\"$secret\"}")

  expect_code "login $user" 200 "$code" || exit 1
  json_field token > "$out"
}

login_and_store admin "$SEED_ADMIN_PASSWORD" /tmp/ironpage_admin_token.out
login_and_store editor "$SEED_EDITOR_PASSWORD" /tmp/ironpage_editor_token.out
login_and_store reviewer "$SEED_REVIEWER_PASSWORD" /tmp/ironpage_reviewer_token.out

export ADMIN_TOKEN="$(cat /tmp/ironpage_admin_token.out)"
export EDITOR_TOKEN="$(cat /tmp/ironpage_editor_token.out)"
export REVIEWER_TOKEN="$(cat /tmp/ironpage_reviewer_token.out)"

run_case() {
  local name="$1"
  local script="$2"
  if bash "$script"; then
    echo "PASS api-suite: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL api-suite: $name"
    FAIL=$((FAIL+1))
  fi
}

run_case "auth and RBAC" "tests/api/test_auth_rbac.sh"
run_case "documents and review" "tests/api/test_documents_review.sh"
run_case "audit notifications backup" "tests/api/test_audit_notify_backup.sh"

TOTAL=$((PASS+FAIL))
echo "API SUMMARY total_suites=$TOTAL passed_suites=$PASS failed_suites=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
