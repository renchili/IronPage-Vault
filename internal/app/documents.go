package app

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func (a *App) persistPDFUpload(c echo.Context, fh *multipart.FileHeader, title string, p Principal) (Document, error) {
	src, err := fh.Open()
	if err != nil {
		return Document{}, err
	}
	defer src.Close()

	docID := makeIdentifier("doc")
	versionID := makeIdentifier("ver")
	dir := filepath.Join(a.cfg.StorageDir, docID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return Document{}, err
	}
	persisted := false
	defer func() {
		if !persisted {
			_ = os.RemoveAll(dir)
		}
	}()

	path := filepath.Join(dir, "v1.pdf")
	dst, err := os.Create(path)
	if err != nil {
		return Document{}, err
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return Document{}, err
	}
	if err := dst.Close(); err != nil {
		return Document{}, err
	}

	info, err := InspectPDF(path, a.cfg.MaxUploadBytes, a.cfg.MaxPDFPages)
	if err != nil {
		return Document{}, err
	}
	if title == "" {
		title = fh.Filename
	}
	titleCipher, err := sealPII(a.cfg.AESKey, title)
	if err != nil {
		return Document{}, err
	}

	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return Document{}, err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(c.Request().Context(), `INSERT INTO documents(id,title,title_ciphertext,status,owner_id,current_version,created_at,updated_at) VALUES($1,'',$2,$3,$4,1,NOW(),NOW())`, docID, titleCipher, StatusDraft, p.UserID); err != nil {
		return Document{}, err
	}
	if _, err = tx.ExecContext(c.Request().Context(), `INSERT INTO document_versions(id,document_id,version_number,file_path,file_sha256,size_bytes,page_count,created_by,created_at) VALUES($1,$2,1,$3,$4,$5,$6,$7,NOW())`, versionID, docID, path, info.SHA256, info.Size, info.PageCount, p.UserID); err != nil {
		return Document{}, err
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "DOCUMENT_UPLOAD", docID, map[string]interface{}{"version": 1, "file_sha256": info.SHA256}); err != nil {
		return Document{}, err
	}
	if err := tx.Commit(); err != nil {
		return Document{}, err
	}
	persisted = true

	var d Document
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, docID); err != nil {
		return d, err
	}
	if err := openDocumentPII(a.cfg.AESKey, &d); err != nil {
		return d, err
	}
	return d, nil
}

func (a *App) uploadDocument(c echo.Context) error {
	p := principal(c)
	fh, err := c.FormFile("file")
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "PDF_REQUIRED", "pdf is required")
	}
	if fh.Size > a.cfg.MaxUploadBytes {
		return apiErr(c, http.StatusBadRequest, "PDF_TOO_LARGE", "pdf too large")
	}
	d, err := a.persistPDFUpload(c, fh, c.FormValue("title"), p)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "PDF_UPLOAD_REJECTED", "pdf upload was not committed")
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": d})
}

func (a *App) batchUploadDocuments(c echo.Context) error {
	p := principal(c)
	form, err := c.MultipartForm()
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_MULTIPART", "invalid multipart")
	}
	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}
	if !IsValidBatchSize(len(files), a.cfg.MaxBatchFiles) {
		return apiErr(c, http.StatusBadRequest, "BATCH_LIMIT_INVALID", "invalid batch size")
	}

	created := []Document{}
	for _, fh := range files {
		if fh.Size > a.cfg.MaxUploadBytes {
			return apiErr(c, http.StatusBadRequest, "PDF_TOO_LARGE", "pdf too large")
		}
		d, err := a.persistPDFUpload(c, fh, fh.Filename, p)
		if err != nil {
			return apiErr(c, http.StatusBadRequest, "BATCH_FILE_REJECTED", "batch file was not committed")
		}
		created = append(created, d)
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": created, "count": len(created)})
}

