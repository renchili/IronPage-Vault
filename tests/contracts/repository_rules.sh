#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

check() {
  local name="$1" cmd="$2"
  if bash -lc "$cmd" >/tmp/ironpage_contract.out 2>&1; then
    echo "PASS contract: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL contract: $name"
    cat /tmp/ironpage_contract.out
    FAIL=$((FAIL+1))
  fi
}

check "rule entrypoints exist" "test -f AGENTS.md && test -f AGENT.md"
check "metadata exists" "test -f metadata.json"
check "canonical API test layout" "test -d tests/api && test -f tests/api/lib.sh && test ! -e API_tests"
check "canonical contract layout" "test -d tests/contracts && test ! -e unit_tests"
check "single acceptance UI" "test -f public/index.html && test ! -e public/manual-test.html"
check "one-command deployer" "test -f scripts/deploy.sh && bash -n scripts/deploy.sh"
check "runtime file exclusions" "grep -Fxq '.env' .gitignore && grep -Fxq '.env' .dockerignore"
check "bootstrap acceptance uses deployer" "grep -q 'scripts/deploy.sh' tests/api/test_bootstrap_restart_docker.sh"

check "application local configuration has no fallback" "! grep -Eq 'env\(\"(HTTP_ADDR|DB_PORT|DB_USER|DB_NAME|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR)\", \"[^\"]+\"\)' internal/app/config.go"
check "sensitive configuration has no fallback" "grep -q 'DBPassword:.*env(\"DB_PASSWORD\", \"\")' internal/app/config.go && grep -q 'JWTSecret:.*env(\"JWT_SECRET\", \"\")' internal/app/config.go && grep -q 'AESKey:.*env(\"AES_KEY\", \"\")' internal/app/config.go"
check "bcrypt password limit is validated" "grep -q '72-byte limit' internal/app/config.go && grep -q 'RejectsBootstrapPasswordAboveBcryptLimit' internal/app/config_test.go"
check "Compose has no fixed container identity" "! grep -q 'container_name:' docker-compose.yml"
check "Compose has no local value fallbacks" "! grep -Eq '\$\{(HOST_PORT|HTTP_PORT|DB_PORT|DB_USER|DB_NAME|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR):-' docker-compose.yml"
check "Compose maps generated database identity" "grep -q 'POSTGRES_USER: \${DB_USER}' docker-compose.yml && grep -q 'POSTGRES_DB: \${DB_NAME}' docker-compose.yml && grep -q 'DB_PORT: \${DB_PORT}' docker-compose.yml"
check "image has no fixed runtime path ENV" "! grep -Eq '^ENV (STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR|HTTP_ADDR)=' Dockerfile"
check "entrypoint uses generated database port" "grep -q 'postgres -p \"\$DB_PORT\"' scripts/entrypoint.sh && ! grep -q -- '-p 5432' scripts/entrypoint.sh"

check "schema and rolling login migration" "test -f migrations/001_schema.sql && test -f migrations/002_login_attempt_window.sql"
check "roles and workflow in schema" "grep -q 'Admin' migrations/001_schema.sql && grep -q 'Editor' migrations/001_schema.sql && grep -q 'Reviewer' migrations/001_schema.sql && grep -q 'Finalized' migrations/001_schema.sql"
check "strict PDF and backup entrypoints" "grep -q 'RewritePDFWithRedactionsStrict' internal/platform/pdf_strict.go && grep -q 'RewritePDFWithBatesStrict' internal/platform/pdf_strict.go && grep -q 'RunBackupArtifactsStrict' internal/platform/backup_strict.go && grep -q 'RunRestoreArtifactsStrict' internal/platform/backup_strict.go"
check "encrypted redaction coordinates" "grep -q 'x_ciphertext' internal/app/review.go && grep -q 'EncryptedRedactionRegions' internal/app/coordinate_crypto.go"
check "rolling lockout and fail-closed errors" "grep -q 'loginAttemptWindow.*15 \* time.Minute' internal/app/auth.go && grep -q 'LOGIN_ATTEMPT_WRITE_ERROR' internal/app/auth.go && grep -q 'LOGIN_STATE_WRITE_ERROR' internal/app/auth.go && grep -q 'AUTH_STATE_READ_ERROR' internal/app/auth.go && grep -q 'SESSION_UPDATE_ERROR' internal/app/auth.go && grep -q 'LOGOUT_WRITE_ERROR' internal/app/auth.go"

check "single serialized workflow" "test \"$(find .github/workflows -maxdepth 1 -type f | wc -l | tr -d ' ')\" = 1 && test -f .github/workflows/ci.yml"
check "workflow has shared lock and failure latch" "grep -q 'concurrency:' .github/workflows/ci.yml && grep -q 'cancel-in-progress: false' .github/workflows/ci.yml && grep -q 'ci_execution_guard.py' .github/workflows/ci.yml"
check "workflow starts no post-failure action" "! grep -RIn 'if: always()' .github/workflows && grep -q 'run_full_regression.sh' .github/workflows/ci.yml"
check "full regression is fail fast" "grep -q 'writes the summary, and exits' ci/run_full_regression.sh && grep -q 'exit \"\$status\"' ci/run_full_regression.sh"
check "local report derives executed stages" "grep -q 'executed_stages' run_tests.sh && grep -q 'Executed stages' run_tests.sh && ! grep -q 'This local report covers Swagger generation' run_tests.sh"

check "documentation consistency gate is in regression" "grep -q 'documentation_consistency' ci/run_full_regression.sh && test -f ci/docs_consistency_check.sh"
check "browser interaction acceptance" "test -f tests/api/test_ui_interaction_acceptance.sh && test -f tests/api/ui_interaction_cdp.py && grep -q 'id=\"login-form\"' public/index.html"
check "generated API documentation path" "test -f scripts/generate_swagger.sh && test -f tests/contracts/swagger_route_coverage.sh"

TOTAL=$((PASS+FAIL))
echo "CONTRACT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
