#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

run_case() {
  local name="$1"
  local script="$2"
  chmod +x "$script"
  if "$script"; then
    echo "PASS api-suite: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL api-suite: $name"
    FAIL=$((FAIL+1))
  fi
}

run_case "auth and RBAC" "API_tests/test_auth_rbac.sh"
run_case "documents and review" "API_tests/test_documents_review.sh"
run_case "audit notifications backup" "API_tests/test_audit_notify_backup.sh"

TOTAL=$((PASS+FAIL))
echo "API SUMMARY total_suites=$TOTAL passed_suites=$PASS failed_suites=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
