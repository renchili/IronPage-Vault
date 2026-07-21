package app

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"ironpage-vault/internal/platform"
	"ironpage-vault/internal/repository"
	"ironpage-vault/internal/service"
)

func backupInterval(value string) time.Duration {
	if value == "" {
		return 0
	}
	interval, err := time.ParseDuration(value)
	if err != nil || interval <= 0 {
		return 0
	}
	return interval
}

func (a *App) startBackupScheduler() {
	interval := backupInterval(os.Getenv("BACKUP_INTERVAL"))
	if interval == 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := a.runScheduledBackup(context.Background()); err != nil {
				log.Printf("scheduled backup failed: %v", err)
			}
		}
	}()
}

func (a *App) runScheduledBackup(ctx context.Context) error {
	id := makeIdentifier("bak")
	if err := os.MkdirAll(a.cfg.BackupDir, 0750); err != nil {
		return err
	}
	counts, err := repository.New(a.db).CountBackupTables(ctx)
	if err != nil {
		return err
	}
	snapshot := service.NewBackupSnapshot(id, a.cfg.DBName, counts)
	artifacts, err := platform.RunBackupArtifactsStrict(id, a.cfg.DSN(), a.cfg.StorageDir, a.cfg.BackupDir)
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
	if _, err := tx.ExecContext(ctx, `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'scheduled_full_backup','Completed',$2,NULL,NOW())`, id, target); err != nil {
		return err
	}
	if err := a.insertAuditRecordWithExecutor(ctx, tx, "", "SCHEDULED_BACKUP_CREATE", "", "scheduler", "scheduler", map[string]interface{}{"backup_id": id, "database_dump_path": artifacts.DatabaseDumpPath, "file_snapshot_path": artifacts.FileSnapshotPath}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
