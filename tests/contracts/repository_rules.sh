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
check "rule entrypoint roles are unambiguous" "grep -q 'mandatory repository entrypoint' AGENTS.md && grep -q 'Project-adapted rules belong in \`AGENT.md\`' AGENTS.md && grep -q '# AGENT Rules for IronPage Vault' AGENT.md"
check "generation is static and externally independent" "grep -q 'static source-completion tasks' skills/project-generation-workflow/SKILL.md && grep -q 'trigger, rerun, retry, dispatch, or wait for CI' skills/project-generation-workflow/SKILL.md && grep -q 'CI triggered or awaited: \`none\`' skills/project-generation-workflow/SKILL.md"
check "generation rejects minimization and early stop" "grep -q 'must not optimize for the smallest change count' skills/project-generation-workflow/SKILL.md && grep -q 'Do not stop scanning after the first P0' skills/project-generation-workflow/SKILL.md && grep -q 'Continue until no known in-scope static defect is deferred' skills/project-generation-workflow/SKILL.md"
check "acceptance Skill is one complete static file" "test -f skills/full-project-acceptance-hard-gates/SKILL.md && test ! -e skills/full-project-acceptance-hard-gates/STATIC_GATES.md && grep -q 'Absolute static-only boundary' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'Missing execution evidence alone must not cause' skills/full-project-acceptance-hard-gates/SKILL.md"
check "acceptance completes every gate without external execution" "grep -q 'Every applicable Gate 0–27 must be completed' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'Inspect workflows only; never trigger or wait for CI' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'Missing test execution, CI, build, deployment, runtime logs, screenshots, or external artifacts does not alter the verdict' skills/full-project-acceptance-hard-gates/SKILL.md"
check "acceptance retains CI safety review" "grep -q 'shared concurrency' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'duplicate collapse' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'failed-revision latches' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'unlock scope' skills/full-project-acceptance-hard-gates/SKILL.md"
check "obsolete execution-gated acceptance rules are absent" "! grep -Fq 'Missing execution evidence must be recorded as \`NOT VERIFIED\`' skills/full-project-acceptance-hard-gates/SKILL.md && ! grep -Fq 'PASS requires an executed generated summary' skills/full-project-acceptance-hard-gates/SKILL.md && ! grep -Fq 'must be executed with reproducible evidence' skills/full-project-acceptance-hard-gates/SKILL.md && ! grep -Fq 'Real interaction flows executed' skills/full-project-acceptance-hard-gates/SKILL.md"
check "metadata exists" "test -f metadata.json"
check "canonical API test layout" "test -d tests/api && test -f tests/api/lib.sh && test ! -e API_tests"
check "canonical contract layout" "test -d tests/contracts && test ! -e unit_tests"
check "upload audit follows document commit" "bash tests/contracts/upload_audit_order.sh"
check "single acceptance UI" "test -f public/index.html && test ! -e public/manual-test.html"
check "air-gapped deployment scope" "test ! -e deploy/aws && test ! -e docs/aws-deployment.md && ! grep -RInE 'AWS|EKS|Lambda|CloudFormation' README.md docs scripts Dockerfile docker-compose.yml"
check "obsolete review process docs removed" "test ! -d docs/review-fixes || test -z \"\$(find docs/review-fixes -type f -print -quit)\""
check "one-command deployer" "test -f scripts/deploy.sh && bash -n scripts/deploy.sh"
check "runtime file exclusions" "grep -Fxq '.env' .gitignore && grep -Fxq '.env' .dockerignore"
check "bootstrap acceptance uses deployer" "grep -q 'scripts/deploy.sh' tests/api/test_bootstrap_restart_docker.sh"
check "fresh deployment checks loopback host port" "grep -q 'select_available_host_port' scripts/deploy.sh && grep -q 'host_port_available' scripts/deploy.sh && grep -q 'localhost did not resolve to an IPv4 loopback address' scripts/deploy.sh"

check "application local configuration has no fallback" "! grep -Eq 'env\(\"(HTTP_ADDR|DB_PORT|DB_USER|DB_NAME|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR)\", \"[^\"]+\"\)' internal/app/config.go && ! grep -q 'DBHost' internal/app/config.go"
check "sensitive configuration has no fallback" "grep -q 'DBPassword:.*env(\"DB_PASSWORD\", \"\")' internal/app/config.go && grep -q 'JWTSecret:.*env(\"JWT_SECRET\", \"\")' internal/app/config.go && grep -q 'AESKey:.*env(\"AES_KEY\", \"\")' internal/app/config.go"
check "bcrypt password limit is validated" "grep -q '72-byte limit' internal/app/config.go && grep -q 'RejectsBootstrapPasswordAboveBcryptLimit' internal/app/config_test.go"
check "Compose has no fixed container identity" "! grep -q 'container_name:' docker-compose.yml"
check "Compose has no local value fallbacks" "! grep -Eq '\$\{(HOST_BIND_ADDRESS|HOST_PORT|HTTP_PORT|DB_PORT|DB_USER|DB_NAME|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR):-' docker-compose.yml"
check "Compose maps generated database identity" "grep -q 'POSTGRES_USER: \${DB_USER}' docker-compose.yml && grep -q 'POSTGRES_DB: \${DB_NAME}' docker-compose.yml && grep -q 'DB_PORT: \${DB_PORT}' docker-compose.yml"
check "image has no fixed runtime path ENV" "! grep -Eq '^ENV (STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR|HTTP_ADDR)=' Dockerfile"
check "entrypoint uses generated database port" "grep -q 'postgres -p \"\$DB_PORT\"' scripts/entrypoint.sh && ! grep -q -- '-p 5432' scripts/entrypoint.sh"

