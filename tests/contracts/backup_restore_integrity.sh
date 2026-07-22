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

require 'pg_advisory_lock_shared($1)' internal/app/operation_barrier.go "shared mutation advisory lock missing"
require 'pg_advisory_lock($1)' internal/app/operation_barrier.go "exclusive backup/restore advisory lock missing"
require 'e.Use(a.maintenanceMiddleware)' internal/app/server.go "maintenance middleware not wired"
require 'e.Use(a.mutationBarrierMiddleware)' internal/app/server.go "mutation barrier middleware not wired"
require 'withExclusiveOperation(c.Request().Context()' internal/app/backup_file.go "manual backup does not own exclusive barrier"
require 'withExclusiveOperation(ctx' internal/app/backup_scheduler.go "scheduled backup does not own exclusive barrier"
require 'withMaintenanceOperation(c.Request().Context()' internal/app/operation_barrier.go "restore middleware does not own maintenance barrier"
require 'admin.POST("/backup/restore", a.restoreBackup, a.restoreMaintenanceMiddleware)' internal/app/server.go "restore maintenance is not applied after Admin route authorization"
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
if grep -Fq 'NULLIF($2' internal/app/domain_events.go; then
  fail "audit actor can still be converted to NULL"
fi

require 'CONFIG_KEY_READ_ONLY' internal/app/admin.go "deployment-owned config rejection missing"
require 'CONFIG_KEY_NOT_MANAGED' internal/app/admin.go "unknown config key rejection missing"
require 'SELECT key,value FROM config_entries WHERE key IN ($1,$2) FOR UPDATE' internal/app/config_management.go "pagination pair is not locked for validation"
require 'validatePaginationValues(defaultSize, maxSize)' internal/app/config_management.go "pagination pair validation missing"
require 'maxSafePageNumber' internal/app/pagination_config.go "safe page upper bound missing"
require 'TestMaximumPageOffsetDoesNotOverflowInt' internal/app/config_management_test.go "offset overflow test definition missing"

require 'PGPASSFILE=' internal/platform/postgres_command.go "PGPASSFILE command authentication missing"
require 'file.Chmod(0600)' internal/platform/postgres_command.go "PGPASSFILE mode is not restricted"
require 'pgDumpCommandArgs' internal/platform/backup_exec.go "pg_dump does not use password-free argument builder"
require 'pgRestoreCommandArgs' internal/platform/backup_exec.go "pg_restore does not use password-free argument builder"
if grep -E 'exec\.Command\("pg_(dump|restore)".*DSN|--dbname.*,.*Password|password=' internal/platform/backup_exec.go; then
  fail "database password may still be placed in postgres subprocess arguments"
fi
require 'TestPostgresCommandArgumentsExcludePassword' internal/platform/postgres_command_test.go "postgres argv exposure test definition missing"
require 'TestPGPassFileUsesRestrictedModeAndEscaping' internal/platform/postgres_command_test.go "PGPASSFILE permission test definition missing"

require 'TestMaintenanceRejectsOrdinaryAndConcurrentRestoreRequests' internal/app/operation_barrier_test.go "maintenance denial test definition missing"
require 'TestRestoreAdmissionRejectsSecondAuthenticationPath' internal/app/operation_barrier_test.go "restore admission test definition missing"
require 'TestRequestedRestoreBecomesInterruptedNotFailed' internal/app/operation_barrier_test.go "interrupted outcome test definition missing"
require 'backup.local_volume' tests/api/test_admin_ops.sh "deployment-owned config API test definition missing"
require 'pagination.default_page_size' tests/api/test_admin_ops.sh "pagination config API test definition missing"

require 'application mutation barrier' docs/backup-recovery.md "backup barrier documentation missing"
require 'Interrupted' docs/backup-recovery.md "interrupted restore documentation missing"
require 'PGPASSFILE' docs/security.md "postgres credential boundary documentation missing"

echo "PASS: backup/restore integrity static contract"
