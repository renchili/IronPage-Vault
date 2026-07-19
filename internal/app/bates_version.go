package app

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
	"ironpage-vault/internal/repository"
	"ironpage-vault/internal/service"
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
	pageCount := v.PageCount
	if pageCount < 1 {
		pageCount = 1
	}
	allocatedStart, err := repository.AllocateBatesRange(c.Request().Context(), tx, req.Start, pageCount)
	if err != nil {
		if errors.Is(err, repository.ErrBatesSequenceOverlap) {
			return apiErr(c, http.StatusConflict, "BATES_SEQUENCE_OVERLAP", "requested Bates start overlaps an allocated sequence")
		}
		return apiErr(c, http.StatusInternalServerError, "BATES_SEQUENCE_ERROR", "could not reserve Bates sequence")
	}
	req.Start = NormalizeBatesStart(allocatedStart)
	newVersion := d.CurrentVersion + 1
	dst := batesVersionPath(v.FilePath, newVersion)
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(dst)
		}
	}()
	if _, err := service.ApplyBatesNumbering(v.FilePath, dst, platform.BatesOptions{Prefix: req.Prefix, Suffix: req.Suffix, ZeroPadding: req.ZeroPadding, StartNumber: req.Start}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BATES_APPLY_ERROR", "could not create Bates PDF version")
	}
	info, err := InspectPDF(dst, a.cfg.MaxUploadBytes, a.cfg.MaxPDFPages)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BATES_VERIFY_ERROR", "could not verify Bates PDF version")
	}
	jobID := makeIdentifier("bts")
	verID := makeIdentifier("ver")
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO bates_jobs(id,document_id,prefix,suffix,zero_padding,start_number,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,NOW())`, jobID, docID, req.Prefix, req.Suffix, req.ZeroPadding, req.Start, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BATES_CREATE_ERROR", "could not create Bates job")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO document_versions(id,document_id,version_number,file_path,file_sha256,size_bytes,page_count,created_by,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`, verID, docID, newVersion, dst, info.SHA256, info.Size, info.PageCount, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "VERSION_CREATE_ERROR", "could not create Bates version")
	}
	if _, err := tx.ExecContext(c.Request().Context(), `UPDATE documents SET current_version=$1,updated_at=NOW() WHERE id=$2`, newVersion, docID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "DOCUMENT_UPDATE_ERROR", "could not activate Bates version")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "BATES_APPLY", docID, map[string]interface{}{"bates_job_id": jobID, "version": newVersion, "start_number": req.Start, "page_count": pageCount, "file_sha256": info.SHA256}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record Bates audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit Bates version")
	}
	committed = true
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": jobID, "status": "created", "version": newVersion, "start_number": req.Start, "end_number": req.Start + pageCount - 1})
}
