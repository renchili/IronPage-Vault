package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ironpage-vault/internal/platform"
	"ironpage-vault/internal/repository"
	"ironpage-vault/internal/service"
)

const backupSchedulePollInterval = time.Minute

type backupSchedule struct {
	Enabled  bool
	Interval time.Duration
}

func backupInterval(value string) time.Duration {
	interval, err := time.ParseDuration(value)
	if err != nil || interval < minimumBackupInterval || interval > maximumBackupInterval {
		return 0
	}
	return interval
}

func (a *App) loadBackupSchedule(ctx context.Context) (backupSchedule, error) {
	rows := []paginationConfigRow{}
	if err := a.db.SelectContext(ctx, &rows, `SELECT key,value FROM config_entries WHERE key IN ($1,$2)`, backupScheduleEnabledKey, backupScheduleIntervalKey); err != nil {
		return backupSchedule{}, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	enabledRaw, enabledOK := values[backupScheduleEnabledKey]
	intervalRaw, intervalOK := values[backupScheduleIntervalKey]
	if !enabledOK || !intervalOK {
		return backupSchedule{}, fmt.Errorf("backup schedule configuration is incomplete")
	}
	enabled, err := strconv.ParseBool(enabledRaw)
	if err != nil {
		return backupSchedule{}, fmt.Errorf("invalid persisted backup schedule enabled value: %w", err)
	}
	interval := backupInterval(intervalRaw)
	if interval == 0 {
		return backupSchedule{}, fmt.Errorf("invalid persisted backup interval")
	}
	return backupSchedule{Enabled: enabled, Interval: interval}, nil
}

func (a *App) scheduledBackupDue(ctx context.Context, interval time.Duration) (bool, error) {
	var lastCompleted sql.NullTime
	if err := a.db.GetContext(ctx, &lastCompleted, `SELECT MAX(created_at) FROM backup_jobs WHERE kind='scheduled_full_backup' AND status='Completed'`); err != nil {
		return false, err
	}
	if !lastCompleted.Valid {
		return true, nil
	}
	return !time.Now().UTC().Before(lastCompleted.Time.UTC().Add(interval)), nil
}

func (a *App) runScheduledBackupIfDue(ctx context.Context) error {
	schedule, err := a.loadBackupSchedule(ctx)
	if err != nil {
		return err
	}
	if !schedule.Enabled {
		return nil
	}
	due, err := a.scheduledBackupDue(ctx, schedule.Interval)
	if err != nil || !due {
		return err
	}
	return a.runScheduledBackup(ctx)
}

func (a *App) startBackupScheduler() {
	go func() {
		if err := a.runScheduledBackupIfDue(context.Background()); err != nil {
			log.Printf("scheduled backup evaluation failed: %v", err)
		}
		ticker := time.NewTicker(backupSchedulePollInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := a.runScheduledBackupIfDue(context.Background()); err != nil {
				log.Printf("scheduled backup evaluation failed: %v", err)
			}
		}
	}()
}

func (a *App) runScheduledBackup(ctx context.Context) error {
	if a.operations == nil {
		return fmt.Errorf("scheduled backup operation barrier is unavailable")
	}
	return a.operations.withExclusiveOperation(ctx, func() error {
		return a.runScheduledBackupLocked(ctx)
	})
}

func (a *App) runScheduledBackupLocked(ctx context.Context) error {
	id := makeIdentifier("bak")
	if err := os.MkdirAll(a.cfg.BackupDir, 0750); err != nil {
		return err
	}
	counts, err := repository.New(a.db).CountBackupTables(ctx)
	if err != nil {
		return err
	}
	snapshot := service.NewBackupSnapshot(id, a.cfg.DBName, counts)
	artifacts, err := platform.RunBackupArtifactsStrict(id, a.cfg.PostgresCommandConfig(), a.cfg.StorageDir, a.cfg.BackupDir)
	target := filepath.Join(a.cfg.BackupDir, id+".json")
	if err != nil {
		cleanupBackupArtifacts(a.cfg.BackupDir, id, target, artifacts)
		return err
	}
	committed := false
	defer func() {
		if !committed {
			cleanupBackupArtifacts(a.cfg.BackupDir, id, target, artifacts)
		}
	}()
	if err := platform.WriteBackupMetadataSnapshot(target, snapshot); err != nil {
		return err
	}
	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'scheduled_full_backup','Completed',$2,$3,NOW())`, id, target, systemPrincipalID); err != nil {
		return err
	}
	if err := a.insertAuditRecordWithExecutor(ctx, tx, systemPrincipalID, "SCHEDULED_BACKUP_CREATE", "", "scheduler", "scheduler", map[string]interface{}{"backup_id": id, "database_dump_path": artifacts.DatabaseDumpPath, "file_snapshot_path": artifacts.FileSnapshotPath, "recovery_boundary": "exclusive_application_mutation_barrier"}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
