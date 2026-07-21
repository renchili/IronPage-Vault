package app

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestRestoreLifecycleJournalIsEncryptedAndReadable(t *testing.T) {
	a := &App{cfg: Config{BackupDir: t.TempDir(), AESKey: "restore-lifecycle-test-key"}}
	record := restoreLifecycleRecord{
		ID:          "rst_test",
		Status:      "Completed",
		Action:      "BACKUP_RESTORE_COMPLETED",
		ActorUserID: "usr_admin",
		RequestID:   "req_restore_test",
		SourceIP:    "192.0.2.10",
		RequestedMetadata: map[string]interface{}{
			"database_dump_path": "/protected/backup.dump",
			"file_snapshot_path": "/protected/files.tar",
		},
		Metadata: map[string]interface{}{
			"result": map[string]interface{}{"database_restore": "pg_restore"},
		},
		UpdatedAt: time.Now().UTC(),
	}

	if err := a.writeRestoreLifecycleRecord(record); err != nil {
		t.Fatalf("write lifecycle journal: %v", err)
	}
	path, err := a.restoreLifecyclePath(record.ID)
	if err != nil {
		t.Fatalf("resolve lifecycle path: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read raw lifecycle journal: %v", err)
	}
	for _, secret := range []string{record.ActorUserID, record.RequestID, record.SourceIP, "/protected/backup.dump", "/protected/files.tar"} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("journal exposed protected value %q", secret)
		}
	}

	opened, err := a.readRestoreLifecycleRecord(path)
	if err != nil {
		t.Fatalf("open lifecycle journal: %v", err)
	}
	if opened.ID != record.ID || opened.ActorUserID != record.ActorUserID || opened.Status != record.Status || opened.Action != record.Action {
		t.Fatalf("opened lifecycle record differs: %#v", opened)
	}
	if opened.SourceIP != record.SourceIP || opened.RequestedMetadata["database_dump_path"] != record.RequestedMetadata["database_dump_path"] {
		t.Fatalf("opened lifecycle protected metadata differs: %#v", opened)
	}
}

func TestRestoreLifecycleJournalRejectsPlaintextEnvelope(t *testing.T) {
	a := &App{cfg: Config{BackupDir: t.TempDir(), AESKey: "restore-lifecycle-test-key"}}
	path, err := a.restoreLifecyclePath("rst_plaintext")
	if err != nil {
		t.Fatalf("resolve lifecycle path: %v", err)
	}
	if err := os.MkdirAll(a.restoreLifecycleDirectory(), 0700); err != nil {
		t.Fatalf("create lifecycle directory: %v", err)
	}
	plaintextEnvelope := `{"algorithm":"AES-256-GCM","ciphertext":"{\"id\":\"rst_plaintext\",\"actor_user_id\":\"usr_admin\",\"request_id\":\"req_plaintext\"}"}`
	if err := os.WriteFile(path, []byte(plaintextEnvelope), 0600); err != nil {
		t.Fatalf("write plaintext lifecycle envelope: %v", err)
	}
	if _, err := a.readRestoreLifecycleRecord(path); err == nil {
		t.Fatal("expected plaintext lifecycle envelope to be rejected")
	}
}

func TestRestoreLifecycleJournalRequiresActingUser(t *testing.T) {
	a := &App{cfg: Config{BackupDir: t.TempDir(), AESKey: "restore-lifecycle-test-key"}}
	err := a.writeRestoreLifecycleRecord(restoreLifecycleRecord{ID: "rst_test", RequestID: "req_restore_test"})
	if err == nil {
		t.Fatal("expected missing acting user to be rejected")
	}
}

func TestRestoreLifecyclePathRejectsTraversal(t *testing.T) {
	a := &App{cfg: Config{BackupDir: t.TempDir(), AESKey: "restore-lifecycle-test-key"}}
	for _, id := range []string{"", "../rst", "nested/rst", `nested\\rst`} {
		if _, err := a.restoreLifecyclePath(id); err == nil {
			t.Fatalf("expected invalid restore id %q to be rejected", id)
		}
	}
}
