#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "FAIL backup/restore integrity: $1"
  exit 1
}

require() {
  local pattern="$1" file="$2" message="$3"
  grep -Fq "$pattern" "$file" || fail "$message"
}

require 'pg_advisory_lock_shared($1)' internal/app/operation_barrier.go "shared request advisory lock missing"
require 'pg_advisory_lock($1)' internal/app/operation_barrier.go "exclusive backup/restore advisory lock missing"
require 'e.Use(a.maintenanceMiddleware)' internal/app/server.go "maintenance middleware not wired"
require 'e.Use(a.mutationBarrierMiddleware)' internal/app/server.go "request barrier middleware not wired"
require 'requiresRequestBarrier' internal/app/operation_barrier.go "authenticated read/write barrier classification missing"
require 'strings.HasPrefix(path, "/api/")' internal/app/operation_barrier.go "API authentication state is not covered by the request barrier"
require 'isExclusiveOperationPath(c.Request().URL.Path)' internal/app/auth.go "exclusive operations do not isolate authentication writes before lock promotion"
require 'withSharedMutation(c.Request().Context()' internal/app/auth.go "exclusive operation authentication state is not protected by a shared barrier"
require 'withExclusiveOperation(c.Request().Context()' internal/app/backup_file.go "manual backup does not own exclusive barrier"
require 'withExclusiveOperation(ctx' internal/app/backup_scheduler.go "scheduled backup does not own exclusive barrier"
require 'withMaintenanceOperation(c.Request().Context()' internal/app/operation_barrier.go "restore middleware does not own maintenance barrier"
require 'admin.POST("/backup/restore", a.restoreBackup, a.restoreMaintenanceMiddleware)' internal/app/server.go "restore maintenance is not applied after Admin route authorization"
require 'admin.POST("/backup/restore/:id/resolve", a.resolveInterruptedRestore, a.exclusiveOperationMiddleware)' internal/app/server.go "restore resolution is not serialized with the exclusive barrier"
require 'isRestoreResolutionPath' internal/app/operation_barrier.go "restore resolution exclusive classification missing"
require 'maintenanceOwnerContextKey' internal/app/restore.go "restore handler does not require middleware maintenance ownership"
require 'restoreAdmission.TryLock()' internal/app/operation_barrier.go "concurrent restore admission guard missing"
require 'MAINTENANCE_MODE' internal/app/operation_barrier.go "maintenance requests do not fail closed"
require 'RESTORE_ALREADY_RUNNING' internal/app/operation_barrier.go "concurrent restore rejection missing"

require 'restoreStatusInterrupted' internal/app/restore_lifecycle.go "interrupted restore state missing"
require 'outcome"] = "unknown"' internal/app/restore_lifecycle.go "interrupted restore does not preserve unknown outcome"
require 'BACKUP_RESTORE_INTERRUPTED' internal/app/restore_lifecycle.go "interrupted restore audit missing"
require 'operator_verification_required' internal/app/restore_lifecycle.go "operator verification requirement missing"
require 'admin.POST("/backup/restore/:id/resolve"' internal/app/server.go "restore resolution route missing"
require 'func (a *App) resolveInterruptedRestore' internal/app/restore.go "restore resolution handler missing"
require 'EnsureSystemPrincipal(c.Request().Context(), a.db, a.cfg)' internal/app/restore.go "restored database does not re-establish the system principal"
if grep -A8 'record.Status == restoreStatusRequested' internal/app/restore_lifecycle.go | grep -q 'restoreStatusFailed'; then
  fail "Requested restore is still converted directly to Failed"
fi

require 'systemPrincipalID' internal/app/database.go "system principal identity missing"
require 'EnsureSystemPrincipal' internal/app/server.go "system principal startup initialization missing"
require "'scheduled_full_backup','Completed',\$2,\$3" internal/app/backup_scheduler.go "scheduled backup created_by is not explicit"
require 'systemPrincipalID, "SCHEDULED_BACKUP_CREATE"' internal/app/backup_scheduler.go "scheduled backup audit actor is not the system principal"
require 'audit acting user is required' internal/app/domain_events.go "audit helper does not reject blank actors"
require 'strings.EqualFold(strings.TrimSpace(req.Username), systemPrincipalUsername)' internal/app/auth.go "system principal username is not rejected at login"
require 'sub == systemPrincipalID' internal/app/auth.go "system principal token subject is not rejected"
require 'system principal cannot log in' tests/api/test_auth_rbac.sh "system principal negative login definition missing"
if grep -Fq 'NULLIF($2' internal/app/domain_events.go; then
  fail "audit actor can still be converted to NULL"
