package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

var errVersionLimitReached = errors.New("document version limit reached")

func nextDocumentVersion(currentVersion, maxVersions int) (int, error) {
	if currentVersion < 1 || maxVersions < 1 || currentVersion >= maxVersions {
		return 0, errVersionLimitReached
	}
	return currentVersion + 1, nil
}

func validRollbackVersion(version, maxVersions int) bool {
	return version >= 1 && version <= maxVersions
}

func (a *App) ensureMutableWithExecutor(ctx context.Context, executor sqlx.ExtContext, docID string, lock bool) (Document, error) {
	var d Document
	query := `SELECT * FROM documents WHERE id=$1`
	if lock {
		query += ` FOR UPDATE`
	}
	if err := sqlx.GetContext(ctx, executor, &d, query, docID); err != nil {
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

func (a *App) ensureMutable(c echo.Context, docID string) (Document, error) {
	return a.ensureMutableWithExecutor(c.Request().Context(), a.db, docID, false)
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

func (a *App) applyBates(c echo.Context) error {
	return a.applyBatesVersion(c)
}
