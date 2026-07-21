package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
	"ironpage-vault/internal/repository"
)

func formatFloat(v float64) string {
	return fmt.Sprintf("%.6f", v)
}

func (a *App) parseRedactionFloat(ciphertext string) (float64, error) {
	plain, err := decryptString(a.cfg.AESKey, ciphertext)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(plain, 64)
}

func (a *App) confirmedRedactionRegionsWithExecutor(ctx context.Context, executor sqlx.ExtContext, docID string, redactionID string) ([]platform.RedactionRegion, error) {
	rows := []repository.EncryptedRedactionRegion{}
	if err := sqlx.SelectContext(ctx, executor, &rows, `SELECT page,x_ciphertext,y_ciphertext,width_ciphertext,height_ciphertext,reason FROM redaction_proposals WHERE document_id=$1 AND id=$2 AND status='Staged' FOR UPDATE`, docID, redactionID); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, echo.NewHTTPError(http.StatusNotFound, "staged redaction not found")
	}
	regions := make([]platform.RedactionRegion, 0, len(rows))
	for _, row := range rows {
		x, err := a.parseRedactionFloat(row.XCiphertext)
		if err != nil {
			return nil, err
		}
		y, err := a.parseRedactionFloat(row.YCiphertext)
		if err != nil {
			return nil, err
		}
		width, err := a.parseRedactionFloat(row.WidthCiphertext)
		if err != nil {
			return nil, err
		}
		height, err := a.parseRedactionFloat(row.HeightCiphertext)
		if err != nil {
			return nil, err
		}
		reason, err := decryptString(a.cfg.AESKey, row.ReasonCiphertext)
		if err != nil {
			return nil, err
		}
		regions = append(regions, platform.RedactionRegion{Page: row.Page, X: x, Y: y, Width: width, Height: height, Reason: reason})
	}
	return regions, nil
}

func (a *App) confirmedRedactionRegions(c echo.Context, docID string, redactionID string) ([]platform.RedactionRegion, error) {
	return a.confirmedRedactionRegionsWithExecutor(c.Request().Context(), a.db, docID, redactionID)
}
