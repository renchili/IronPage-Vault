package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AuditLogFilters struct {
	ActorUserID    string
	DocumentID     string
	ActionType     string
	RequestID      string
	SourceIPLookup string
	CreatedFrom    string
	CreatedTo      string
}

type AuditLogRow struct {
	ID                 string          `db:"id" json:"id"`
	ActorUserID        *string         `db:"actor_user_id" json:"actor_user_id,omitempty"`
	DocumentID         *string         `db:"document_id" json:"document_id,omitempty"`
	ActionType         string          `db:"action_type" json:"action_type"`
	RequestID          string          `db:"request_id" json:"request_id"`
	SourceIP           string          `db:"source_ip" json:"source_ip"`
	SourceIPLookup     string          `db:"source_ip_lookup" json:"-"`
	SourceIPCiphertext string          `db:"source_ip_ciphertext" json:"-"`
	Metadata           json.RawMessage `db:"metadata" json:"metadata"`
	MetadataCiphertext string          `db:"metadata_ciphertext" json:"-"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
}

func BuildAuditLogListQuery(filters AuditLogFilters, limit int, offset int) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := []interface{}{}

	add := func(expr string, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(expr, len(args)))
	}

	add("actor_user_id=$%d", filters.ActorUserID)
	add("document_id=$%d", filters.DocumentID)
	add("action_type=$%d", filters.ActionType)
	add("request_id=$%d", filters.RequestID)
	add("source_ip_lookup=$%d", filters.SourceIPLookup)

	if v := strings.TrimSpace(filters.CreatedFrom); v != "" {
		args = append(args, v)
		clauses = append(clauses, fmt.Sprintf("created_at >= $%d::timestamptz", len(args)))
	}
	if v := strings.TrimSpace(filters.CreatedTo); v != "" {
		args = append(args, v)
		clauses = append(clauses, fmt.Sprintf("created_at <= $%d::timestamptz", len(args)))
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(
		`SELECT id,actor_user_id,document_id,action_type,request_id,source_ip,source_ip_lookup,source_ip_ciphertext,metadata,metadata_ciphertext,created_at FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		strings.Join(clauses, " AND "),
		len(args)-1,
		len(args),
	)
	return query, args
}

func (r Repository) ListAuditLogs(ctx context.Context, filters AuditLogFilters, limit int, offset int) ([]AuditLogRow, error) {
	query, args := BuildAuditLogListQuery(filters, limit, offset)
	rows := []AuditLogRow{}
	if err := r.DB.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}
