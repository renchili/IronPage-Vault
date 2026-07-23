#!/usr/bin/env bash
set -euo pipefail

FAIL=0

check() {
  local name="$1" cmd="$2"
  if bash -lc "$cmd" >/tmp/ironpage_structure_rule.out 2>&1; then
    echo "PASS structure: $name"
  else
    echo "FAIL structure: $name"
    cat /tmp/ironpage_structure_rule.out
    FAIL=$((FAIL+1))
  fi
}

check "service PDF uses strict redaction only" "grep -q 'RewritePDFWithRedactionsStrict' internal/service/pdf.go"
check "service PDF uses strict Bates only" "grep -q 'RewritePDFWithBatesStrict' internal/service/pdf.go"
check "compare legacy environment dependency removed" "! grep -R 'LEFT_VERSION_ID' tests/api/test_compare_acceptance.sh tests/api/test_compare_self_contained.sh"
check "restore empty body is rejected" "! grep -R 'backup/restore.*202\|admin restore route accepts request' tests/api"
check "restore reapplies required schema before completion" "grep -q 'RunMigrations(a.db, a.cfg.MigrationsDir)' internal/app/restore.go && grep -q 'schema_migrations_applied' internal/app/restore.go"
check "finalized test walks approved state" "grep -q 'Redaction Pending' tests/api/test_finalized_immutability.sh && grep -q 'Approved' tests/api/test_finalized_immutability.sh"
check "PDF content test checks redacted text" "grep -q 'SECRET_NEVER_APPEAR' tests/api/test_pdf_content_acceptance.sh && grep -q 'pdftotext' tests/api/test_pdf_content_acceptance.sh"
check "Bates content test checks label" "grep -q 'CNT-001' tests/api/test_pdf_content_acceptance.sh"
check "backup metadata paths are asserted" "grep -q 'expect_json_nonempty.*artifacts.database_dump_path' tests/api/test_admin_ops.sh && grep -q 'expect_json_nonempty.*artifacts.file_snapshot_path' tests/api/test_admin_ops.sh"
check "backup restore consumes returned paths" "grep -q 'DB_DUMP_PATH=.*database_dump_path' tests/api/test_admin_ops.sh && grep -q 'FILE_SNAPSHOT_PATH=.*file_snapshot_path' tests/api/test_admin_ops.sh"
check "Bates sequence multi-document test exists" "test -f tests/api/test_bates_sequence_multi_doc.sh"
check "strict dependency negative test exists" "test -f tests/api/test_strict_dependency_failures.sh"
check "mention side-effect test exists" "test -f tests/api/test_notification_mention_side_effect.sh"
check "product limits are fixed constants" "grep -q 'productMaxUploadBytes int64 = 200 \* 1024 \* 1024' internal/app/config.go && grep -q 'productMaxPDFPages.*= 500' internal/app/config.go && grep -q 'productMaxBatchFiles.*= 250' internal/app/config.go && grep -q 'productMaxVersions.*= 50' internal/app/config.go"
check "product limit environment overrides are ignored" "grep -q 'TestLoadConfigIgnoresProductLimitEnvironmentOverrides' internal/app/config_test.go && ! grep -q 'envInt(\"MAX_' internal/app/config.go"
check "request timestamp 60-second boundary is explicit" "grep -q 'exact 60-second-old timestamp should be accepted' internal/core/rules_test.go && grep -q '61-second-old timestamp is rejected' tests/api/test_request_guard_edges.sh"
check "canonical UI is the only served HTML" "test \"$(find public -maxdepth 1 -type f -name '*.html' | wc -l | tr -d ' ')\" = 1 && test -f public/index.html"
check "canonical test directories are unambiguous" "test -d tests/api && test -d tests/contracts && test ! -e API_tests && test ! -e unit_tests"

if [ "$FAIL" -ne 0 ]; then exit 1; fi
