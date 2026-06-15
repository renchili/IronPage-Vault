package app

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"

    "github.com/labstack/echo/v4"
)

func (a *App) uploadDocument(c echo.Context) error {
    p := principal(c)
    fh, err := c.FormFile("file")
    if err != nil { return apiErr(c, http.StatusBadRequest, "PDF_REQUIRED", "multipart file field file is required") }
    if fh.Size > a.cfg.MaxUploadBytes { return apiErr(c, http.StatusBadRequest, "PDF_TOO_LARGE", "document exceeds 200 MB") }
    src, err := fh.Open(); if err != nil { return apiErr(c, http.StatusBadRequest, "PDF_OPEN_ERROR", "could not open upload") }
    defer src.Close()
    docID := makeIdentifier("doc")
    versionID := makeIdentifier("ver")
    dir := filepath.Join(a.cfg.StorageDir, docID)
    if err := os.MkdirAll(dir, 0750); err != nil { return apiErr(c, http.StatusInternalServerError, "STORAGE_ERROR", "could not create storage directory") }
    path := filepath.Join(dir, "v1.pdf")
    dst, err := os.Create(path); if err != nil { return apiErr(c, http.StatusInternalServerError, "STORAGE_ERROR", "could not store file") }
    if _, err := io.Copy(dst, src); err != nil { dst.Close(); return apiErr(c, http.StatusInternalServerError, "STORAGE_ERROR", "could not write file") }
    dst.Close()
    info, err := InspectPDF(path, a.cfg.MaxUploadBytes, a.cfg.MaxPDFPages)
    if err != nil { os.RemoveAll(dir); return apiErr(c, http.StatusBadRequest, "PDF_VALIDATION_FAILED", err.Error()) }
    title := c.FormValue("title"); if title == "" { title = fh.Filename }
    tx, err := a.db.BeginTxx(c.Request().Context(), nil); if err != nil { return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction") }
    defer tx.Rollback()
    _, err = tx.ExecContext(c.Request().Context(), `INSERT INTO documents(id,title,status,owner_id,current_version,created_at,updated_at) VALUES($1,$2,$3,$4,1,NOW(),NOW())`, docID, title, StatusDraft, p.UserID)
    if err != nil { return apiErr(c, http.StatusInternalServerError, "DOCUMENT_CREATE_ERROR", "could not create document") }
    _, err = tx.ExecContext(c.Request().Context(), `INSERT INTO document_versions(id,document_id,version_number,file_path,file_sha256,size_bytes,page_count,created_by,created_at) VALUES($1,$2,1,$3,$4,$5,$6,$7,NOW())`, versionID, docID, path, info.SHA256, info.Size, info.PageCount, p.UserID)
    if err != nil { return apiErr(c, http.StatusInternalServerError, "VERSION_CREATE_ERROR", "could not create version") }
    _, _ = tx.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'DOCUMENT_UPLOAD',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
    if err := tx.Commit(); err != nil { return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit upload") }
    var d Document
    _ = a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, docID)
    return c.JSON(http.StatusCreated, map[string]interface{}{"data":d})
}

func (a *App) batchUploadDocuments(c echo.Context) error {
    form, err := c.MultipartForm(); if err != nil { return apiErr(c, http.StatusBadRequest, "INVALID_MULTIPART", "multipart form is required") }
    files := form.File["files"]
    if len(files) == 0 { files = form.File["file"] }
    if len(files) == 0 { return apiErr(c, http.StatusBadRequest, "FILES_REQUIRED", "files are required") }
    if len(files) > a.cfg.MaxBatchFiles { return apiErr(c, http.StatusBadRequest, "BATCH_LIMIT_EXCEEDED", "batch import supports up to 250 files") }
    return c.JSON(http.StatusCreated, map[string]interface{}{"accepted_files":len(files),"note":"submit files individually or extend this endpoint for streaming batch persistence"})
}

func (a *App) listDocuments(c echo.Context) error {
    page, size := parsePage(c, a.cfg)
    rows := []Document{}
    if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT * FROM documents ORDER BY created_at DESC LIMIT $1 OFFSET $2`, size, (page-1)*size); err != nil { return apiErr(c, http.StatusInternalServerError, "DOCUMENT_QUERY_ERROR", "could not list documents") }
    return c.JSON(http.StatusOK, map[string]interface{}{"data":rows,"page":page,"page_size":size})
}

func (a *App) getDocument(c echo.Context) error {
    var d Document
    if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, c.Param("id")); err != nil { return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found") }
    return c.JSON(http.StatusOK, map[string]interface{}{"data":d})
}

func (a *App) currentVersion(c echo.Context, docID string) (Document, DocumentVersion, error) {
    var d Document; var v DocumentVersion
    if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, docID); err != nil { return d, v, err }
    if err := a.db.GetContext(c.Request().Context(), &v, `SELECT * FROM document_versions WHERE document_id=$1 AND version_number=$2`, docID, d.CurrentVersion); err != nil { return d, v, err }
    return d, v, nil
}

func (a *App) downloadDocument(c echo.Context) error {
    _, v, err := a.currentVersion(c, c.Param("id")); if err != nil { return apiErr(c, http.StatusNotFound, "DOCUMENT_VERSION_NOT_FOUND", "current version not found") }
    return c.File(v.FilePath)
}

func (a *App) listVersions(c echo.Context) error {
    rows := []DocumentVersion{}
    if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT * FROM document_versions WHERE document_id=$1 ORDER BY version_number DESC`, c.Param("id")); err != nil { return apiErr(c, http.StatusInternalServerError, "VERSION_QUERY_ERROR", "could not list versions") }
    return c.JSON(http.StatusOK, map[string]interface{}{"data":rows})
}

func (a *App) rollbackVersion(c echo.Context) error {
    return apiErr(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", fmt.Sprintf("rollback endpoint is reserved for document %s", c.Param("id")))
}
