package platform

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type BackupArtifactManifest struct {
	BackupID         string    `json:"backup_id"`
	CreatedAt        time.Time `json:"created_at"`
	DatabaseDumpPath string    `json:"database_dump_path"`
	FileSnapshotPath string    `json:"file_snapshot_path"`
	DatabaseDumpMode string    `json:"database_dump_mode"`
	FileSnapshotMode string    `json:"file_snapshot_mode"`
	RestoreSupported bool      `json:"restore_supported"`
}

func RunBackupArtifacts(id string, dsn string, storageDir string, backupDir string) (BackupArtifactManifest, error) {
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return BackupArtifactManifest{}, err
	}
	manifest := BackupArtifactManifest{
		BackupID:         id,
		CreatedAt:        time.Now().UTC(),
		DatabaseDumpPath: filepath.Join(backupDir, id+".dump"),
		FileSnapshotPath: filepath.Join(backupDir, id+"_files.tar"),
		RestoreSupported: true,
	}
	if _, err := exec.LookPath("pg_dump"); err == nil && dsn != "" {
		cmd := exec.Command("pg_dump", "--format=custom", "--file", manifest.DatabaseDumpPath, dsn)
		if err := cmd.Run(); err != nil {
			manifest.DatabaseDumpMode = "pg_dump_failed_metadata_only"
			_ = os.WriteFile(manifest.DatabaseDumpPath+".error", []byte(err.Error()), 0640)
		} else {
			manifest.DatabaseDumpMode = "pg_dump_custom"
		}
	} else {
		manifest.DatabaseDumpMode = "pg_dump_unavailable_metadata_only"
		_ = os.WriteFile(manifest.DatabaseDumpPath+".missing", []byte("pg_dump or DSN unavailable"), 0640)
	}
	if _, err := exec.LookPath("tar"); err == nil {
		cmd := exec.Command("tar", "-cf", manifest.FileSnapshotPath, "-C", storageDir, ".")
		if err := cmd.Run(); err != nil {
			manifest.FileSnapshotMode = "tar_failed"
			_ = os.WriteFile(manifest.FileSnapshotPath+".error", []byte(err.Error()), 0640)
		} else {
			manifest.FileSnapshotMode = "tar"
		}
	} else {
		manifest.FileSnapshotMode = "tar_unavailable"
		_ = os.WriteFile(manifest.FileSnapshotPath+".missing", []byte("tar unavailable"), 0640)
	}
	raw, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(backupDir, id+"_manifest.json"), raw, 0640); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func RunRestoreArtifacts(dsn string, databaseDumpPath string, fileSnapshotPath string, storageDir string) (map[string]string, error) {
	result := map[string]string{}
	if databaseDumpPath != "" {
		if _, err := exec.LookPath("pg_restore"); err == nil && dsn != "" {
			cmd := exec.Command("pg_restore", "--clean", "--if-exists", "--dbname", dsn, databaseDumpPath)
			if err := cmd.Run(); err != nil {
				result["database_restore"] = "pg_restore_failed"
				result["database_restore_error"] = err.Error()
			} else {
				result["database_restore"] = "pg_restore"
			}
		} else {
			result["database_restore"] = "pg_restore_unavailable"
		}
	}
	if fileSnapshotPath != "" {
		if _, err := exec.LookPath("tar"); err == nil {
			_ = os.MkdirAll(storageDir, 0750)
			cmd := exec.Command("tar", "-xf", fileSnapshotPath, "-C", storageDir)
			if err := cmd.Run(); err != nil {
				result["file_restore"] = "tar_failed"
				result["file_restore_error"] = err.Error()
			} else {
				result["file_restore"] = "tar"
			}
		} else {
			result["file_restore"] = "tar_unavailable"
		}
	}
	return result, nil
}
