package app

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/core"
	"ironpage-vault/internal/platform"
)

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
	var d Document
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if !canTransitionDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this principal transition scope")
	}
	if d.Status == StatusFinalized {
		return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable")
	}
	if req.Status != nextWorkflowStatus(d.Status) {
		return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_TRANSITION", "status must follow the configured chain")
	}
	if req.Status == StatusFinalized && p.Role != RoleEditor {
		return apiErr(c, http.StatusForbidden, "EDITOR_REQUIRED", "only editor may finalize")
	}
	if p.Role != RoleReviewer && p.Role != RoleEditor {
		return apiErr(c, http.StatusForbidden, "FORBIDDEN", "role cannot transition documents")
	}
	_, err := a.db.ExecContext(c.Request().Context(), `UPDATE documents SET status=$1, finalized_at=CASE WHEN $1='Finalized' THEN NOW() ELSE finalized_at END, updated_at=NOW() WHERE id=$2`, req.Status, d.ID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_UPDATE_ERROR", "workflow transition failed")
	}
	_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO document_status_history(id,document_id,from_status,to_status,changed_by,created_at) VALUES($1,$2,$3,$4,$5,NOW())`, makeIdentifier("wfh"), d.ID, d.Status, req.Status, p.UserID)
	_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'WORKFLOW_TRANSITION',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, d.ID, currentRequestID(c), c.RealIP())
	a.notifyUser(c, d.OwnerID, "workflow.transition", "Document moved to "+req.Status, d.ID)
	return c.JSON(http.StatusOK, map[string]interface{}{"status": req.Status, "changed_at": time.Now().UTC()})
}

func (a *App) finalizeDocument(c echo.Context) error {
	p := principal(c)
	var d Document
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if !canEditDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this editor scope")
	}
	if d.Status == StatusFinalized {
		return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable")
	}
	if d.Status != StatusApproved {
		return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_TRANSITION", "document must be Approved before Finalized")
	}
	_, err := a.db.ExecContext(c.Request().Context(), `UPDATE documents SET status='Finalized', finalized_at=NOW(), updated_at=NOW() WHERE id=$1`, d.ID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "FINALIZE_ERROR", "finalize failed")
	}
	_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'DOCUMENT_FINALIZE',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, d.ID, currentRequestID(c), c.RealIP())
	return c.JSON(http.StatusOK, map[string]interface{}{"status": StatusFinalized})
}

func versionComparisonResult(left DocumentVersion, right DocumentVersion, leftRaw []byte, rightRaw []byte) map[string]interface{} {
	result := map[string]interface{}{"left_version_id": left.ID, "right_version_id": right.ID, "comparison_kind": "binary_metadata", "text_diff_supported": false, "bbox_supported": false, "same_sha256": left.FileSHA256 == right.FileSHA256, "same_size": left.SizeBytes == right.SizeBytes, "same_page_count": left.PageCount == right.PageCount, "byte_length_delta": len(rightRaw) - len(leftRaw), "added": []interface{}{}, "removed": []interface{}{}, "modified": []interface{}{}}
	if !bytes.Equal(leftRaw, rightRaw) {
		result["modified"] = []map[string]interface{}{{"page": 1, "bbox": map[string]int{"x": 0, "y": 0, "w": 0, "h": 0}, "text": "binary content differs between supplied versions"}}
	}
	return result
}

func versionTextComparisonResult(left DocumentVersion, right DocumentVersion, leftRaw []byte, rightRaw []byte) map[string]interface{} {
	result := versionComparisonResult(left, right, leftRaw, rightRaw)
	result["comparison_kind"] = "text_and_binary"
	result["text_diff_supported"] = true
	leftText, leftMode, leftErr := platform.ExtractPDFText(left.FilePath)
	rightText, rightMode, rightErr := platform.ExtractPDFText(right.FilePath)
	result["text_extract_left_mode"] = leftMode
	result["text_extract_right_mode"] = rightMode
	if leftErr != nil || rightErr != nil {
		result["text_diff_supported"] = false
		return result
	}
	if leftText != rightText {
		result["modified"] = []map[string]interface{}{{"page": 1, "bbox": map[string]int{"x": 0, "y": 0, "w": 0, "h": 0}, "text": "extracted text differs between supplied versions"}}
	}
	return result
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
	var left DocumentVersion
	var right DocumentVersion
	if err := a.db.GetContext(c.Request().Context(), &left, `SELECT * FROM document_versions WHERE id=$1`, req.LeftVersionID); err != nil {
		return apiErr(c, http.StatusNotFound, "LEFT_VERSION_NOT_FOUND", "left version not found")
	}
	if err := a.db.GetContext(c.Request().Context(), &right, `SELECT * FROM document_versions WHERE id=$1`, req.RightVersionID); err != nil {
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
	leftRaw, err := os.ReadFile(left.FilePath)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "LEFT_FILE_READ_ERROR", "could not read left version")
	}
	rightRaw, err := os.ReadFile(right.FilePath)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "RIGHT_FILE_READ_ERROR", "could not read right version")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": versionTextComparisonResult(left, right, leftRaw, rightRaw)})
}
