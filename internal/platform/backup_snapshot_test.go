package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewBackupMetadataSnapshot(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	snapshot := NewBackupMetadataSnapshot("bak_1", "ironpage", []BackupTableCount{{Table: "users", Count: 2}}, createdAt)

	if snapshot.BackupID != "bak_1" || snapshot.Database != "ironpage" || !snapshot.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if len(snapshot.Tables) != 1 || snapshot.Tables[0].Table != "users" || snapshot.Tables[0].Count != 2 {
		t.Fatalf("unexpected tables: %#v", snapshot.Tables)
	}
}

func TestWriteBackupMetadataSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot.json")
	snapshot := NewBackupMetadataSnapshot("bak_1", "ironpage", []BackupTableCount{{Table: "users", Count: 2}}, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))

	if err := WriteBackupMetadataSnapshot(path, snapshot); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, `"backup_id": "bak_1"`) || !strings.Contains(text, `"database": "ironpage"`) {
		t.Fatalf("unexpected snapshot json: %s", text)
	}
}
