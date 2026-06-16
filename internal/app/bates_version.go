package app

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
)

func batesVersionPath(currentPath string, newVersion int) string {
	return strings.TrimSuffix(currentPath, ".pdf") + "_bates_v" + strconv.Itoa(newVersion) + ".pdf"
}

func batesMarker(prefix, suffix string, zeroPadding int, start int) string {
	return "Bates applied: prefix=" + prefix + ", suffix=" + suffix + ", zero_padding=" + strconv.Itoa(zeroPadding) + ", start=" + strconv.Itoa(start)
}

func (a *App) applyBatesVersion(c echo.Context) error {
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
	var req struct {
		Prefix      string `json:"prefix"`
		Suffix      string `json:"suffix"`
		ZeroPadding int    `json:"zero_padding"`
		Start       int    `json:"start"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_BATES_REQUEST", "invalid request body")
	}
	if !IsValidBatesPadding(req.ZeroPadding) {
		return apiErr(c, http.StatusBadRequest, "INVALID_ZERO_PADDING", "zero padding must be between 0 and 10")
	}
	req.Start = NormalizeBatesStart(req.Start)
	newVersion := d.CurrentVersion + 1
	dst := batesVersionPath(v.FilePath, newVersion)
	marker := batesMarker(req.Prefix, req.Suffix, req.ZeroPadding, req.Start)
	if err := platform.AppendPDFMetadataMarkerFile(v.FilePath, dst, marker); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BATES_APPLY_ERROR", "could not create Bates PDF version")
	}
	info, err := InspectPDF(dst, a.cfg.MaxUploadBytes, a.cfg.MaxPDFPages)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BATES_VERIFY_ERROR", "could not verify Bates PDF version")
	}
	jobID := makeIdentifier("bts")
	verID := makeIdentifier("ver")
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(c.Request().Context(), `INSERT INTO bates_jobs(id,document_id,prefix,suffix,zero_padding,start_number,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,NOW())`, jobID, docID, req.Prefix, req.Suffix, req.ZeroPadding, req.Start, p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BATES_CREATE_ERROR", "could not create Bates job")
	}
	_, err = tx.ExecContext(c.Request().Context(), `INSERT INTO document_versions(id,document_id,version_number,file_path,file_sha256,size_bytes,page_count,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`, verID, docID, newVersion, dst, info.SHA256, info.Size, info.PageCount, p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "VERSION_CREATE_ERROR", "could not create Bates version")
	}
	_, _ = tx.ExecContext(c.Request().Context(), `UPDATE documents SET current_version=$1,updated_at=NOW() WHERE id=$2`, newVersion, docID)
	_, _ = tx.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,'BATES_APPLY',$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, docID, currentRequestID(c), c.RealIP())
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit Bates version")
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": jobID, "status": "created", "version": newVersion})
}
