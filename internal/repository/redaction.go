package repository

import (
	"context"

	"ironpage-vault/internal/platform"
)

func (r Repository) RedactionRegions(ctx context.Context, docID string, redactionID string) ([]platform.RedactionRegion, error) {
	rows := []struct {
		Page   int     `db:"page"`
		X      float64 `db:"x"`
		Y      float64 `db:"y"`
		Width  float64 `db:"width"`
		Height float64 `db:"height"`
		Reason string  `db:"reason"`
	}{}
	if err := r.DB.SelectContext(ctx, &rows, `SELECT page,x,y,width,height,reason FROM redaction_proposals WHERE document_id=$1 AND id=$2`, docID, redactionID); err != nil {
		return nil, err
	}
	regions := make([]platform.RedactionRegion, 0, len(rows))
	for _, row := range rows {
		regions = append(regions, platform.RedactionRegion{Page: row.Page, X: row.X, Y: row.Y, Width: row.Width, Height: row.Height, Reason: row.Reason})
	}
	return regions, nil
}