fi

require 'CONFIG_KEY_READ_ONLY' internal/app/admin.go "deployment-owned config rejection missing"
require 'CONFIG_KEY_NOT_MANAGED' internal/app/admin.go "unknown config key rejection missing"
require 'SELECT key,value FROM config_entries WHERE key IN ($1,$2) FOR UPDATE' internal/app/config_management.go "pagination pair is not locked for validation"
require 'validatePaginationValues(defaultSize, maxSize)' internal/app/config_management.go "pagination pair validation missing"
require 'backup.schedule_enabled' internal/app/config_management.go "Admin-managed backup enabled key missing"
require 'backup.interval' internal/app/config_management.go "Admin-managed backup interval key missing"
require 'loadBackupSchedule' internal/app/backup_scheduler.go "scheduler does not reload persisted configuration"
require 'runScheduledBackupIfDue' internal/app/backup_scheduler.go "scheduler due evaluation missing"
require 'MAX(created_at)' internal/app/backup_scheduler.go "scheduler does not persist restart timing"
if grep -R -q 'BACKUP_INTERVAL' internal/app docker-compose.yml scripts; then
  fail "scheduled backup still depends on an unreachable environment variable"
fi
require 'backup.schedule_enabled' tests/api/test_admin_ops.sh "Admin backup enabled API test missing"
require 'backup.interval' tests/api/test_admin_ops.sh "Admin backup interval API test missing"
require 'TestBackupScheduleConfigurationValidation' internal/app/backup_scheduler_test.go "backup schedule validation tests missing"
require 'maxSafePageNumber' internal/app/pagination_config.go "safe page upper bound missing"
require 'TestMaximumPageOffsetDoesNotOverflowInt' internal/app/config_management_test.go "offset overflow test definition missing"

require 'CREATE TABLE IF NOT EXISTS document_files' migrations/004_required_entities_and_backup_schedule.sql "document_files entity missing"
require 'CREATE TABLE IF NOT EXISTS redaction_confirmations' migrations/004_required_entities_and_backup_schedule.sql "redaction_confirmations entity missing"
require 'CREATE TABLE IF NOT EXISTS document_diffs' migrations/004_required_entities_and_backup_schedule.sql "document_diffs entity missing"
require 'insertDocumentFileWithExecutor' internal/app/documents.go "document file entity is not written on upload"
require 'insertDocumentFileWithExecutor' internal/app/redactions.go "redacted file entity is not written"
require 'redaction_confirmations' internal/app/redactions.go "redaction confirmation entity is not written"
require 'insertDocumentFileWithExecutor' internal/app/bates_version.go "Bates file entity is not written"
require 'document_diffs' internal/app/workflows.go "document diff entity is not written"
require 'DOCUMENT_DIFF_CREATE' internal/app/workflows.go "document diff persistence is not audited"
require 'document_files' internal/repository/backup.go "backup snapshots omit document file entities"
require 'redaction_confirmations' internal/repository/backup.go "backup snapshots omit redaction confirmations"
require 'document_diffs' internal/repository/backup.go "backup snapshots omit document diffs"

require 'api.PageCountFile(path)' internal/platform/pdf.go "PDF page tree parser missing"
if grep -Fq 'bytes.Count' internal/platform/pdf.go; then
  fail "PDF page count still uses byte substring counting"
fi
require 'TestInspectPDFPageBoundaries' internal/platform/pdf_test.go "0/1/499/500/501 PDF boundary tests missing"
require 'TestInspectPDFReadsPageFromCompressedObjectStream' internal/platform/pdf_test.go "compressed object stream page test missing"
require 'TestInspectPDFIgnoresPagesRootAndCompressedStreamTokens' internal/platform/pdf_test.go "Pages root/token false-positive test missing"

require 'TestVersionLimitAllowsFortyNineToFifty' internal/app/version_limit_test.go "49 to 50 version boundary missing"
require 'TestVersionLimitRejectsFiftyToFiftyOne' internal/app/version_limit_test.go "50 to 51 version rejection missing"
require 'nextDocumentVersion' internal/app/redactions.go "redaction does not use shared version ceiling"
require 'nextDocumentVersion' internal/app/bates_version.go "Bates does not use shared version ceiling"
if test "$(grep -n 'nextDocumentVersion' internal/app/redactions.go | head -1 | cut -d: -f1)" -ge "$(grep -n 'ApplyRedactionBurnIn' internal/app/redactions.go | head -1 | cut -d: -f1)"; then
  fail "redaction version limit is checked after file generation"
