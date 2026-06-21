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
check "scheduler is gated by interval env" "grep -q 'BACKUP_INTERVAL' internal/app/backup_scheduler.go"
check "scheduler parses positive duration" "grep -q 'time.ParseDuration' internal/app/backup_scheduler.go"
check "scheduler uses ticker" "grep -q 'time.NewTicker' internal/app/backup_scheduler.go"
check "scheduler calls worker on tick" "grep -q 'runScheduledBackup(context.Background())' internal/app/backup_scheduler.go"
check "worker counts backup tables" "grep -q 'CountBackupTables' internal/app/backup_scheduler.go"
check "worker uses strict artifacts" "grep -q 'RunBackupArtifactsStrict' internal/app/backup_scheduler.go"
check "worker writes metadata snapshot" "grep -q 'WriteBackupMetadataSnapshot' internal/app/backup_scheduler.go"
check "worker records scheduled job" "grep -q 'scheduled_full_backup' internal/app/backup_scheduler.go"
check "worker records scheduler audit" "grep -q 'SCHEDULED_BACKUP_CREATE' internal/app/backup_scheduler.go"
check "interval parser has unit test" "grep -q 'TestBackupIntervalParsing' internal/app/backup_interval_test.go"

TOTAL=$((PASS+FAIL))
echo "SCHEDULED BACKUP CONTRACT SUMMARY total=$TOTAL passed=$PASS failed=$FAIL"
if [ "$FAIL" -ne 0 ]; then exit 1; fi
