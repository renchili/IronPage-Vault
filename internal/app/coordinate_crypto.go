package app

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
)

func formatFloat(v float64) string {
	return fmt.Sprintf("%.6f", v)
}

func (a *App) confirmedRedactionRegions(c echo.Context, docID string, redactionID string) ([]platform.RedactionRegion, error) {
	rows := []struct {
		Page   int     `db:"page"`
		X      float64 `db:"x"`
		Y      float64 `db:"y"`
		Width  float64 `db:"width"`
		Height float64 `db:"height"`
		Reason string  `db:"reason"`
	}{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT page,x,y,width,height,reason FROM redaction_proposals WHERE document_id=$1 AND id=$2`, docID, redactionID); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, echo.NewHTTPError(http.StatusNotFound, "redaction not found")
	}
	regions := make([]platform.RedactionRegion, 0, len(rows))
	for _, r := range rows {
		regions = append(regions, platform.RedactionRegion{Page: r.Page, X: r.X, Y: r.Y, Width: r.Width, Height: r.Height, Reason: r.Reason})
	}
	return regions, nil
}
