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
check "one-command deployer exists" "test -f scripts/deploy.sh && bash -n scripts/deploy.sh"
check "generated runtime file is excluded" "grep -Fxq '.env' .gitignore && grep -Fxq '.env' .dockerignore"
check "one-command bootstrap acceptance exists" "grep -q 'scripts/deploy.sh' API_tests/test_bootstrap_restart_docker.sh"
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
check "Swaggo annotation sources exist" "test -f internal/app/swagger_auth_admin.go && test -f internal/app/swagger_documents.go"
check "Swagger route coverage test exists" "test -f unit_tests/test_swagger_route_coverage.sh"
check "Swagger route coverage passes" "bash unit_tests/test_swagger_route_coverage.sh"
check "ci workflow exists" "test -f .github/workflows/ci.yml"
check "strict pdf entrypoints exist" "grep -q 'RewritePDFWithRedactionsStrict' internal/platform/pdf_strict.go && grep -q 'RewritePDFWithBatesStrict' internal/platform/pdf_strict.go"
check "strict backup entrypoints exist" "grep -q 'RunBackupArtifactsStrict' internal/platform/backup_strict.go && grep -q 'RunRestoreArtifactsStrict' internal/platform/backup_strict.go"
check "redaction coordinates stored encrypted only" "grep -q 'x_ciphertext' internal/app/review.go && grep -q 'y_ciphertext' internal/app/review.go && grep -q 'width_ciphertext' internal/app/review.go && grep -q 'height_ciphertext' internal/app/review.go"
check "redaction burn-in uses encrypted coordinates" "grep -q 'EncryptedRedactionRegions' internal/app/coordinate_crypto.go"
check "redaction list hides coordinate plaintext" "! grep -q 'SELECT id,document_id,page,x,y,width,height,reason' internal/app/review.go"
check "restore strict failure path exists" "grep -q 'RunRestoreArtifactsStrict' internal/app/restore.go"
check "self-contained compare test exists" "test -f API_tests/test_compare_self_contained.sh"
check "structure rules suite exists" "test -f unit_tests/test_structure_rules.sh"
check "strict dependency failure API test exists" "test -f API_tests/test_strict_dependency_failures.sh"
check "bates sequence multi doc API test exists" "test -f API_tests/test_bates_sequence_multi_doc.sh"
check "platform strict tests exist" "test -f internal/platform/pdf_strict_test.go && test -f internal/platform/backup_strict_test.go"
check "scheduled backup interval test exists" "test -f internal/app/backup_interval_test.go"
check "runtime configuration tests exist" "test -f internal/app/config_test.go"
check "sensitive config has no fallback" "grep -q 'DBPassword:.*env(\"DB_PASSWORD\", \"\")' internal/app/config.go && grep -q 'JWTSecret:.*env(\"JWT_SECRET\", \"\")' internal/app/config.go && grep -q 'AESKey:.*env(\"AES_KEY\", \"\")' internal/app/config.go"
check "initial users use secure dispatcher" "grep -q 'EnsureInitialUsers' internal/app/server.go && grep -q 'empty user store requires' internal/app/database.go"
check "acceptance mode gates seed users" "grep -q 'seed users require acceptance mode' internal/app/database.go && grep -q 'bootstrap admin values are not allowed in acceptance mode' internal/app/config.go"
check "compose requires sensitive values" "grep -q 'DB_PASSWORD is required' docker-compose.yml && grep -q 'JWT_SECRET is required' docker-compose.yml && grep -q 'AES_KEY is required' docker-compose.yml"
check "compose maps database configuration consistently" "grep -q 'POSTGRES_USER: \${DB_USER:-ironpage}' docker-compose.yml && grep -q 'POSTGRES_DB: \${DB_NAME:-ironpage}' docker-compose.yml && grep -q 'DB_HOST: \${DB_HOST:-127.0.0.1}' docker-compose.yml"
check "compose wires bootstrap variables" "grep -q 'BOOTSTRAP_ADMIN_USERNAME' docker-compose.yml && grep -q 'BOOTSTRAP_ADMIN_PASSWORD' docker-compose.yml"
check "test UI has no embedded password values" "! grep -R -E 'type=\"password\"[^>]*value=|const passwords[[:space:]]*=' public"
check "API acceptance requires injected credentials" "grep -q 'SEED_ADMIN_PASSWORD is required for acceptance tests' API_tests/test_api_flow.sh"

