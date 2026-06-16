package app

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
	"ironpage-vault/internal/repository"
	"ironpage-vault/internal/service"
)

func (a *App) collectBackupSnapshot(c echo.Context, id string) (platform.BackupMetadataSnapshot, error) {
	counts, err := repository.New(a.db).CountBackupTables(c.Request().Context())
	if err != nil {
		return platform.BackupMetadataSnapshot{}, err
	}
	return service.NewBackupSnapshot(id, a.cfg.DBName, counts), nil
}

func (a *App) runBackupMetadataSnapshot(c echo.Context) error {
	p := principal(c)
	id := makeIdentifier("bak")
	if err := os.MkdirAll(a.cfg.BackupDir, 0750); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_DIR_ERROR", "could not create backup directory")
	}
	target := filepath.Join(a.cfg.BackupDir, id+".json")
	snapshot, err := a.collectBackupSnapshot(c, id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_SNAPSHOT_ERROR", "could not collect backup metadata snapshot")
	}
	artifacts, err := platform.RunBackupArtifactsStrict(id, a.cfg.DSN(), a.cfg.StorageDir, a.cfg.BackupDir)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_ARTIFACT_ERROR", "strict backup artifacts were not completed")
	}
	if err := platform.WriteBackupMetadataSnapshot(target, snapshot); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_WRITE_ERROR", "could not write backup metadata snapshot file")
	}
	err = repository.New(a.db).InsertBackupJob(c.Request().Context(), id, target, p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "could not record backup job")
	}
	a.audit(c, p.UserID, "BACKUP_CREATE", "", map[string]interface{}{"database_dump_mode": artifacts.DatabaseDumpMode, "file_snapshot_mode": artifacts.FileSnapshotMode})
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "status": "Completed", "target_path": target, "kind": "full_backup", "created_at": snapshot.CreatedAt, "restore_supported": artifacts.RestoreSupported, "artifacts": artifacts})
}

func (a *App) runBackupFile(c echo.Context) error {
	return a.runBackupMetadataSnapshot(c)
}
