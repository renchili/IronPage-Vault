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
check "generation requires buildable UI decisions" "grep -q 'Frontend design and implementation contract' skills/project-generation-workflow/SKILL.md && grep -q 'exact icon library and icon name' skills/project-generation-workflow/SKILL.md && grep -q 'Special-interaction contract' skills/project-generation-workflow/SKILL.md && grep -q 'Platform and app-review compliance' skills/project-generation-workflow/SKILL.md && grep -q 'frontend portion is incomplete' skills/project-generation-workflow/SKILL.md"
check "generation does not invent UI artifact formats" "grep -q 'Do not invent YAML, JSON, schema, manifest, token-registry, or “review pack” deliverables' skills/project-generation-workflow/SKILL.md && grep -q 'Do not replace an interactive prototype with a prose specification' skills/project-generation-workflow/SKILL.md"
check "acceptance Skill is one complete static file" "test -f skills/full-project-acceptance-hard-gates/SKILL.md && test ! -e skills/full-project-acceptance-hard-gates/STATIC_GATES.md && grep -q 'Absolute static-only boundary' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'Missing execution evidence alone must not cause' skills/full-project-acceptance-hard-gates/SKILL.md"
check "acceptance completes every gate without external execution" "grep -q 'Every applicable Gate 0–27 must be completed' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'Inspect workflows only; never trigger or wait for CI' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'Missing test execution, CI, build, deployment, runtime logs, screenshots, or external artifacts does not alter the verdict' skills/full-project-acceptance-hard-gates/SKILL.md"
check "acceptance rejects non-buildable UI" "grep -q 'exact icon library and icon name' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'developers must still choose material icons' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'visual-only definitions of special interactions' skills/full-project-acceptance-hard-gates/SKILL.md && grep -q 'UI implementation readiness and platform review' skills/full-project-acceptance-hard-gates/SKILL.md"
check "obsolete execution-gated acceptance rules are absent" "! grep -Fq 'Missing execution evidence must be recorded as \`NOT VERIFIED\`' skills/full-project-acceptance-hard-gates/SKILL.md && ! grep -Fq 'PASS requires an executed generated summary' skills/full-project-acceptance-hard-gates/SKILL.md"
check "metadata exists" "test -f metadata.json"
check "canonical API test layout" "test -d tests/api && test -f tests/api/lib.sh && test ! -e API_tests"
check "canonical contract layout" "test -d tests/contracts && test ! -e unit_tests"
check "upload document version and audit are atomic" "bash tests/contracts/upload_audit_order.sh"
check "single acceptance UI" "test -f public/index.html && test ! -e public/manual-test.html"
check "air-gapped deployment scope" "test ! -e deploy/aws && test ! -e docs/aws-deployment.md && ! grep -RInE 'AWS|EKS|Lambda|CloudFormation' README.md docs scripts Dockerfile docker-compose.yml"
check "obsolete process status documents removed" "test ! -e docs/implementation-status.md && test ! -e docs/test-effectiveness-followup.md && test ! -d docs/review-fixes"
check "one-command deployer" "test -f scripts/deploy.sh && bash -n scripts/deploy.sh"
check "runtime file exclusions" "grep -Fxq '.env' .gitignore && grep -Fxq '.env' .dockerignore"
check "bootstrap acceptance uses deployer" "grep -q 'scripts/deploy.sh' tests/api/test_bootstrap_restart_docker.sh"
check "fresh deployment checks loopback host port" "grep -q 'select_available_host_port' scripts/deploy.sh && grep -q 'host_port_available' scripts/deploy.sh && grep -q 'localhost did not resolve to an IPv4 loopback address' scripts/deploy.sh"

check "application local configuration has no fallback" "! grep -Eq 'env\(\"(HTTP_ADDR|DB_PORT|DB_USER|DB_NAME|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR)\", \"[^\"]+\"\)' internal/app/config.go && ! grep -q 'DBHost' internal/app/config.go"
check "sensitive configuration has no fallback" "grep -q 'DBPassword:.*env(\"DB_PASSWORD\", \"\")' internal/app/config.go && grep -q 'JWTSecret:.*env(\"JWT_SECRET\", \"\")' internal/app/config.go && grep -q 'AESKey:.*env(\"AES_KEY\", \"\")' internal/app/config.go"
check "bcrypt password limit is validated" "grep -q '72-byte limit' internal/app/config.go && grep -q 'RejectsBootstrapPasswordAboveBcryptLimit' internal/app/config_test.go"
check "Compose has no fixed container identity" "! grep -q 'container_name:' docker-compose.yml"
check "Compose has no local value fallbacks" "! grep -Eq '\$\{(HOST_BIND_ADDRESS|HOST_PORT|HTTP_PORT|DB_PORT|DB_USER|DB_NAME|STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR):-' docker-compose.yml"
check "image has no fixed runtime path ENV" "! grep -Eq '^ENV (STORAGE_DIR|BACKUP_DIR|MIGRATIONS_DIR|PUBLIC_DIR|HTTP_ADDR)=' Dockerfile"
check "schema does not seed a fixed backup path" "! grep -q '/var/lib/ironpage/backups' migrations/001_schema.sql && grep -q 'EnsureRuntimeConfiguration' internal/app/database.go"