fi
if test "$(grep -n 'nextDocumentVersion' internal/app/bates_version.go | head -1 | cut -d: -f1)" -ge "$(grep -n 'AllocateBatesRange' internal/app/bates_version.go | head -1 | cut -d: -f1)"; then
  fail "Bates version limit is checked after sequence allocation"
fi

require '59-second-old timestamp is accepted' tests/api/test_request_guard_edges.sh "fresh timestamp middleware test missing"
require '61-second-old timestamp is rejected' tests/api/test_request_guard_edges.sh "expired timestamp middleware test missing"
require '61-second-future timestamp is rejected' tests/api/test_request_guard_edges.sh "future timestamp middleware test missing"
require 'same JWT rejects duplicate request id' tests/api/test_request_guard_edges.sh "same-token replay rejection test missing"
require 'request id scope permits a different JWT' tests/api/test_request_guard_edges.sh "cross-token replay scope test missing"

for phrase in \
  'rejects rollback' \
  'rejects redaction proposal' \
  'rejects redaction confirmation' \
  'rejects annotation creation' \
  'rejects annotation disposition' \
  'rejects Bates mutation' \
  'rejects workflow transition' \
  'rejects repeated finalization'; do
  require "$phrase" tests/api/test_finalized_immutability.sh "Finalized matrix missing: $phrase"
done
require '"versions:$BASE_VERSIONS:$AFTER_VERSIONS"' tests/api/test_finalized_immutability.sh "Finalized version side-effect snapshot missing"
require '"audits:$BASE_AUDITS:$AFTER_AUDITS"' tests/api/test_finalized_immutability.sh "Finalized audit side-effect snapshot missing"
require '"notifications:$BASE_NOTIFICATIONS:$AFTER_NOTIFICATIONS"' tests/api/test_finalized_immutability.sh "Finalized notification side-effect snapshot missing"
require 'finalized denials leave $name unchanged' tests/api/test_finalized_immutability.sh "Finalized unchanged-count assertion missing"

require 'PGPASSFILE=' internal/platform/postgres_command.go "PGPASSFILE command authentication missing"
require 'file.Chmod(0600)' internal/platform/postgres_command.go "PGPASSFILE mode is not restricted"
require 'pgDumpCommandArgs' internal/platform/backup_exec.go "pg_dump does not use password-free argument builder"
require 'pgRestoreCommandArgs' internal/platform/backup_exec.go "pg_restore does not use password-free argument builder"
if grep -E 'exec\.Command\("pg_(dump|restore)".*DSN|--dbname.*,.*Password|password=' internal/platform/backup_exec.go; then
  fail "database password may still be placed in postgres subprocess arguments"
fi
require 'TestPostgresCommandArgumentsExcludePassword' internal/platform/postgres_command_test.go "postgres argv exposure test definition missing"
require 'TestPGPassFileUsesRestrictedModeAndEscaping' internal/platform/postgres_command_test.go "PGPASSFILE permission test definition missing"

require 'TestAPIRequestsIncludeAuthenticationStateInBarrier' internal/app/operation_barrier_test.go "authenticated GET/write barrier definition missing"
require 'TestBackupBarrierCoversUploadRedactionAndBatesRequests' internal/app/operation_barrier_test.go "backup concurrency mutation coverage definition missing"
require 'TestMaintenanceRejectsOrdinaryAndConcurrentRestoreRequests' internal/app/operation_barrier_test.go "maintenance denial test definition missing"
require 'TestRestoreAdmissionRejectsSecondAuthenticationPath' internal/app/operation_barrier_test.go "restore admission test definition missing"
require 'TestBackupRestoreAndResolutionUseExclusiveOperationPaths' internal/app/operation_barrier_test.go "exclusive operation classification test missing"
require 'TestRequestedRestoreBecomesInterruptedNotFailed' internal/app/operation_barrier_test.go "interrupted outcome test definition missing"
require 'TestRestoreSuccessBeforeTerminalJournalRequiresInterruptedResolution' internal/app/operation_barrier_test.go "post-platform pre-terminal crash definition missing"
require 'backup.local_volume' tests/api/test_admin_ops.sh "deployment-owned config API test definition missing"
require 'pagination.default_page_size' tests/api/test_admin_ops.sh "pagination config API test definition missing"

require 'application mutation barrier' docs/backup-recovery.md "backup barrier documentation missing"
require 'Interrupted' docs/backup-recovery.md "interrupted restore documentation missing"
require 'PGPASSFILE' docs/security.md "postgres credential boundary documentation missing"

echo "PASS: backup/restore integrity static contract"
