package app

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
)

func (a *App) recordRestoreState(c echo.Context, id string, status string, action string, metadata map[string]interface{}) error {
	p := principal(c)
	targetPath, _ := metadata["database_dump_path"].(string)
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'restore',$2,$3,$4,NOW()) ON CONFLICT(id) DO UPDATE SET status=excluded.status,target_path=excluded.target_path,created_by=excluded.created_by`, id, status, targetPath, p.UserID); err != nil {
		return err
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, action, "", metadata); err != nil {
		return err
	}
	return tx.Commit()
}

func (a *App) restoreBackup(c echo.Context) error {
	var req struct {
		DatabaseDumpPath string `json:"database_dump_path"`
		FileSnapshotPath string `json:"file_snapshot_path"`
	}
	if err := c.Bind(&req); err != nil || req.DatabaseDumpPath == "" || req.FileSnapshotPath == "" {
		return apiErr(c, http.StatusBadRequest, "INVALID_RESTORE_REQUEST", "database_dump_path and file_snapshot_path are required")
	}
	id := makeIdentifier("rst")
	baseMetadata := map[string]interface{}{
		"restore_id":          id,
		"database_dump_path": req.DatabaseDumpPath,
		"file_snapshot_path": req.FileSnapshotPath,
	}
	if err := a.recordRestoreState(c, id, "Requested", "BACKUP_RESTORE_REQUESTED", baseMetadata); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not record restore request")
	}
	result, err := platform.RunRestoreArtifactsStrict(a.cfg.DSN(), req.DatabaseDumpPath, req.FileSnapshotPath, a.cfg.StorageDir)
	if err != nil {
		failed := map[string]interface{}{
			"restore_id":          id,
			"database_dump_path": req.DatabaseDumpPath,
			"file_snapshot_path": req.FileSnapshotPath,
			"result":              result,
			"error":               err.Error(),
		}
		if stateErr := a.recordRestoreState(c, id, "Failed", "BACKUP_RESTORE_FAILED", failed); stateErr != nil {
			return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "restore failed and its failure state could not be recorded")
		}
		return apiErr(c, http.StatusBadRequest, "RESTORE_ARTIFACT_REQUIRED", "strict restore requires readable database and filesystem artifacts")
	}
	completed := map[string]interface{}{
		"restore_id":          id,
		"database_dump_path": req.DatabaseDumpPath,
		"file_snapshot_path": req.FileSnapshotPath,
		"result":              result,
	}
	if err := a.recordRestoreState(c, id, "Completed", "BACKUP_RESTORE_COMPLETED", completed); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "restore completed but its completion audit could not be recorded")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"id": id, "status": "Restored", "result": result})
}
