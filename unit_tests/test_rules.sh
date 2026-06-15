#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

check() {
  local name="$1"
  local cmd="$2"
  if bash -lc "$cmd" >/tmp/ironpage_unit.out 2>&1; then
    echo "PASS unit: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL unit: $name"
    cat /tmp/ironpage_unit.out
    FAIL=$((FAIL+1))
  fi
}

check "metadata exists" "test -f metadata.json"
check "agent rules exist" "test -f AGENT.md"
check "compose exists" "test -f docker-compose.yml"
check "migration exists" "test -f migrations/001_schema.sql"
check "sample pdf exists" "test -f testdata/pdfs/sample_contract.pdf"
check "pdf header valid" "head -c 5 testdata/pdfs/sample_contract.pdf | grep -q '%PDF-'"
check "roles in schema" "grep -q 'Admin' migrations/001_schema.sql && grep -q 'Editor' migrations/001_schema.sql && grep -q 'Reviewer' migrations/001_schema.sql"
check "workflow in schema" "grep -q 'Finalized' migrations/001_schema.sql && grep -q 'Under Review' migrations/001_schema.sql"
check "api tests directory" "test -d API_tests"

TOTAL=$((PASS+FAIL))
echo "UNIT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
