package app

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/service"
)

func (a *App) proposeRedaction(c echo.Context) error {
	p := principal(c)
	docID := c.Param("id")
	var req struct {
		Page   int     `json:"page"`
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
		Reason string  `json:"reason"`
	}
	if err := c.Bind(&req); err != nil || !IsValidRedactionRegion(req.Page, req.Width, req.Height) {
		return apiErr(c, http.StatusBadRequest, "INVALID_REDACTION_REGION", "page and positive coordinates are required")
	}
	reason, err := encryptString(a.cfg.AESKey, req.Reason)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt redaction reason")
	}
	xCipher, err := encryptString(a.cfg.AESKey, formatFloat(req.X))
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt redaction coordinate")
	}
	yCipher, err := encryptString(a.cfg.AESKey, formatFloat(req.Y))
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt redaction coordinate")
	}
	widthCipher, err := encryptString(a.cfg.AESKey, formatFloat(req.Width))
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt redaction coordinate")
	}
	heightCipher, err := encryptString(a.cfg.AESKey, formatFloat(req.Height))
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt redaction coordinate")
	}

	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	d, err := a.ensureMutableWithExecutor(c.Request().Context(), tx, docID, true)
	if err != nil {
		return mutableError(c, err)
	}
	if !canEditDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this editor scope")
	}
	id := makeIdentifier("red")
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO redaction_proposals(id,document_id,page,x,y,width,height,x_ciphertext,y_ciphertext,width_ciphertext,height_ciphertext,reason,status,created_by,created_at) VALUES($1,$2,$3,0,0,0,0,$4,$5,$6,$7,$8,'Staged',$9,NOW())`, id, docID, req.Page, xCipher, yCipher, widthCipher, heightCipher, reason, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_CREATE_ERROR", "could not stage redaction")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "REDACTION_PROPOSE", docID, map[string]interface{}{"redaction_id": id, "page": req.Page}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record redaction proposal audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit redaction proposal")
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "status": "Staged"})
}

type redactionListRow struct {
	ID         string `db:"id" json:"id"`
	DocumentID string `db:"document_id" json:"document_id"`
	Page       int    `db:"page" json:"page"`
	Status     string `db:"status" json:"status"`
	CreatedBy  string `db:"created_by" json:"created_by"`
	CreatedAt  string `db:"created_at" json:"created_at"`
}

func (a *App) listRedactions(c echo.Context) error {
	p := principal(c)
	var d Document
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if !canReadDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this principal scope")
	}
	rows := []redactionListRow{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,page,status,created_by,created_at::text AS created_at FROM redaction_proposals WHERE document_id=$1 ORDER BY created_at DESC`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_QUERY_ERROR", "could not list redactions")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) confirmRedaction(c echo.Context) error {
	p := principal(c)
	docID := c.Param("id")
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	d, err := a.ensureMutableWithExecutor(c.Request().Context(), tx, docID, true)
	if err != nil {
		return mutableError(c, err)
	}
	if !canEditDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this editor scope")
	}
	var v DocumentVersion
	if err := tx.GetContext(c.Request().Context(), &v, `SELECT * FROM document_versions WHERE document_id=$1 AND version_number=$2`, docID, d.CurrentVersion); err != nil {
		return apiErr(c, http.StatusNotFound, "VERSION_NOT_FOUND", "current version not found")
	}
	if d.CurrentVersion >= a.cfg.MaxVersions {
		return apiErr(c, http.StatusConflict, "VERSION_LIMIT_REACHED", "document already has 50 versions")
	}
	regions, err := a.confirmedRedactionRegionsWithExecutor(c.Request().Context(), tx, docID, c.Param("redaction_id"))
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok && he.Code == http.StatusNotFound {
			return apiErr(c, http.StatusNotFound, "REDACTION_NOT_FOUND", "staged redaction not found")
		}
		return apiErr(c, http.StatusInternalServerError, "REDACTION_QUERY_ERROR", "could not load redaction regions")
	}

	newVersion := d.CurrentVersion + 1
	dst := redactedVersionPath(v.FilePath, newVersion)
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(dst)
		}
	}()
	if _, err := service.ApplyRedactionBurnIn(v.FilePath, dst, regions); err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_BURNIN_ERROR", "could not create redacted binary")
	}
	info, err := InspectPDF(dst, a.cfg.MaxUploadBytes, a.cfg.MaxPDFPages)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_VERIFY_ERROR", "could not verify redacted binary")
	}
	verID := makeIdentifier("ver")
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO document_versions(id,document_id,version_number,file_path,file_sha256,size_bytes,page_count,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`, verID, docID, newVersion, dst, info.SHA256, info.Size, info.PageCount, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "VERSION_CREATE_ERROR", "could not create redacted version")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `UPDATE documents SET current_version=$1,status='Redaction Pending',updated_at=NOW() WHERE id=$2`, newVersion, docID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "DOCUMENT_UPDATE_ERROR", "could not activate redacted version")
	}
	result, err := tx.ExecContext(c.Request().Context(), `UPDATE redaction_proposals SET status='Confirmed',confirmed_by=$1,confirmed_at=NOW() WHERE id=$2 AND document_id=$3 AND status='Staged'`, p.UserID, c.Param("redaction_id"), docID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_UPDATE_ERROR", "could not confirm redaction")
	}
	changed, err := result.RowsAffected()
	if err != nil || changed != 1 {
		return apiErr(c, http.StatusConflict, "REDACTION_STATE_ERROR", "redaction was not in a confirmable state")
	}
	if d.Status != StatusRedactionPending {
		if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO document_status_history(id,document_id,from_status,to_status,changed_by,created_at) VALUES($1,$2,$3,$4,$5,NOW())`, makeIdentifier("wfh"), docID, d.Status, StatusRedactionPending, p.UserID); err != nil {
			return apiErr(c, http.StatusInternalServerError, "WORKFLOW_HISTORY_ERROR", "could not record redaction status history")
		}
		if err := a.notifyUserWithExecutor(c, tx, d.OwnerID, "workflow.transition", "Document moved to " + StatusRedactionPending, docID); err != nil {
			return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_CREATE_ERROR", "could not record redaction status notification")
		}
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "REDACTION_CONFIRM", docID, map[string]interface{}{"redaction_id": c.Param("redaction_id"), "version": newVersion, "file_sha256": info.SHA256}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record redaction audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit redaction")
	}
	committed = true
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "Confirmed", "version": newVersion})
}