func (a *App) listDocuments(c echo.Context) error {
	p := principal(c)
	page, size, err := a.configuredPage(c)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "PAGINATION_CONFIG_ERROR", "could not read pagination configuration")
	}
	rows := []Document{}
	where, args := documentListWhereClause(p)
	args = append(args, size, (page-1)*size)
	query := fmt.Sprintf(`SELECT * FROM documents WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))
	if err := a.db.SelectContext(c.Request().Context(), &rows, query, args...); err != nil {
		return apiErr(c, http.StatusInternalServerError, "DOCUMENT_QUERY_ERROR", "could not list documents")
	}
	for i := range rows {
		if err := openDocumentPII(a.cfg.AESKey, &rows[i]); err != nil {
			return apiErr(c, http.StatusInternalServerError, "DOCUMENT_QUERY_ERROR", "could not read documents")
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows, "page": page, "page_size": size})
}

func (a *App) getDocument(c echo.Context) error {
	p := principal(c)
	var d Document
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, c.Param("id")); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if !canReadDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this principal scope")
	}
	if err := openDocumentPII(a.cfg.AESKey, &d); err != nil {
		return apiErr(c, http.StatusInternalServerError, "DOCUMENT_QUERY_ERROR", "could not read document")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": d})
}

func (a *App) currentVersion(c echo.Context, docID string) (Document, DocumentVersion, error) {
	var d Document
	var v DocumentVersion
	if err := a.db.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1`, docID); err != nil {
		return d, v, err
	}
	if err := a.db.GetContext(c.Request().Context(), &v, `SELECT * FROM document_versions WHERE document_id=$1 AND version_number=$2`, docID, d.CurrentVersion); err != nil {
		return d, v, err
	}
	if err := openDocumentPII(a.cfg.AESKey, &d); err != nil {
		return d, v, err
	}
	return d, v, nil
}

func (a *App) downloadDocument(c echo.Context) error {
	p := principal(c)
	d, v, err := a.currentVersion(c, c.Param("id"))
	if err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_VERSION_NOT_FOUND", "current version not found")
	}
	if !canReadDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this principal scope")
	}
	return c.File(v.FilePath)
}

func (a *App) listVersions(c echo.Context) error {
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
	rows := []DocumentVersion{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT * FROM document_versions WHERE document_id=$1 ORDER BY version_number DESC LIMIT $2 OFFSET $3`, c.Param("id"), size, (page-1)*size); err != nil {
		return apiErr(c, http.StatusInternalServerError, "VERSION_QUERY_ERROR", "could not list versions")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows, "page": page, "page_size": size})
}

func (a *App) rollbackVersion(c echo.Context) error {
	p := principal(c)
	docID := c.Param("id")
	var req struct {
		Version int `json:"version"`
	}
	if err := c.Bind(&req); err != nil || req.Version < 1 {
		return apiErr(c, http.StatusBadRequest, "VERSION_REQUIRED", "version must be positive")
	}

	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()

	var d Document
	if err := tx.GetContext(c.Request().Context(), &d, `SELECT * FROM documents WHERE id=$1 FOR UPDATE`, docID); err != nil {
		return apiErr(c, http.StatusNotFound, "DOCUMENT_NOT_FOUND", "document not found")
	}
	if d.Status == StatusFinalized {
		return apiErr(c, http.StatusConflict, "DOCUMENT_FINALIZED", "finalized documents are immutable")
	}
	if !canEditDocumentObject(p, d) {
		return apiErr(c, http.StatusForbidden, "DOCUMENT_ACCESS_DENIED", "document is outside this editor scope")
	}

	var exists int
	if err := tx.GetContext(c.Request().Context(), &exists, `SELECT COUNT(*) FROM document_versions WHERE document_id=$1 AND version_number=$2`, docID, req.Version); err != nil {
		return apiErr(c, http.StatusInternalServerError, "VERSION_QUERY_ERROR", "could not verify version")
	}
	if exists == 0 {
		return apiErr(c, http.StatusBadRequest, "VERSION_NOT_FOUND", "target version does not exist")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `UPDATE documents SET current_version=$1,updated_at=NOW() WHERE id=$2`, req.Version, docID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "ROLLBACK_ERROR", "rollback failed")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "VERSION_ROLLBACK", docID, map[string]interface{}{"from_version": d.CurrentVersion, "to_version": req.Version}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record audit log")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit rollback")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "rolled_back", "version": req.Version})
}