check "schema and rolling login migration" "test -f migrations/001_schema.sql && test -f migrations/002_login_attempt_window.sql"
check "roles and workflow in schema" "grep -q 'Admin' migrations/001_schema.sql && grep -q 'Editor' migrations/001_schema.sql && grep -q 'Reviewer' migrations/001_schema.sql && grep -q 'Finalized' migrations/001_schema.sql"
check "strict PDF and backup entrypoints" "grep -q 'RewritePDFWithRedactionsStrict' internal/platform/pdf_strict.go && grep -q 'RewritePDFWithBatesStrict' internal/platform/pdf_strict.go && grep -q 'RunBackupArtifactsStrict' internal/platform/backup_strict.go && grep -q 'RunRestoreArtifactsStrict' internal/platform/backup_strict.go"
check "encrypted redaction coordinates" "grep -q 'x_ciphertext' internal/app/review.go && grep -q 'EncryptedRedactionRegions' internal/app/coordinate_crypto.go"
check "rolling lockout and fail-closed errors" "grep -q 'loginAttemptWindow.*15 \* time.Minute' internal/app/auth.go && grep -q 'LOGIN_ATTEMPT_WRITE_ERROR' internal/app/auth.go && grep -q 'LOGIN_STATE_WRITE_ERROR' internal/app/auth.go && grep -q 'AUTH_STATE_READ_ERROR' internal/app/auth.go && grep -q 'SESSION_UPDATE_ERROR' internal/app/auth.go && grep -q 'LOGOUT_WRITE_ERROR' internal/app/auth.go"

check "single static workflow" "test \"\$(find .github/workflows -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \) | wc -l | tr -d ' ')\" = 1 && test -f .github/workflows/ci.yml"
check "workflow collapses duplicate target runs" "grep -q 'cancel-in-progress: true' .github/workflows/ci.yml && grep -q 'Cancelled superseded active run' .github/workflows/ci.yml"
check "admission precedes checkout" "test \"\$(grep -n 'actions/github-script@v7' .github/workflows/ci.yml | head -1 | cut -d: -f1)\" -lt \"\$(grep -n 'actions/checkout@v4' .github/workflows/ci.yml | head -1 | cut -d: -f1)\""
check "admission does not sleep" "! grep -Eq 'time\.sleep|sleep [0-9]' .github/workflows/ci.yml"
check "workflow paginates complete scoped history" "grep -q 'github.paginate' .github/workflows/ci.yml && grep -q 'listWorkflowRuns' .github/workflows/ci.yml && grep -q 'workflow_id: workflowId' .github/workflows/ci.yml"
check "cooldown is scoped to the exact revision" "grep -q 'latestCompletedSameRevision' .github/workflows/ci.yml && grep -q 'run.head_sha === currentSha' .github/workflows/ci.yml && grep -q 'same-revision admission cooldown' .github/workflows/ci.yml"
check "new revision is documented as admissible" "grep -q 'new revision is admitted immediately' ci/BOUNDARY.md && grep -q 'corrective commit can be checked' README.md docs/testing.md"
check "one-time unlock is exact and auditable" "grep -q 'unlock_failed_run_id' .github/workflows/ci.yml && grep -q 'unlock_reason' .github/workflows/ci.yml && grep -q 'alreadyConsumed' .github/workflows/ci.yml && grep -q 'One-time unlock.*already been consumed' .github/workflows/ci.yml"
check "ordinary reruns do not bypass admission" "grep -q 'currentAttempt > 1' .github/workflows/ci.yml"
check "workflow is static and fail-fast" "! grep -RIn 'if: always()' .github/workflows && ! grep -q 'run_full_regression.sh' .github/workflows/ci.yml && ! grep -Eq 'go test|go vet|docker (build|compose)|run_tests.sh' .github/workflows/ci.yml"
check "workflow exposes static gates" "grep -q 'shell_syntax_check.sh' .github/workflows/ci.yml && grep -q 'source_inventory.py' .github/workflows/ci.yml && grep -q 'docs_consistency_check.sh' .github/workflows/ci.yml && grep -q 'tests/contracts/repository_rules.sh' .github/workflows/ci.yml"
check "successful static inventory is retained" "grep -q 'actions/upload-artifact@v4' .github/workflows/ci.yml && grep -q 'artifacts/static-acceptance/source-inventory.json' .github/workflows/ci.yml && grep -q 'retention-days: 90' .github/workflows/ci.yml"
check "source inventory audits generic path hazards" "grep -q 'non-ASCII tracked path' ci/source_inventory.py && grep -q 'case-only path collision' ci/source_inventory.py && grep -q 'near-duplicate sibling paths' ci/source_inventory.py && grep -q 'mixed hyphen/underscore naming' ci/source_inventory.py && grep -q 'path_hygiene_findings' ci/source_inventory.py && grep -q 'allowed_path_exceptions' ci/source_inventory.py"
check "full regression helper remains fail fast" "grep -q 'summary, and exits' ci/run_full_regression.sh && grep -q 'exit \"\$status\"' ci/run_full_regression.sh"
check "local report derives executed stages" "grep -q 'executed_stages' run_tests.sh && grep -q 'Executed stages' run_tests.sh && ! grep -q 'This local report covers Swagger generation' run_tests.sh"

check "documentation consistency gate is in static workflow" "grep -q 'docs_consistency_check.sh' .github/workflows/ci.yml && test -f ci/docs_consistency_check.sh"
check "browser interaction acceptance" "test -f tests/api/test_ui_interaction_acceptance.sh && test -f tests/api/ui_interaction_cdp.py && grep -q 'id=\"login-form\"' public/index.html"
check "generated API documentation path" "test -f scripts/generate_swagger.sh && test -f tests/contracts/swagger_route_coverage.sh"

TOTAL=$((PASS+FAIL))
echo "CONTRACT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
