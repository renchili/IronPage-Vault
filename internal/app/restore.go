package app

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
)

func (a *App) restoreBackup(c echo.Context) error {
	p := principal(c)
	var req struct {
		DatabaseDumpPath string `json:"database_dump_path"`
		FileSnapshotPath string `json:"file_snapshot_path"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_RESTORE_REQUEST", "invalid restore request")
	}
	result, err := platform.RunRestoreArtifacts(a.cfg.DSN(), req.DatabaseDumpPath, req.FileSnapshotPath, a.cfg.StorageDir)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_ERROR", "restore failed")
	}
	a.audit(c, p.UserID, "BACKUP_RESTORE", "", map[string]interface{}{"database_dump_path": req.DatabaseDumpPath, "file_snapshot_path": req.FileSnapshotPath})
	return c.JSON(http.StatusAccepted, map[string]interface{}{"status": "RestoreRequested", "result": result})
}
