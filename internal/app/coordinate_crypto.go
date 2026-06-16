package app

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/platform"
	"ironpage-vault/internal/repository"
)

func formatFloat(v float64) string {
	return fmt.Sprintf("%.6f", v)
}

func (a *App) confirmedRedactionRegions(c echo.Context, docID string, redactionID string) ([]platform.RedactionRegion, error) {
	regions, err := repository.New(a.db).RedactionRegions(c.Request().Context(), docID, redactionID)
	if err != nil {
		return nil, err
	}
	if len(regions) == 0 {
		return nil, echo.NewHTTPError(http.StatusNotFound, "redaction not found")
	}
	return regions, nil
}
