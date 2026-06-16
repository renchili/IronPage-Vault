package store

import (
	"strings"
	"testing"
)

func TestBuildAuditLogListQueryWithFilters(t *testing.T) {
	query, args := BuildAuditLogListQuery(AuditLogFilters{
		ActorUserID: "usr_1",
		ActionType:  "DOCUMENT_UPLOAD",
		CreatedFrom: "2026-01-01T00:00:00Z",
	}, 25, 50)

	for _, want := range []string{"actor_user_id=$1", "action_type=$2", "created_at >= $3::timestamptz", "LIMIT $4 OFFSET $5"} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %q: %s", want, query)
		}
	}
	if len(args) != 5 {
		t.Fatalf("args len=%d want 5", len(args))
	}
}

func TestBuildAuditLogListQuerySkipsBlankFilters(t *testing.T) {
	query, args := BuildAuditLogListQuery(AuditLogFilters{ActorUserID: "   "}, 10, 0)
	if strings.Contains(query, "actor_user_id=") {
		t.Fatalf("blank actor filter should be skipped: %s", query)
	}
	if len(args) != 2 || args[0] != 10 || args[1] != 0 {
		t.Fatalf("unexpected args: %#v", args)
	}
}