check "container image has no runtime credential defaults" "! grep -Eq '^ENV (POSTGRES_USER|POSTGRES_PASSWORD|POSTGRES_DB|DB_USER|DB_PASSWORD|DB_NAME|JWT_SECRET|AES_KEY|SEED_.*PASSWORD)=' Dockerfile"
check "entrypoint has no database identity fallback" "! grep -Eq 'POSTGRES_USER:-|POSTGRES_PASSWORD:-|POSTGRES_DB:-|DB_USER:-|DB_PASSWORD:-|DB_NAME:-' scripts/entrypoint.sh"
check "entrypoint requires PostgreSQL runtime values" "grep -q 'POSTGRES_USER' scripts/entrypoint.sh && grep -q 'POSTGRES_PASSWORD' scripts/entrypoint.sh && grep -q 'POSTGRES_DB' scripts/entrypoint.sh"
check "entrypoint requires application runtime values" "grep -q 'DB_PASSWORD' scripts/entrypoint.sh && grep -q 'JWT_SECRET' scripts/entrypoint.sh && grep -q 'AES_KEY' scripts/entrypoint.sh"
check "entrypoint checks single-container database consistency" "grep -q 'POSTGRES_USER and DB_USER must match' scripts/entrypoint.sh && grep -q 'POSTGRES_PASSWORD and DB_PASSWORD must match' scripts/entrypoint.sh && grep -q 'POSTGRES_DB and DB_NAME must match' scripts/entrypoint.sh"
check "entrypoint gates acceptance identities" "grep -q 'SEED_ADMIN_PASSWORD' scripts/entrypoint.sh && grep -q 'ACCEPTANCE_MODE' scripts/entrypoint.sh"

check "rolling login attempt migration exists" "test -f migrations/002_login_attempt_window.sql && grep -q 'CREATE TABLE IF NOT EXISTS login_attempts' migrations/002_login_attempt_window.sql"
check "rolling lockout implementation exists" "grep -q 'loginAttemptWindow.*15 \\* time.Minute' internal/app/auth.go && grep -q 'SELECT id FROM users WHERE id=\\$1 FOR UPDATE' internal/app/auth.go"
check "authentication state fails closed" "grep -q 'LOGIN_ATTEMPT_WRITE_ERROR' internal/app/auth.go && grep -q 'LOGIN_STATE_WRITE_ERROR' internal/app/auth.go && grep -q 'AUTH_STATE_READ_ERROR' internal/app/auth.go && grep -q 'SESSION_UPDATE_ERROR' internal/app/auth.go && grep -q 'LOGOUT_WRITE_ERROR' internal/app/auth.go"
check "rolling lockout Docker test exists" "test -f API_tests/test_auth_lockout_docker.sh"
check "normal bootstrap restart Docker test exists" "test -f API_tests/test_bootstrap_restart_docker.sh"
check "browser interaction acceptance exists" "test -f API_tests/test_ui_interaction_acceptance.sh && test -f API_tests/ui_interaction_cdp.py"
check "acceptance orchestrator executes new evidence" "grep -q 'test_bootstrap_restart_docker.sh' ci/docker_acceptance.sh && grep -q 'test_auth_lockout_docker.sh' ci/docker_acceptance.sh && grep -q 'test_ui_interaction_acceptance.sh' ci/docker_acceptance.sh"
check "documentation consistency gate exists" "test -f ci/docs_consistency_check.sh && test -f .github/workflows/documentation-consistency.yml && grep -q 'bash ci/docs_consistency_check.sh' .github/workflows/documentation-consistency.yml"
check "acceptance UI supports form and live status" "grep -q 'id=\"login-form\"' public/index.html && grep -q 'aria-live=\"polite\"' public/index.html && grep -q 'focus-visible' public/index.html"

TOTAL=$((PASS+FAIL))
echo "UNIT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
