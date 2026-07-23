package app

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/core"
	"ironpage-vault/internal/service"
)

// nextWorkflowStatus preserves the default-chain compatibility helper used by
// package tests. Runtime transitions resolve the persisted Admin-managed chain.
func nextWorkflowStatus(current string) string {
	return core.NextWorkflowStatus(current)
}

func (a *App) transitionDocument(c echo.Context) error {
	p := principal(c)
	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil || strings.TrimSpace(req.Status) == "" {
		return apiErr(c, http.StatusBadRequest, "STATUS_REQUIRED", "target status is required")
	}
	req.Status = strings.TrimSpace(req.Status)
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(c.Request().Context(), `LOCK TABLE workflow_status_definitions IN SHARE MODE`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_LOCK_ERROR", "could not lock workflow definitions")
	}
	var d Document
	if err := tx.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1 FOR UPDATE`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if !canTransitionDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this principal transition scope")
	}
	if d.Status == StatusFinalized {
		return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable")
	}
	next, err := a.nextWorkflowDefinition(c.Request().Context(), tx, d.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_TRANSITION", "current status has no configured successor")
		}
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_QUERY_ERROR", "could not resolve configured workflow")
	}
	if req.Status != next.Name {
		return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_TRANSITION", "status must follow the configured chain")
	}
	if next.Name == StatusFinalized && p.Role != RoleEditor {
		return apiErr(c, http.StatusForbidden, "EDITOR_REQUIRED", "only editor may finalize")
	}
	if p.Role != RoleReviewer && p.Role != RoleEditor {
		return apiErr(c, http.StatusForbidden, "FORBIDDEN", "role cannot transition documents")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `UPDATE documents SET status=$1, finalized_at=CASE WHEN $1='Finalized' THEN NOW() ELSE finalized_at END, updated_at=NOW() WHERE id=$2`, req.Status, d.ID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_UPDATE_ERROR", "workflow transition failed")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO document_status_history(id,document_id,from_status,to_status,changed_by,created_at) VALUES($1,$2,$3,$4,$5,NOW())`, makeIdentifier("wfh"), d.ID, d.Status, req.Status, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_HISTORY_ERROR", "could not record workflow history")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "WORKFLOW_TRANSITION", d.ID, map[string]interface{}{"from_status": d.Status, "to_status": req.Status}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record workflow audit")
	}
	if err := a.notifyUserWithExecutor(c, tx, d.OwnerID, "workflow.transition", "Document moved to "+req.Status, d.ID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_CREATE_ERROR", "could not record workflow notification")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit workflow transition")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": req.Status, "changed_at": time.Now().UTC()})
}

func (a *App) finalizeDocument(c echo.Context) error {
	p := principal(c)
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(c.Request().Context(), `LOCK TABLE workflow_status_definitions IN SHARE MODE`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_LOCK_ERROR", "could not lock workflow definitions")
	}
	var d Document
	if err := tx.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1 FOR UPDATE`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if !canEditDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this editor scope")
	}
	if d.Status == StatusFinalized {
		return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable")
	}
	next, err := a.nextWorkflowDefinition(c.Request().Context(), tx, d.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_TRANSITION", "document is not at the configured pre-final status")
		}
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_QUERY_ERROR", "could not resolve configured workflow")
	}
	if next.Name != StatusFinalized || next.Mutable {
		return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_TRANSITION", "document must be at the configured status immediately before Finalized")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `UPDATE documents SET status='Finalized', finalized_at=NOW(), updated_at=NOW() WHERE id=$1`, d.ID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "FINALIZE_ERROR", "finalize failed")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO document_status_history(id,document_id,from_status,to_status,changed_by,created_at) VALUES($1,$2,$3,'Finalized',$4,NOW())`, makeIdentifier("wfh"), d.ID, d.Status, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_HISTORY_ERROR", "could not record finalization history")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "DOCUMENT_FINALIZE", d.ID, map[string]interface{}{"from_status": d.Status, "to_status": StatusFinalized}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record finalization audit")
	}
	if err := a.notifyUserWithExecutor(c, tx, d.OwnerID, "workflow.transition", "Document moved to "+StatusFinalized, d.ID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_CREATE_ERROR", "could not record finalization notification")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit finalization")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": StatusFinalized})
}

func versionTextComparisonResult(left DocumentVersion, right DocumentVersion) map[string]interface{} {
	return service.CompareVersionFiles(
		service.VersionFile{ID: left.ID, FilePath: left.FilePath, SHA256: left.FileSHA256, SizeBytes: left.SizeBytes, PageCount: left.PageCount},
		service.VersionFile{ID: right.ID, FilePath: right.FilePath, SHA256: right.FileSHA256, SizeBytes: right.SizeBytes, PageCount: right.PageCount},
	)
}

func (a *App) compareVersions(c echo.Context) error {
	p := principal(c)
	var req struct {
		LeftVersionID  string `json:"left_version_id"`
		RightVersionID string `json:"right_version_id"`
	}
	if err := c.Bind(&req); err != nil || req.LeftVersionID == "" || req.RightVersionID == "" {
		return apiErr(c, http.StatusBadRequest, "VERSION_IDS_REQUIRED", "left_version_id and right_version_id are required")
	}
	versionQuery := `SELECT version.id,version.document_id,version.version_number,file.file_path,file.file_sha256,file.size_bytes,file.page_count,version.created_by,version.created_at FROM document_versions AS version JOIN document_files AS file ON file.version_id=version.id WHERE version.id=$1`
	var left DocumentVersion
	var right DocumentVersion
	if err := a.db.GetContext(c.Request().Context(), &left, versionQuery, req.LeftVersionID); err != nil {
		return apiErr(c, http.StatusNotFound, "LEFT_VERSION_NOT_FOUND", "left version not found")
	}
	if err := a.db.GetContext(c.Request().Context(), &right, versionQuery, req.RightVersionID); err != nil {
		return apiErr(c, http.StatusNotFound, "RIGHT_VERSION_NOT_FOUND", "right version not found")
	}
	var leftDoc Document
	var rightDoc Document
	if err := a.db.GetContext(c.Request().Context(), &leftDoc, `SELECT * FROM documents WHERE id=$1`, left.DocumentID); err != nil {
		return apiErr(c, http.StatusNotFound, "LEFT_DOCUMENT_NOT_FOUND", "left document not found")
	}
	if err := a.db.GetContext(c.Request().Context(), &rightDoc, `SELECT * FROM documents WHERE id=$1`, right.DocumentID); err != nil {
		return apiErr(c, http.StatusNotFound, "RIGHT_DOCUMENT_NOT_FOUND", "right document not found")
	}
	if !canReadDocumentObject(p, leftDoc) || !canReadDocumentObject(p, rightDoc) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "version comparison is outside this principal scope")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": versionTextComparisonResult(left, right)})
}
