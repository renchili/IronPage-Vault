package repository

import "context"

type EncryptedRedactionRegion struct {
	Page             int    `db:"page"`
	XCiphertext      string `db:"x_ciphertext"`
	YCiphertext      string `db:"y_ciphertext"`
	WidthCiphertext  string `db:"width_ciphertext"`
	HeightCiphertext string `db:"height_ciphertext"`
	ReasonCiphertext string `db:"reason"`
}

func (r Repository) EncryptedRedactionRegions(ctx context.Context, docID string, redactionID string) ([]EncryptedRedactionRegion, error) {
	rows := []EncryptedRedactionRegion{}
	if err := r.DB.SelectContext(ctx, &rows, `SELECT page,x_ciphertext,y_ciphertext,width_ciphertext,height_ciphertext,reason FROM redaction_proposals WHERE document_id=$1 AND id=$2`, docID, redactionID); err != nil {
		return nil, err
	}
	return rows, nil
}
