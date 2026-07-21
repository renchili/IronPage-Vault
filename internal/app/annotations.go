package app

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type annotationResponse struct {
	ID           string     `db:"id" json:"id"`
	DocumentID   string     `db:"document_id" json:"document_id"`
	AuthorUserID string     `db:"author_user_id" json:"author_user_id"`
	Type         string     `db:"type" json:"type"`
	Page         int        `db:"page" json:"page"`
	X            float64    `db:"x" json:"x"`
	Y            float64    `db:"y" json:"y"`
	Width        float64    `db:"width" json:"width"`
	Height       float64    `db:"height" json:"height"`
	Comment      string     `db:"comment" json:"comment"`
	Disposition  string     `db:"disposition" json:"disposition"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

func (a *App) createAnnotation(c echo.Context) error {
	p := principal(c)
	docID := c.Param("id")
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
	commentCipher, err := encryptString(a.cfg.AESKey, req.Comment)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt annotation comment")
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
	if !canReviewDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this reviewer scope")
	}
	id := makeIdentifier("ann")
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO annotations(id,document_id,author_user_id,type,page,x,y,width,height,comment,disposition,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())`, id, docID, p.UserID, req.Type, req.Page, req.X, req.Y, req.Width, req.Height, commentCipher, req.Disposition); err != nil {
		return apiErr(c, http.StatusInternalServerError, "ANNOTATION_CREATE_ERROR", "could not create annotation")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "ANNOTATION_CREATE", docID, map[string]interface{}{"annotation_id": id, "type": req.Type, "page": req.Page}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record annotation audit")
	}
	if err := a.notifyMentionedUsersWithExecutor(c.Request().Context(), tx, plainComment, docID, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_CREATE_ERROR", "could not create annotation mention notifications")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit annotation")
	}
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
	page, size, err := a.configuredPage(c)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "PAGINATION_CONFIG_ERROR", "could not read pagination configuration")
	}
	rows := []annotationResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,author_user_id,type,page,x,y,width,height,comment,disposition,created_at,updated_at FROM annotations WHERE document_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, c.Param("id"), size, (page-1)*size); err != nil {
		return apiErr(c, http.StatusInternalServerError, "ANNOTATION_QUERY_ERROR", "could not list annotations")
	}
	for index := range rows {
		plain, err := decryptString(a.cfg.AESKey, rows[index].Comment)
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "ANNOTATION_QUERY_ERROR", "could not read annotation comment")
		}
		rows[index].Comment = plain
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows, "page": page, "page_size": size})
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
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	var docID string
	if err := tx.GetContext(c.Request().Context(), &docID, `SELECT document_id FROM annotations WHERE id=$1 FOR UPDATE`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "ANNOTATION_NOT_FOUND", "annotation not found")
	}
	d, err := a.ensureMutableWithExecutor(c.Request().Context(), tx, docID, true)
	if err != nil {
		return mutableError(c, err)
	}
	if !canReviewDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this reviewer scope")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `UPDATE annotations SET disposition=$1,updated_at=NOW() WHERE id=$2`, req.Disposition, c.Param("id")); err != nil {
		return apiErr(c, http.StatusInternalServerError, "ANNOTATION_UPDATE_ERROR", "could not update disposition")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "ANNOTATION_DISPOSITION", docID, map[string]interface{}{"annotation_id": c.Param("id"), "disposition": req.Disposition}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record annotation disposition audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit annotation disposition")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "updated", "disposition": req.Disposition})
}
