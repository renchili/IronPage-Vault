#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

run_suite() {
  local name="$1"
  local script="$2"
  if [ ! -f "$script" ]; then
    echo "FAIL suite: $name missing $script"
    FAIL=$((FAIL+1))
    return
  fi
  chmod +x "$script"
  if "$script"; then
    echo "PASS suite: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL suite: $name"
    FAIL=$((FAIL+1))
  fi
}

run_cmd_suite() {
  local name="$1"
  shift
  if "$@"; then
    echo "PASS suite: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL suite: $name"
    FAIL=$((FAIL+1))
  fi
}

run_suite "unit-structure" "unit_tests/test_rules.sh"
run_cmd_suite "go-unit" go test ./...
run_suite "api" "API_tests/test_api_flow.sh"

TOTAL=$((PASS+FAIL))
echo "TEST SUMMARY total_suites=$TOTAL passed_suites=$PASS failed_suites=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