check "schema and upgrade migrations exist" "test -f migrations/001_schema.sql && test -f migrations/002_login_attempt_window.sql && test -f migrations/003_audit_lookup_and_runtime_config.sql"
check "roles and workflow in schema" "grep -q 'Admin' migrations/001_schema.sql && grep -q 'Editor' migrations/001_schema.sql && grep -q 'Reviewer' migrations/001_schema.sql && grep -q 'Finalized' migrations/001_schema.sql"
check "workflow definitions are Admin-managed and runtime-resolved" "grep -Fq 'admin.PUT(\"/workflow-statuses\"' internal/app/server.go && grep -q 'replaceWorkflowStatuses' internal/app/workflow_definitions.go && grep -q 'nextWorkflowDefinition' internal/app/workflows.go"
check "workflow required states and locks are preserved" "grep -q 'required workflow statuses must preserve their domain order' internal/app/workflow_definitions.go && grep -q 'ACCESS EXCLUSIVE MODE' internal/app/workflow_definitions.go && grep -q 'SHARE MODE' internal/app/workflows.go"
check "workflow mutation side effects share a transaction" "grep -q 'document_status_history' internal/app/workflows.go && grep -q 'auditWithExecutor(c, tx' internal/app/workflows.go && grep -q 'notifyUserWithExecutor(c, tx' internal/app/workflows.go"
check "strict PDF and backup entrypoints" "grep -q 'RewritePDFWithRedactionsStrict' internal/platform/pdf_strict.go && grep -q 'RewritePDFWithBatesStrict' internal/platform/pdf_strict.go && grep -q 'RunBackupArtifactsStrict' internal/platform/backup_strict.go && grep -q 'RunRestoreArtifactsStrict' internal/platform/backup_strict.go"
check "redaction mutation checks every state write" "grep -q 'x_ciphertext' internal/app/redactions.go && grep -q 'DOCUMENT_UPDATE_ERROR' internal/app/redactions.go && grep -q 'REDACTION_STATE_ERROR' internal/app/redactions.go && grep -q 'os.Remove(dst)' internal/app/redactions.go"
check "Bates range version document and audit share a transaction" "grep -q 'AllocateBatesRange' internal/app/bates_version.go internal/repository/bates.go && ! grep -q 'AllocateBatesStart' internal/repository/bates.go && grep -q 'auditWithExecutor(c, tx' internal/app/bates_version.go && grep -q 'os.Remove(dst)' internal/app/bates_version.go"
check "audit and notification helpers return errors" "grep -q 'func (a \*App) audit.* error' internal/app/domain_events.go && grep -q 'func (a \*App) notifyUser.* error' internal/app/domain_events.go && ! grep -q '_ = a.insertAuditRecord' internal/app/domain_events.go"
check "audit API opens protected source and metadata" "grep -q 'source_ip_lookup' internal/repository/audit.go && grep -q 'openAuditPII' internal/app/audit_filters.go && grep -q 'EnsureAuditSourceIPLookups' internal/app/server.go"
check "notification cap is serialized" "grep -Fq 'SELECT id FROM users WHERE id=\$1 FOR UPDATE' internal/app/notifications.go && grep -q 'sort.Strings(usernames)' internal/app/mentions.go"
check "rolling lockout and auth audit fail closed" "grep -q 'loginAttemptWindow.*15 \* time.Minute' internal/app/auth.go && grep -q 'auditWithExecutor(c, tx, userID, \"LOGIN_FAILED\"' internal/app/auth.go && grep -q 'auditWithExecutor(c, tx, p.UserID, \"LOGOUT\"' internal/app/auth.go"
check "restore stages files and preserves acting user" "grep -q 'unsafe path in filesystem snapshot' internal/platform/backup_exec.go && grep -q -- '--single-transaction' internal/platform/backup_exec.go && grep -q 'ActorUserID:       p.UserID' internal/app/restore.go && grep -q 'BACKUP_RESTORE_COMPLETED' internal/app/restore.go && ! grep -Eq 'recordRestoreState\([^\n]*, \"\"' internal/app/restore.go"
check "restore lifecycle is durably reconciled" "test -f internal/app/restore_lifecycle.go && test -f internal/app/restore_lifecycle_test.go && grep -q 'writeRestoreLifecycleRecord' internal/app/restore.go && grep -q 'reconcileRestoreLifecycle' internal/app/server.go internal/app/restore_lifecycle.go && grep -q 'restore process ended before terminal lifecycle persistence' internal/app/restore_lifecycle.go && grep -q 'ensureRestoreAudit' internal/app/restore_lifecycle.go"
check "restore lifecycle journal is protected" "grep -q 'encryptString(a.cfg.AESKey' internal/app/restore_lifecycle.go && grep -q 'decryptString(a.cfg.AESKey' internal/app/restore_lifecycle.go && grep -q 'AES-256-GCM' internal/app/restore_lifecycle.go && grep -q 'journal exposed protected value' internal/app/restore_lifecycle_test.go"
check "all collection handlers use configured pagination" "test \"\$(grep -c 'configuredPage(c)' internal/app/admin.go)\" -ge 6 && grep -A20 'func (a \*App) listVersions' internal/app/documents.go | grep -q 'LIMIT \$2 OFFSET \$3' && grep -A25 'func (a \*App) listRedactions' internal/app/redactions.go | grep -q 'LIMIT \$2 OFFSET \$3' && grep -A30 'func (a \*App) listAnnotations' internal/app/annotations.go | grep -q 'LIMIT \$2 OFFSET \$3'"
check "pagination contracts cover every collection" "grep -Fq '/api/documents' tests/api/test_api_contracts.sh && grep -Fq '/api/admin/users' tests/api/test_api_contracts.sh && grep -Fq '/api/admin/config' tests/api/test_api_contracts.sh && grep -Fq '/api/admin/workflow-statuses' tests/api/test_api_contracts.sh && grep -Fq '/api/admin/notification-templates' tests/api/test_api_contracts.sh && grep -Fq '/api/admin/backup/jobs' tests/api/test_api_contracts.sh && grep -Fq '/versions' tests/api/test_api_contracts.sh && grep -Fq '/redactions' tests/api/test_api_contracts.sh && grep -Fq '/annotations' tests/api/test_api_contracts.sh && grep -Fq '/api/notifications' tests/api/test_api_contracts.sh && grep -Fq '/api/audit-logs' tests/api/test_api_contracts.sh && grep -q 'page_size=0' tests/api/test_api_contracts.sh && grep -q 'page_size=1' tests/api/test_api_contracts.sh && grep -q 'page_size=101' tests/api/test_api_contracts.sh"
check "Swaggo collection annotations include page parameters" "test \"\$(grep -R '@Param page query int' internal/app/swagger_*.go | wc -l | tr -d ' ')\" -ge 11 && test \"\$(grep -R '@Param page_size query int' internal/app/swagger_*.go | wc -l | tr -d ' ')\" -ge 11"

