#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

check() {
  local name="$1"
  local cmd="$2"
  if bash -lc "$cmd" >/tmp/ironpage_scheduled_backup_contract.out 2>&1; then
    echo "PASS scheduled-backup: $name"
    PASS=$((PASS+1))
  else
    echo "FAIL scheduled-backup: $name"
    cat /tmp/ironpage_scheduled_backup_contract.out
    FAIL=$((FAIL+1))
  fi
}

check "startup invokes scheduler" "grep -q 'a.startBackupScheduler()' internal/app/server.go"
check "schedule is PostgreSQL persisted" "grep -q 'backup.schedule_enabled' migrations/004_required_entities_and_backup_schedule.sql && grep -q 'backup.interval' migrations/004_required_entities_and_backup_schedule.sql"
check "scheduler has no unreachable environment switch" "! grep -q 'BACKUP_INTERVAL' internal/app/backup_scheduler.go docker-compose.yml scripts/deploy.sh scripts/entrypoint.sh"
check "scheduler reloads persisted configuration" "grep -q 'loadBackupSchedule' internal/app/backup_scheduler.go && grep -q 'SELECT key,value FROM config_entries' internal/app/backup_scheduler.go"
check "scheduler validates interval bounds" "grep -q 'minimumBackupInterval' internal/app/config_management.go && grep -q 'maximumBackupInterval' internal/app/config_management.go"
check "scheduler polls for configuration changes" "grep -q 'backupSchedulePollInterval' internal/app/backup_scheduler.go && grep -q 'time.NewTicker' internal/app/backup_scheduler.go"
check "scheduler evaluates startup and ticks" "test \"$(grep -c 'runScheduledBackupIfDue(context.Background())' internal/app/backup_scheduler.go)\" -ge 2"
check "scheduler checks last completed run" "grep -q 'MAX(created_at).*scheduled_full_backup' internal/app/backup_scheduler.go"
check "worker counts backup tables" "grep -q 'CountBackupTables' internal/app/backup_scheduler.go"
check "worker uses strict artifacts" "grep -q 'RunBackupArtifactsStrict' internal/app/backup_scheduler.go"
check "worker writes metadata snapshot" "grep -q 'WriteBackupMetadataSnapshot' internal/app/backup_scheduler.go"
check "worker records scheduled job" "grep -q 'scheduled_full_backup' internal/app/backup_scheduler.go"
check "worker records scheduler audit" "grep -q 'SCHEDULED_BACKUP_CREATE' internal/app/backup_scheduler.go"
check "interval parser has boundary tests" "grep -q 'TestBackupIntervalParsing' internal/app/backup_interval_test.go && grep -q 'maximum' internal/app/backup_interval_test.go"
check "Admin API tests backup schedule" "grep -q 'backup.schedule_enabled' tests/api/test_admin_ops.sh && grep -q 'backup.interval' tests/api/test_admin_ops.sh"

TOTAL=$((PASS+FAIL))
echo "SCHEDULED BACKUP CONTRACT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
