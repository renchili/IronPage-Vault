package app

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
)

type auditLookupBackfillRow struct {
	ID                 string `db:"id"`
	SourceIP           string `db:"source_ip"`
	SourceIPCiphertext string `db:"source_ip_ciphertext"`
}

// EnsureAuditSourceIPLookups upgrades existing audit rows so protected source-IP
// filters remain usable after introducing deterministic lookup keys.
func EnsureAuditSourceIPLookups(ctx context.Context, db *sqlx.DB, aesKey string) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows := []auditLookupBackfillRow{}
	if err := tx.SelectContext(ctx, &rows, `SELECT id,source_ip,source_ip_ciphertext FROM audit_logs WHERE source_ip_lookup='' FOR UPDATE`); err != nil {
		return err
	}
	for _, row := range rows {
		sourceIP := row.SourceIP
		if strings.TrimSpace(row.SourceIPCiphertext) != "" {
			plain, err := decryptString(aesKey, row.SourceIPCiphertext)
			if err != nil {
				return err
			}
			sourceIP = plain
		}
		if strings.TrimSpace(sourceIP) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE audit_logs SET source_ip_lookup=$1 WHERE id=$2`, piiLookupKey(aesKey, sourceIP), row.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
