package app

import (
	"fmt"
	"net/http"
	"strconv"

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

func (a *App) confirmedRedactionRegions(c echo.Context, docID string, redactionID string) ([]platform.RedactionRegion, error) {
	rows, err := repository.New(a.db).EncryptedRedactionRegions(c.Request().Context(), docID, redactionID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, echo.NewHTTPError(http.StatusNotFound, "redaction not found")
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
		reason, _ := decryptString(a.cfg.AESKey, row.ReasonCiphertext)
		regions = append(regions, platform.RedactionRegion{Page: row.Page, X: x, Y: y, Width: width, Height: height, Reason: reason})
	}
	return regions, nil
}
