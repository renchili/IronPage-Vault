package app

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
)

func (a *App) recordRestoreState(c echo.Context, record restoreLifecycleRecord) error {
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := a.persistRestoreLifecycleRecord(c.Request().Context(), tx, record); err != nil {
		return err
	}
	return tx.Commit()
}

func (a *App) restoreBackup(c echo.Context) error {
	p := principal(c)
	var req struct {
		DatabaseDumpPath string `json:"database_dump_path"`
		FileSnapshotPath string `json:"file_snapshot_path"`
	}
	if err := c.Bind(&req); err != nil || req.DatabaseDumpPath == "" || req.FileSnapshotPath == "" {
		return apiErr(c, http.StatusBadRequest, "INVALID_RESTORE_REQUEST", "database_dump_path and file_snapshot_path are required")
	}
	id := makeIdentifier("rst")
	baseMetadata := map[string]interface{}{
		"restore_id":           id,
		"requested_by_user_id": p.UserID,
		"database_dump_path":   req.DatabaseDumpPath,
		"file_snapshot_path":   req.FileSnapshotPath,
	}
	record := restoreLifecycleRecord{
		ID:                id,
		Status:            "Requested",
		Action:            "BACKUP_RESTORE_REQUESTED",
		ActorUserID:       p.UserID,
		RequestID:         currentRequestID(c),
		SourceIP:          c.RealIP(),
		RequestedMetadata: cloneRestoreMetadata(baseMetadata),
		Metadata:          cloneRestoreMetadata(baseMetadata),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := a.writeRestoreLifecycleRecord(record); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not create durable restore lifecycle journal")
	}
	if err := a.recordRestoreState(c, record); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not record restore request")
	}

	result, err := platform.RunRestoreArtifactsStrict(a.cfg.DSN(), req.DatabaseDumpPath, req.FileSnapshotPath, a.cfg.StorageDir)
	if err != nil {
		failed := cloneRestoreMetadata(baseMetadata)
		failed["result"] = result
		failed["error"] = err.Error()
		record.Status = "Failed"
		record.Action = "BACKUP_RESTORE_FAILED"
		record.Metadata = failed
		record.UpdatedAt = time.Now().UTC()
		if journalErr := a.writeRestoreLifecycleRecord(record); journalErr != nil {
			return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "restore failed and its durable failure state could not be recorded")
		}
		if stateErr := a.recordRestoreState(c, record); stateErr != nil {
			return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "restore failed and its failure state could not be recorded")
		}
		if cleanupErr := a.removeRestoreLifecycleRecord(id); cleanupErr != nil {
			return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "restore failure was recorded but lifecycle journal cleanup failed")
		}
		return apiErr(c, http.StatusBadRequest, "RESTORE_ARTIFACT_REQUIRED", "strict restore requires readable database and filesystem artifacts")
	}

	completed := cloneRestoreMetadata(baseMetadata)
	completed["result"] = result
	record.Status = "Completed"
	record.Action = "BACKUP_RESTORE_COMPLETED"
	record.Metadata = completed
	record.UpdatedAt = time.Now().UTC()
	if err := a.writeRestoreLifecycleRecord(record); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "restore completed but its durable completion state could not be recorded")
	}
	if err := a.recordRestoreState(c, record); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "restore completed but its completion state and audit could not be recorded")
	}
	response := map[string]interface{}{"id": id, "status": "Restored", "result": result}
	if err := a.removeRestoreLifecycleRecord(id); err != nil {
		response["lifecycle_reconciliation_pending"] = true
	}
	return c.JSON(http.StatusOK, response)
}
