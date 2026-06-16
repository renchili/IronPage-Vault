package app

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
)

func (a *App) ensureMutable(c echo.Context, docID string) (Document, error) {
	var d Document
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, docID); err != nil {
		if err == sql.ErrNoRows {
			return d, echo.NewHTTPError(http.StatusNotFound, "document not found")
		}
		return d, err
	}
	if d.Status == StatusFinalized {
		return d, echo.NewHTTPError(http.StatusConflict, "finalized")
	}
	return d, nil
}

func mutableError(c echo.Context, err error) error {
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code == http.StatusNotFound {
			return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
		}
		if he.Code == http.StatusConflict {
			return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable")
		}
	}
	return apiErr(c, http.StatusInternalServerError, "DOCUMENT_STATE_ERROR", "could not verify document state")
}

func (a *App) proposeRedaction(c echo.Context) error {
	p := principal(c)
	docID := c.Param("id")
	d, err := a.ensureMutable(c, docID)
	if err != nil {
		return mutableError(c, err)
	}
	if !canEditDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this editor scope")
	}
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
	id := makeIdentifier("red")
	_, err = a.db.ExecContext(c.Request().Context(), `INSERT INTO redaction_proposals(id,document_id,page,x,y,width,height,reason,status,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'Staged',$9,NOW())`, id, docID, req.Page, req.X, req.Y, req.Width, req.Height, reason, p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_CREATE_ERROR", "could not stage redaction")
	}
	a.audit(c, p.UserID, "REDACTION_PROPOSE", docID, nil)
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "status": "Staged"})
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
	rows := []map[string]interface{}{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,page,x,y,width,height,reason,status,created_by,created_at FROM redaction_proposals WHERE document_id=$1 ORDER BY created_at DESC`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_QUERY_ERROR", "could not list redactions")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) confirmRedaction(c echo.Context) error {
	p := principal(c)
	docID := c.Param("id")
	d, err := a.ensureMutable(c, docID)
	if err != nil {
		return mutableError(c, err)
	}
	if !canEditDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this editor scope")
	}
	_, v, err := a.currentVersion(c, docID)
	if err != nil {
		return apiErr(c, http.StatusNotFound, "VERSION_NOT_FOUND", "current version not found")
	}
	if d.CurrentVersion >= a.cfg.MaxVersions {
		return apiErr(c, http.StatusConflict, "VERSION_LIMIT_REACHED", "document already has 50 versions")
	}
	newVersion := d.CurrentVersion + 1
	dst := redactedVersionPath(v.FilePath, newVersion)
	if err := platform.AppendPDFMetadataMarkerFile(v.FilePath, dst, redactionTransformMarker()); err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_BURNIN_ERROR", "could not create redacted binary")
	}
	info, err := InspectPDF(dst, a.cfg.MaxUploadBytes, a.cfg.MaxPDFPages)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "REDACTION_VERIFY_ERROR", "could not verify redacted binary")
	}
	verID := makeIdentifier("ver")
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(c.Request().Context(), `INSERT INTO document_versions(id,document_id,version_number,file_path,file_sha256,size_bytes,page_count,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`, verID, docID, newVersion, dst, info.SHA256, info.Size, info.PageCount, p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "VERSION_CREATE_ERROR", "could not create redacted version")
	}
	_, _ = tx.ExecContext(c.Request().Context(), `UPDATE documents SET current_version=$1,status='Redaction Pending',updated_at=NOW() WHERE id=$2`, newVersion, docID)
	_, _ = tx.ExecContext(c.Request().Context(), `UPDATE redaction_proposals SET status='Confirmed',confirmed_by=$1,confirmed_at=NOW() WHERE id=$2 AND document_id=$3`, p.UserID, c.Param("redaction_id"), docID)
	_, _ = tx.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'REDACTION_CONFIRM',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit redaction")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "Confirmed", "version": newVersion})
}

func (a *App) createAnnotation(c echo.Context) error {
	p := principal(c)
	docID := c.Param("id")
	d, err := a.ensureMutable(c, docID)
	if err != nil {
		return mutableError(c, err)
	}
	if !canReviewDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this reviewer scope")
	}
	var req struct {
		Type        string  `json:"type"`
		Page        int     `json:"page"`
		X           float64 `json:"x"`
		Y           float64 `json:"y"`
		Width       float64 `json:"width"`
		Height      float64 `json:"height"`
		Comment     string  `json:"comment"`
		Disposition string  `json:"disposition"`
	}
	if err := c.Bind(&req); err != nil || req.Page < 1 || !IsValidAnnotationComment(req.Comment) {
		return apiErr(c, http.StatusBadRequest, "INVALID_ANNOTATION", "valid page and comment up to 2000 characters are required")
	}
	if !IsValidAnnotationType(req.Type) {
		return apiErr(c, http.StatusBadRequest, "INVALID_ANNOTATION_TYPE", "annotation type is not supported")
	}
	req.Disposition = DefaultDisposition(req.Disposition)
	if !IsValidDisposition(req.Disposition) {
		return apiErr(c, http.StatusBadRequest, "INVALID_DISPOSITION", "disposition must be Approved, Rejected, or Needs Discussion")
	}
	plainComment := req.Comment
	comment, err := encryptString(a.cfg.AESKey, req.Comment)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt annotation comment")
	}
	id := makeIdentifier("ann")
	_, err = a.db.ExecContext(c.Request().Context(), `INSERT INTO annotations(id,document_id,author_user_id,type,page,x,y,width,height,comment,disposition,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())`, id, docID, p.UserID, req.Type, req.Page, req.X, req.Y, req.Width, req.Height, comment, req.Disposition)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ANNOTATION_CREATE_ERROR", "could not create annotation")
	}
	a.audit(c, p.UserID, "ANNOTATION_CREATE", docID, nil)
	a.notifyMentionedUsers(c, plainComment, docID, p.UserID)
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "disposition": req.Disposition})
}

func (a *App) listAnnotations(c echo.Context) error {
	p := principal(c)
	var d Document
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if !canReadDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this principal scope")
	}
	rows := []map[string]interface{}{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,author_user_id,type,page,x,y,width,height,comment,disposition,created_at FROM annotations WHERE document_id=$1 ORDER BY created_at DESC`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusInternalServerError, "ANNOTATION_QUERY_ERROR", "could not list annotations")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) updateAnnotationDisposition(c echo.Context) error {
	p := principal(c)
	var req struct {
		Disposition string `json:"disposition"`
	}
	if err := c.Bind(&req); err != nil || req.Disposition == "" {
		return apiErr(c, http.StatusBadRequest, "DISPOSITION_REQUIRED", "disposition is required")
	}
	if !IsValidDisposition(req.Disposition) {
		return apiErr(c, http.StatusBadRequest, "INVALID_DISPOSITION", "disposition must be Approved, Rejected, or Needs Discussion")
	}
	var docID string
	if err := a.db.GetContext(c.Request().Context(), &docID, `SELECT document_id FROM annotations WHERE id=$1`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "ANNOTATION_NOT_FOUND", "annotation not found")
	}
	d, err := a.ensureMutable(c, docID)
	if err != nil {
		return mutableError(c, err)
	}
	if !canReviewDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this reviewer scope")
	}
	_, err = a.db.ExecContext(c.Request().Context(), `UPDATE annotations SET disposition=$1,updated_at=NOW() WHERE id=$2`, req.Disposition, c.Param("id"))
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ANNOTATION_UPDATE_ERROR", "could not update disposition")
	}
	a.audit(c, p.UserID, "ANNOTATION_DISPOSITION", docID, nil)
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "updated", "disposition": req.Disposition})
}

func (a *App) applyBates(c echo.Context) error {
	return a.applyBatesVersion(c)
}
