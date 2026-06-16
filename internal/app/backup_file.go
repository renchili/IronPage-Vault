package app

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
)

func (a *App) collectBackupSnapshot(c echo.Context, id string) (platform.BackupMetadataSnapshot, error) {
	counts := []platform.BackupTableCount{}
	for _, table := range platform.BackupSnapshotTables() {
		var n int
		if err := a.db.GetContext(c.Request().Context(), &n, "SELECT COUNT(*) FROM "+table); err != nil {
			return platform.BackupMetadataSnapshot{}, err
		}
		counts = append(counts, platform.BackupTableCount{Table: table, Count: n})
	}
	return platform.NewBackupMetadataSnapshot(id, a.cfg.DBName, counts, time.Now().UTC()), nil
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
	if err := platform.WriteBackupMetadataSnapshot(target, snapshot); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_WRITE_ERROR", "could not write backup metadata snapshot file")
	}
	_, err = a.db.ExecContext(c.Request().Context(), `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'metadata_snapshot','Completed',$2,$3,NOW())`, id, target, p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "could not record backup metadata snapshot job")
	}
	a.audit(c, p.UserID, "BACKUP_CREATE", "", nil)
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "status": "Completed", "target_path": target, "kind": "metadata_snapshot", "created_at": snapshot.CreatedAt, "restore_supported": false})
}

func (a *App) runBackupFile(c echo.Context) error {
	return a.runBackupMetadataSnapshot(c)
}
