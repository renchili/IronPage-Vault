#!/usr/bin/env bash
set -euo pipefail

FAIL=0

check() {
  local name="$1"
  local cmd="$2"
  if bash -lc "$cmd" >/tmp/ironpage_structure_rule.out 2>&1; then
    echo "PASS structure: $name"
  else
    echo "FAIL structure: $name"
    cat /tmp/ironpage_structure_rule.out
    FAIL=$((FAIL+1))
  fi
}

check "service pdf uses strict redaction only" "grep -q 'RewritePDFWithRedactionsStrict' internal/service/pdf.go"
check "service pdf uses strict bates only" "grep -q 'RewritePDFWithBatesStrict' internal/service/pdf.go"
check "compare legacy env dependency removed" "! grep -R 'LEFT_VERSION_ID' API_tests/test_compare_acceptance.sh API_tests/test_compare_self_contained.sh"
check "restore empty body no longer expected as 202" "! grep -R 'backup/restore.*202\\|admin restore route accepts request' API_tests"
check "finalized test walks approved state" "grep -q 'Redaction Pending' API_tests/test_finalized_immutability.sh && grep -q 'Approved' API_tests/test_finalized_immutability.sh"
check "pdf content test checks redacted text" "grep -q 'SECRET_NEVER_APPEAR' API_tests/test_pdf_content_acceptance.sh && grep -q 'pdftotext' API_tests/test_pdf_content_acceptance.sh"
check "bates content test checks label" "grep -q 'CNT-001' API_tests/test_pdf_content_acceptance.sh"
check "backup artifact existence checked" "grep -q 'dump artifact file missing' API_tests/test_admin_ops.sh && grep -q 'file artifact missing' API_tests/test_admin_ops.sh"
check "bates sequence multi doc test exists" "test -f API_tests/test_bates_sequence_multi_doc.sh"
check "strict dependency negative test exists" "test -f API_tests/test_strict_dependency_failures.sh"
check "mention side effect test exists" "test -f API_tests/test_notification_mention_side_effect.sh"

if [ "$FAIL" -ne 0 ]; then exit 1; fi
