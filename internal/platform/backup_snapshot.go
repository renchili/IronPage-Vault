package platform

import (
	"encoding/json"
	"os"
	"time"
)

type BackupTableCount struct {
	Table string `json:"table" db:"table_name"`
	Count int    `json:"count" db:"row_count"`
}

type BackupMetadataSnapshot struct {
	BackupID  string             `json:"backup_id"`
	CreatedAt time.Time          `json:"created_at"`
	Database  string             `json:"database"`
	Tables    []BackupTableCount `json:"tables"`
}

func BackupSnapshotTables() []string {
	return []string{"users", "documents", "document_versions", "audit_logs", "notifications", "backup_jobs"}
}

func NewBackupMetadataSnapshot(id string, database string, tables []BackupTableCount, createdAt time.Time) BackupMetadataSnapshot {
	return BackupMetadataSnapshot{BackupID: id, CreatedAt: createdAt, Database: database, Tables: tables}
}

func WriteBackupMetadataSnapshot(path string, snapshot BackupMetadataSnapshot) error {
	raw, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0640)
}
