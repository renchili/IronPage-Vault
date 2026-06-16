package platform

import (
	"fmt"
	"os"
)

func RunBackupArtifactsStrict(id string, dsn string, storageDir string, backupDir string) (BackupArtifactManifest, error) {
	manifest, err := RunBackupArtifacts(id, dsn, storageDir, backupDir)
	if err != nil {
		return manifest, err
	}
	if manifest.DatabaseDumpMode != "pg_dump_custom" {
		return manifest, fmt.Errorf("strict backup requires successful pg_dump_custom, got %s", manifest.DatabaseDumpMode)
	}
	if manifest.FileSnapshotMode != "tar" {
		return manifest, fmt.Errorf("strict backup requires successful tar snapshot, got %s", manifest.FileSnapshotMode)
	}
	if _, err := os.Stat(manifest.DatabaseDumpPath); err != nil {
		return manifest, fmt.Errorf("database dump artifact missing: %w", err)
	}
	if _, err := os.Stat(manifest.FileSnapshotPath); err != nil {
		return manifest, fmt.Errorf("file snapshot artifact missing: %w", err)
	}
	manifest.RestoreSupported = true
	return manifest, nil
}

func RunRestoreArtifactsStrict(dsn string, databaseDumpPath string, fileSnapshotPath string, storageDir string) (map[string]string, error) {
	if databaseDumpPath == "" || fileSnapshotPath == "" {
		return nil, fmt.Errorf("database_dump_path and file_snapshot_path are required")
	}
	if _, err := os.Stat(databaseDumpPath); err != nil {
		return nil, fmt.Errorf("database dump artifact not readable: %w", err)
	}
	if _, err := os.Stat(fileSnapshotPath); err != nil {
		return nil, fmt.Errorf("file snapshot artifact not readable: %w", err)
	}
	result, err := RunRestoreArtifacts(dsn, databaseDumpPath, fileSnapshotPath, storageDir)
	if err != nil {
		return result, err
	}
	if result["database_restore"] != "pg_restore" {
		return result, fmt.Errorf("strict restore requires pg_restore, got %s", result["database_restore"])
	}
	if result["file_restore"] != "tar" {
		return result, fmt.Errorf("strict restore requires tar extraction, got %s", result["file_restore"])
	}
	return result, nil
}
