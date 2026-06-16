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
check "workflow unit test exists" "test -f internal/app/workflow_test.go"
check "pdf unit test exists" "test -f internal/app/pdf_test.go"
check "domain rule unit test exists" "test -f internal/app/rules_test.go"
check "domain rule helper exists" "test -f internal/app/rules.go"
check "crypto helper exists" "test -f internal/app/crypto.go"
check "crypto unit test exists" "test -f internal/app/crypto_test.go"
check "access helper exists" "test -f internal/app/access.go"
check "access unit test exists" "test -f internal/app/access_test.go"
check "mention helper exists" "test -f internal/app/mentions.go"
check "mention unit test exists" "test -f internal/app/mentions_test.go"
check "manual backend test UI exists" "test -f public/manual-test.html"
check "public swagger yaml exists" "test -f public/swagger.yaml"
check "ci workflow exists" "test -f .github/workflows/ci.yml"
check "strict pdf entrypoints exist" "grep -q 'RewritePDFWithRedactionsStrict' internal/platform/pdf_strict.go && grep -q 'RewritePDFWithBatesStrict' internal/platform/pdf_strict.go"
check "strict backup entrypoints exist" "grep -q 'RunBackupArtifactsStrict' internal/platform/backup_strict.go && grep -q 'RunRestoreArtifactsStrict' internal/platform/backup_strict.go"
check "redaction coordinates stored encrypted only" "grep -q 'x_ciphertext' internal/app/review.go && ! grep -q 'req.X, req.Y, req.Width, req.Height' internal/app/review.go"
check "redaction burn-in uses encrypted coordinates" "grep -q 'EncryptedRedactionRegions' internal/app/coordinate_crypto.go"
check "redaction list hides coordinate plaintext" "! grep -q 'SELECT id,document_id,page,x,y,width,height,reason' internal/app/review.go"
check "restore strict failure path exists" "grep -q 'RunRestoreArtifactsStrict' internal/app/restore.go"
check "self-contained compare test exists" "test -f API_tests/test_compare_self_contained.sh"

TOTAL=$((PASS+FAIL))
echo "UNIT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
