package app

import (
	"os"
	"strings"
	"testing"
)

func TestBackupRunRouteUsesMetadataSnapshotHandler(t *testing.T) {
	raw, err := os.ReadFile("server.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if !strings.Contains(src, `admin.POST("/backup/run", a.runBackupMetadataSnapshot)`) {
		t.Fatalf("backup run route should use metadata snapshot handler")
	}
}

func TestBackupMetadataSnapshotResponseDocumentsStrictFullBackup(t *testing.T) {
	raw, err := os.ReadFile("backup_file.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, want := range []string{
		"RunBackupArtifactsStrict",
		`"restore_supported": artifacts.RestoreSupported`,
		`"kind": "full_backup"`,
		`"artifacts": artifacts`,
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("strict backup response missing %q", want)
		}
	}
}
