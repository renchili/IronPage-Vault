package app

import (
	"os"
	"path/filepath"

	"ironpage-vault/internal/platform"
)

func cleanupBackupArtifacts(backupDir string, id string, metadataPath string, manifest platform.BackupArtifactManifest) {
	paths := []string{
		metadataPath,
		manifest.DatabaseDumpPath,
		manifest.FileSnapshotPath,
		filepath.Join(backupDir, id+"_manifest.json"),
		manifest.DatabaseDumpPath + ".error",
		manifest.DatabaseDumpPath + ".missing",
		manifest.FileSnapshotPath + ".error",
		manifest.FileSnapshotPath + ".missing",
	}
	for _, path := range paths {
		if path != "" {
			_ = os.Remove(path)
		}
	}
}
