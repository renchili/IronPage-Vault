package app

import (
	"net/http"
	"os"
	"strings"
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
	owner, _ := c.Get(maintenanceOwnerContextKey).(bool)
	if !owner {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_BARRIER_ERROR", "restore request did not acquire exclusive maintenance ownership")
	}
	return a.restoreBackupLocked(c)
}

func (a *App) restoreBackupLocked(c echo.Context) error {
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
		"maintenance_mode":     "exclusive_http_and_mutation_barrier",
	}
	record := restoreLifecycleRecord{
		ID:                 id,
		Status:             restoreStatusRequested,
		Action:             "BACKUP_RESTORE_REQUESTED",
		ActorUserID:        p.UserID,
		OutcomeActorUserID: p.UserID,
		RequestID:          currentRequestID(c),
		SourceIP:           c.RealIP(),
		RequestedMetadata:  cloneRestoreMetadata(baseMetadata),
		Metadata:           cloneRestoreMetadata(baseMetadata),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := a.writeRestoreLifecycleRecord(record); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not create durable restore lifecycle journal")
	}
	if err := a.recordRestoreState(c, record); err != nil {
		if cleanupErr := a.removeRestoreLifecycleRecord(id); cleanupErr != nil {
			return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not record restore request or clean its lifecycle journal")
		}
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not record restore request")
	}

	result, err := platform.RunRestoreArtifactsStrict(a.cfg.PostgresCommandConfig(), req.DatabaseDumpPath, req.FileSnapshotPath, a.cfg.StorageDir)
	if err != nil {
		failed := cloneRestoreMetadata(baseMetadata)
		failed["result"] = result
		failed["error"] = err.Error()
		record.Status = restoreStatusFailed
		record.Action = "BACKUP_RESTORE_FAILED"
		record.OutcomeActorUserID = p.UserID
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
		return apiErr(c, http.StatusBadRequest, "RESTORE_ARTIFACT_REQUIRED", "strict restore requires valid database and filesystem artifacts")
	}

	completed := cloneRestoreMetadata(baseMetadata)
	completed["result"] = result
	record.Status = restoreStatusCompleted
	record.Action = "BACKUP_RESTORE_COMPLETED"
	record.OutcomeActorUserID = p.UserID
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

func (a *App) resolveInterruptedRestore(c echo.Context) error {
	p := principal(c)
	var req struct {
		Status           string `json:"status"`
		VerificationNote string `json:"verification_note"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_RESTORE_RESOLUTION", "status and verification_note are required")
	}
	if req.Status != restoreStatusCompleted && req.Status != restoreStatusFailed {
		return apiErr(c, http.StatusBadRequest, "INVALID_RESTORE_RESOLUTION", "status must be Completed or Failed")
	}
	if strings.TrimSpace(req.VerificationNote) == "" {
		return apiErr(c, http.StatusBadRequest, "INVALID_RESTORE_RESOLUTION", "verification_note is required")
	}
	path, err := a.restoreLifecyclePath(c.Param("id"))
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_RESTORE_ID", "restore id is invalid")
	}
	record, err := a.readRestoreLifecycleRecord(path)
	if os.IsNotExist(err) {
		return apiErr(c, http.StatusNotFound, "RESTORE_RECONCILIATION_NOT_FOUND", "interrupted restore journal was not found")
	}
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_JOURNAL_ERROR", "interrupted restore journal could not be read safely")
	}
	if record.Status != restoreStatusInterrupted {
		return apiErr(c, http.StatusConflict, "RESTORE_NOT_INTERRUPTED", "only an Interrupted restore may be resolved")
	}
	record.Status = req.Status
	record.OutcomeActorUserID = p.UserID
	if req.Status == restoreStatusCompleted {
		record.Action = "BACKUP_RESTORE_RECONCILED_COMPLETED"
	} else {
		record.Action = "BACKUP_RESTORE_RECONCILED_FAILED"
	}
	record.Metadata = cloneRestoreMetadata(record.RequestedMetadata)
	record.Metadata["verification_note"] = strings.TrimSpace(req.VerificationNote)
	record.Metadata["resolved_by_user_id"] = p.UserID
	record.Metadata["previous_status"] = restoreStatusInterrupted
	record.UpdatedAt = time.Now().UTC()
	if err := a.writeRestoreLifecycleRecord(record); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not persist the restore resolution journal")
	}
	if err := a.recordRestoreState(c, record); err != nil {
		return apiErr(c, http.StatusInternalServerError, "RESTORE_STATE_ERROR", "could not persist the restore resolution")
	}
	response := map[string]interface{}{"id": record.ID, "status": record.Status, "verification_note": strings.TrimSpace(req.VerificationNote)}
	if err := a.removeRestoreLifecycleRecord(record.ID); err != nil {
		response["lifecycle_reconciliation_pending"] = true
	}
	return c.JSON(http.StatusOK, response)
}
