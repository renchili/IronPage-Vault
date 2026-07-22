package platform

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func RunBackupArtifacts(id string, postgres PostgresCommandConfig, storageDir string, backupDir string) (BackupArtifactManifest, error) {
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
	if _, err := exec.LookPath("pg_dump"); err == nil {
		if err := runPostgresCommand("pg_dump", pgDumpCommandArgs(postgres, manifest.DatabaseDumpPath), postgres); err != nil {
			manifest.DatabaseDumpMode = "pg_dump_failed_metadata_only"
			_ = os.WriteFile(manifest.DatabaseDumpPath+".error", []byte(err.Error()), 0640)
		} else {
			manifest.DatabaseDumpMode = "pg_dump_custom"
		}
	} else {
		manifest.DatabaseDumpMode = "pg_dump_unavailable_metadata_only"
		_ = os.WriteFile(manifest.DatabaseDumpPath+".missing", []byte("pg_dump unavailable"), 0640)
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
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return manifest, err
	}
	if err := os.WriteFile(filepath.Join(backupDir, id+"_manifest.json"), raw, 0640); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func extractTarSnapshot(snapshotPath string, targetDir string) error {
	handle, err := os.Open(snapshotPath)
	if err != nil {
		return err
	}
	defer handle.Close()
	reader := tar.NewReader(handle)
	root := filepath.Clean(targetDir) + string(os.PathSeparator)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.Clean(header.Name)
		if name == "." {
			continue
		}
		if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("unsafe path in filesystem snapshot: %s", header.Name)
		}
		destination := filepath.Join(targetDir, name)
		if !strings.HasPrefix(filepath.Clean(destination)+string(os.PathSeparator), root) {
			return fmt.Errorf("filesystem snapshot escapes restore root: %s", header.Name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destination, 0750); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(destination), 0750); err != nil {
				return err
			}
			output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(output, reader)
			closeErr := output.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			return fmt.Errorf("unsupported filesystem snapshot entry type for %s", header.Name)
		}
	}
}

func installStagedStorage(storageDir string, stagedDir string) (func() error, func() error, error) {
	parent := filepath.Dir(storageDir)
	if err := os.MkdirAll(parent, 0750); err != nil {
		return nil, nil, err
	}
	previous, err := os.MkdirTemp(parent, ".ironpage-restore-previous-")
	if err != nil {
		return nil, nil, err
	}
	if err := os.Remove(previous); err != nil {
		return nil, nil, err
	}
	hadPrevious := false
	if _, err := os.Stat(storageDir); err == nil {
		if err := os.Rename(storageDir, previous); err != nil {
			return nil, nil, err
		}
		hadPrevious = true
	} else if !os.IsNotExist(err) {
		return nil, nil, err
	}
	if err := os.Rename(stagedDir, storageDir); err != nil {
		if hadPrevious {
			_ = os.Rename(previous, storageDir)
		}
		return nil, nil, err
	}
	rollback := func() error {
		if err := os.RemoveAll(storageDir); err != nil {
			return err
		}
		if hadPrevious {
			return os.Rename(previous, storageDir)
		}
		return nil
	}
	commit := func() error {
		if hadPrevious {
			return os.RemoveAll(previous)
		}
		return nil
	}
	return rollback, commit, nil
}

func RunRestoreArtifacts(postgres PostgresCommandConfig, databaseDumpPath string, fileSnapshotPath string, storageDir string) (map[string]string, error) {
	result := map[string]string{}
	if err := postgres.validate(); err != nil {
		result["database_restore"] = "postgres_configuration_invalid"
		return result, err
	}
	if _, err := exec.LookPath("pg_restore"); err != nil {
		result["database_restore"] = "pg_restore_unavailable"
		return result, err
	}
	parent := filepath.Dir(storageDir)
	if err := os.MkdirAll(parent, 0750); err != nil {
		return result, err
	}
	stagedDir, err := os.MkdirTemp(parent, ".ironpage-restore-stage-")
	if err != nil {
		return result, err
	}
	stagedInstalled := false
	defer func() {
		if !stagedInstalled {
			_ = os.RemoveAll(stagedDir)
		}
	}()
	if err := extractTarSnapshot(fileSnapshotPath, stagedDir); err != nil {
		result["file_restore"] = "tar_failed"
		result["file_restore_error"] = err.Error()
		return result, err
	}
	result["file_restore"] = "staged"
	rollbackStorage, commitStorage, err := installStagedStorage(storageDir, stagedDir)
	if err != nil {
		result["file_restore"] = "install_failed"
		result["file_restore_error"] = err.Error()
		return result, err
	}
	stagedInstalled = true
	if err := runPostgresCommand("pg_restore", pgRestoreCommandArgs(postgres, databaseDumpPath), postgres); err != nil {
		result["database_restore"] = "pg_restore_failed"
		result["database_restore_error"] = err.Error()
		if rollbackErr := rollbackStorage(); rollbackErr != nil {
			result["file_rollback_error"] = rollbackErr.Error()
			return result, fmt.Errorf("database restore failed and filesystem rollback failed: %v; %w", rollbackErr, err)
		}
		result["file_restore"] = "rolled_back"
		return result, err
	}
	result["database_restore"] = "pg_restore"
	if err := commitStorage(); err != nil {
		result["file_restore"] = "cleanup_failed"
		result["file_cleanup_error"] = err.Error()
		return result, err
	}
	result["file_restore"] = "tar"
	return result, nil
}
