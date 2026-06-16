package app

import (
    "os"
    "strings"
    "testing"
)

func TestBackupRunRouteUsesMetadataSnapshotHandler(t *testing.T) {
    raw, err := os.ReadFile("server.go")
    if err != nil { t.Fatal(err) }
    src := string(raw)
    if !strings.Contains(src, `admin.POST("/backup/run", a.runBackupMetadataSnapshot)`) {
        t.Fatalf("backup run route should use metadata snapshot handler")
    }
}

func TestBackupMetadataSnapshotResponseDocumentsRestoreUnsupported(t *testing.T) {
    raw, err := os.ReadFile("backup_file.go")
    if err != nil { t.Fatal(err) }
    src := string(raw)
    if !strings.Contains(src, `"restore_supported":false`) {
        t.Fatalf("metadata snapshot response should declare restore_supported=false")
    }
}