check "single static workflow" "test \"\$(find .github/workflows -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \) | wc -l | tr -d ' ')\" = 1 && test -f .github/workflows/ci.yml"
check "workflow collapses duplicate target runs" "grep -q 'cancel-in-progress: true' .github/workflows/ci.yml && grep -q 'Cancelled superseded active run' .github/workflows/ci.yml"
check "admission precedes checkout" "test \"\$(grep -n 'actions/github-script@v7' .github/workflows/ci.yml | head -1 | cut -d: -f1)\" -lt \"\$(grep -n 'actions/checkout@v4' .github/workflows/ci.yml | head -1 | cut -d: -f1)\""
check "manual target is canonicalized" "grep -q 'Manual target.*must equal the selected branch' .github/workflows/ci.yml && grep -q 'sameRevision' .github/workflows/ci.yml && grep -q 'sameRepository' .github/workflows/ci.yml && grep -q 'exact open PR revision' .github/workflows/ci.yml"
check "workflow paginates complete scoped history" "grep -q 'github.paginate' .github/workflows/ci.yml && grep -q 'listWorkflowRuns' .github/workflows/ci.yml && grep -q 'workflow_id: workflowId' .github/workflows/ci.yml"
check "cooldown is scoped to the exact revision" "grep -q 'latestCompletedSameRevision' .github/workflows/ci.yml && grep -q 'run.head_sha === currentSha' .github/workflows/ci.yml && grep -q 'same-revision admission cooldown' .github/workflows/ci.yml"
check "one-time unlock is exact and auditable" "grep -q 'unlock_failed_run_id' .github/workflows/ci.yml && grep -q 'unlock_reason' .github/workflows/ci.yml && grep -q 'alreadyConsumed' .github/workflows/ci.yml"
check "workflow is static and fail-fast" "! grep -RIn 'if: always()' .github/workflows && ! grep -q 'run_full_regression.sh' .github/workflows/ci.yml && ! grep -Eq 'go test|go vet|docker (build|compose)|run_tests.sh' .github/workflows/ci.yml"
check "successful static inventory is retained" "grep -q 'actions/upload-artifact@v4' .github/workflows/ci.yml && grep -q 'artifacts/static-acceptance/source-inventory.json' .github/workflows/ci.yml"
check "local report derives executed stages" "grep -q 'executed_stages' run_tests.sh && grep -q 'Executed stages' run_tests.sh"
check "generated API documentation path" "test -f scripts/generate_swagger.sh && grep -q 'require_route put /api/admin/workflow-statuses' tests/contracts/swagger_route_coverage.sh"

TOTAL=$((PASS+FAIL))
echo "CONTRACT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
