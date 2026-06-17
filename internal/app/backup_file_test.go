package app

import (
	"testing"

	"ironpage-vault/internal/repository"
)

func TestBackupSnapshotTables(t *testing.T) {
	got := repository.BackupSnapshotTables()
	required := []string{"users", "documents", "document_versions", "audit_logs", "notifications", "backup_jobs"}
	if len(got) != len(required) {
		t.Fatalf("table count=%d want %d", len(got), len(required))
	}
	seen := map[string]bool{}
	for _, table := range got {
		seen[table] = true
	}
	for _, table := range required {
		if !seen[table] {
			t.Fatalf("missing backup snapshot table %q", table)
		}
	}
}
