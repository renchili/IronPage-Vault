package app

import (
    "net/http"
    "strconv"
    "strings"

    "github.com/labstack/echo/v4"
)

func (a *App) ensureMutable(c echo.Context, docID string) (Document, error) {
    var d Document
    if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, docID); err != nil { return d, err }
    if d.Status == StatusFinalized { return d, echo.NewHTTPError(http.StatusConflict, "finalized") }
    return d, nil
}

func (a *App) proposeRedaction(c echo.Context) error {
    p := principal(c); docID := c.Param("id")
    if _, err := a.ensureMutable(c, docID); err != nil { return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable") }
    var req struct{ Page int `json:"page"`; X float64 `json:"x"`; Y float64 `json:"y"`; Width float64 `json:"width"`; Height float64 `json:"height"`; Reason string `json:"reason"` }
    if err := c.Bind(&req); err != nil || req.Page < 1 || req.Width <= 0 || req.Height <= 0 { return apiErr(c, http.StatusBadRequest, "INVALID_REDACTION_REGION", "page and positive coordinates are required") }
    id := makeIdentifier("red")
    _, err := a.db.ExecContext(c.Request().Context(), `INSERT INTO redaction_proposals(id,document_id,page,x,y,width,height,reason,status,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'Staged',$9,NOW())`, id, docID, req.Page, req.X, req.Y, req.Width, req.Height, req.Reason, p.UserID)
    if err != nil { return apiErr(c, http.StatusInternalServerError, "REDACTION_CREATE_ERROR", "could not stage redaction") }
    _, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'REDACTION_PROPOSE',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
    return c.JSON(http.StatusCreated, map[string]interface{}{"id":id,"status":"Staged"})
}

func (a *App) listRedactions(c echo.Context) error {
    rows := []map[string]interface{}{}
    if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,page,x,y,width,height,reason,status,created_by,created_at FROM redaction_proposals WHERE document_id=$1 ORDER BY created_at DESC`, c.Param("id")); err != nil { return apiErr(c, http.StatusInternalServerError, "REDACTION_QUERY_ERROR", "could not list redactions") }
    return c.JSON(http.StatusOK, map[string]interface{}{"data":rows})
}

func (a *App) confirmRedaction(c echo.Context) error {
    p := principal(c); docID := c.Param("id")
    d, err := a.ensureMutable(c, docID); if err != nil { return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable") }
    _, v, err := a.currentVersion(c, docID); if err != nil { return apiErr(c, http.StatusNotFound, "VERSION_NOT_FOUND", "current version not found") }
    if d.CurrentVersion >= a.cfg.MaxVersions { return apiErr(c, http.StatusConflict, "VERSION_LIMIT_REACHED", "document already has 50 versions") }
    newVersion := d.CurrentVersion + 1
    dst := strings.TrimSuffix(v.FilePath, ".pdf") + "_redacted_v" + strconv.Itoa(newVersion) + ".pdf"
    if err := ApplyAppendOnlyPDFTransform(v.FilePath, dst, "redaction burn-in confirmed"); err != nil { return apiErr(c, http.StatusInternalServerError, "REDACTION_BURNIN_ERROR", "could not create redacted binary") }
    info, _ := InspectPDF(dst, a.cfg.MaxUploadBytes, a.cfg.MaxPDFPages)
    verID := makeIdentifier("ver")
    tx, err := a.db.BeginTxx(c.Request().Context(), nil); if err != nil { return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction") }
    defer tx.Rollback()
    _, err = tx.ExecContext(c.Request().Context(), `INSERT INTO document_versions(id,document_id,version_number,file_path,file_sha256,size_bytes,page_count,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`, verID, docID, newVersion, dst, info.SHA256, info.Size, info.PageCount, p.UserID)
    if err != nil { return apiErr(c, http.StatusInternalServerError, "VERSION_CREATE_ERROR", "could not create redacted version") }
    _, _ = tx.ExecContext(c.Request().Context(), `UPDATE documents SET current_version=$1,status='Redaction Pending',updated_at=NOW() WHERE id=$2`, newVersion, docID)
    _, _ = tx.ExecContext(c.Request().Context(), `UPDATE redaction_proposals SET status='Confirmed',confirmed_by=$1,confirmed_at=NOW() WHERE id=$2 AND document_id=$3`, p.UserID, c.Param("redaction_id"), docID)
    _, _ = tx.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'REDACTION_CONFIRM',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
    if err := tx.Commit(); err != nil { return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit redaction") }
    return c.JSON(http.StatusOK, map[string]interface{}{"status":"Confirmed","version":newVersion})
}

func (a *App) createAnnotation(c echo.Context) error {
    p := principal(c); docID := c.Param("id")
    if _, err := a.ensureMutable(c, docID); err != nil { return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable") }
    var req struct{ Type string `json:"type"`; Page int `json:"page"`; X float64 `json:"x"`; Y float64 `json:"y"`; Width float64 `json:"width"`; Height float64 `json:"height"`; Comment string `json:"comment"`; Disposition string `json:"disposition"` }
    if err := c.Bind(&req); err != nil || req.Page < 1 || len(req.Comment) > 2000 { return apiErr(c, http.StatusBadRequest, "INVALID_ANNOTATION", "valid page and comment up to 2000 characters are required") }
    if req.Disposition == "" { req.Disposition = "Needs Discussion" }
    id := makeIdentifier("ann")
    _, err := a.db.ExecContext(c.Request().Context(), `INSERT INTO annotations(id,document_id,author_user_id,type,page,x,y,width,height,comment,disposition,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())`, id, docID, p.UserID, req.Type, req.Page, req.X, req.Y, req.Width, req.Height, req.Comment, req.Disposition)
    if err != nil { return apiErr(c, http.StatusInternalServerError, "ANNOTATION_CREATE_ERROR", "could not create annotation") }
    _, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'ANNOTATION_CREATE',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
    return c.JSON(http.StatusCreated, map[string]interface{}{"id":id,"disposition":req.Disposition})
}

func (a *App) listAnnotations(c echo.Context) error {
    rows := []map[string]interface{}{}
    if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,author_user_id,type,page,x,y,width,height,comment,disposition,created_at FROM annotations WHERE document_id=$1 ORDER BY created_at DESC`, c.Param("id")); err != nil { return apiErr(c, http.StatusInternalServerError, "ANNOTATION_QUERY_ERROR", "could not list annotations") }
    return c.JSON(http.StatusOK, map[string]interface{}{"data":rows})
}

func (a *App) updateAnnotationDisposition(c echo.Context) error {
    p := principal(c)
    var req struct{ Disposition string `json:"disposition"` }
    if err := c.Bind(&req); err != nil || req.Disposition == "" { return apiErr(c, http.StatusBadRequest, "DISPOSITION_REQUIRED", "disposition is required") }
    allowed := map[string]bool{"Approved":true,"Rejected":true,"Needs Discussion":true}
    if !allowed[req.Disposition] { return apiErr(c, http.StatusBadRequest, "INVALID_DISPOSITION", "disposition must be Approved, Rejected, or Needs Discussion") }
    var docID string
    if err := a.db.GetContext(c.Request().Context(), &docID, `SELECT document_id FROM annotations WHERE id=$1`, c.Param("id")); err != nil { return apiErr(c, http.StatusNotFound, "ANNOTATION_NOT_FOUND", "annotation not found") }
    if _, err := a.ensureMutable(c, docID); err != nil { return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable") }
    _, err := a.db.ExecContext(c.Request().Context(), `UPDATE annotations SET disposition=$1,updated_at=NOW() WHERE id=$2`, req.Disposition, c.Param("id"))
    if err != nil { return apiErr(c, http.StatusInternalServerError, "ANNOTATION_UPDATE_ERROR", "could not update disposition") }
    _, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'ANNOTATION_DISPOSITION',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
    return c.JSON(http.StatusOK, map[string]interface{}{"status":"updated","disposition":req.Disposition})
}

func (a *App) applyBates(c echo.Context) error {
    p := principal(c); docID := c.Param("id")
    if _, err := a.ensureMutable(c, docID); err != nil { return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable") }
    var req struct{ Prefix string `json:"prefix"`; Suffix string `json:"suffix"`; ZeroPadding int `json:"zero_padding"`; Start int `json:"start"` }
    if err := c.Bind(&req); err != nil { return apiErr(c, http.StatusBadRequest, "INVALID_BATES_REQUEST", "invalid request body") }
    if req.ZeroPadding < 0 || req.ZeroPadding > 10 { return apiErr(c, http.StatusBadRequest, "INVALID_ZERO_PADDING", "zero padding must be between 0 and 10") }
    if req.Start < 1 { req.Start = 1 }
    id := makeIdentifier("bts")
    _, err := a.db.ExecContext(c.Request().Context(), `INSERT INTO bates_jobs(id,document_id,prefix,suffix,zero_padding,start_number,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,NOW())`, id, docID, req.Prefix, req.Suffix, req.ZeroPadding, req.Start, p.UserID)
    if err != nil { return apiErr(c, http.StatusInternalServerError, "BATES_CREATE_ERROR", "could not create Bates job") }
    _, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'BATES_APPLY',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
    return c.JSON(http.StatusCreated, map[string]interface{}{"id":id,"status":"created"})
}
