package store

import (
	"fmt"
	"strings"
)

type AuditLogFilters struct {
	ActorUserID string
	DocumentID  string
	ActionType  string
	RequestID   string
	SourceIP    string
	CreatedFrom string
	CreatedTo   string
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
	add("source_ip=$%d", filters.SourceIP)

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
		`SELECT id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		strings.Join(clauses, " AND "),
		len(args)-1,
		len(args),
	)
	return query, args
}
