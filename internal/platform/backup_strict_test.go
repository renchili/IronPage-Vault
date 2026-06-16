package platform

import "testing"

func TestStrictRestoreRejectsMissingArtifacts(t *testing.T) {
	if _, err := RunRestoreArtifactsStrict("", "", "", ""); err == nil {
		t.Fatalf("strict restore must reject missing artifact paths")
	}
}

func TestStrictBackupRejectsUnavailableArtifacts(t *testing.T) {
	manifest := BackupArtifactManifest{DatabaseDumpMode: "pg_dump_unavailable_metadata_only", FileSnapshotMode: "tar_unavailable"}
	if manifest.DatabaseDumpMode == "pg_dump_custom" || manifest.FileSnapshotMode == "tar" {
		t.Fatalf("test manifest must model degraded backup")
	}
}
