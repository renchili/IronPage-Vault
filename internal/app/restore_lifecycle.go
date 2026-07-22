package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	restoreLifecycleDirectoryName = ".restore-lifecycle"
	restoreStatusRequested        = "Requested"
	restoreStatusCompleted        = "Completed"
	restoreStatusFailed           = "Failed"
	restoreStatusInterrupted      = "Interrupted"
)

type restoreLifecycleRecord struct {
	ID                 string                 `json:"id"`
	Status             string                 `json:"status"`
	Action             string                 `json:"action"`
	ActorUserID        string                 `json:"actor_user_id"`
	OutcomeActorUserID string                 `json:"outcome_actor_user_id,omitempty"`
	RequestID          string                 `json:"request_id"`
	SourceIP           string                 `json:"source_ip"`
	RequestedMetadata  map[string]interface{} `json:"requested_metadata"`
	Metadata           map[string]interface{} `json:"metadata"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

func (a *App) restoreLifecycleDirectory() string {
	return filepath.Join(a.cfg.BackupDir, restoreLifecycleDirectoryName)
}

func (a *App) restoreLifecyclePath(id string) (string, error) {
	if id == "" || filepath.Base(id) != id || strings.ContainsAny(id, `/\`) {
		return "", fmt.Errorf("invalid restore lifecycle id")
	}
	return filepath.Join(a.restoreLifecycleDirectory(), id+".json"), nil
}

func (a *App) writeRestoreLifecycleRecord(record restoreLifecycleRecord) error {
	path, err := a.restoreLifecyclePath(record.ID)
	if err != nil {
		return err
	}
	if record.ActorUserID == "" || record.RequestID == "" {
		return fmt.Errorf("restore lifecycle requires actor and request id")
	}
	directory := a.restoreLifecycleDirectory()
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	if err := os.Chmod(directory, 0700); err != nil {
		return err
	}
	plaintext, err := json.Marshal(record)
	if err != nil {
		return err
	}
	ciphertext, err := encryptString(a.cfg.AESKey, string(plaintext))
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(protectedMetadata{Algorithm: "AES-256-GCM", Ciphertext: ciphertext})
	if err != nil {
		return err
	}
	temporaryPath := path + ".tmp"
	file, err := os.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err := file.Write(encoded); err != nil {
		_ = file.Close()
		_ = os.Remove(temporaryPath)
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(temporaryPath)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer directoryHandle.Close()
	return directoryHandle.Sync()
}

func (a *App) readRestoreLifecycleRecord(path string) (restoreLifecycleRecord, error) {
	encoded, err := os.ReadFile(path)
	if err != nil {
		return restoreLifecycleRecord{}, err
	}
	var envelope protectedMetadata
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return restoreLifecycleRecord{}, err
	}
	if envelope.Algorithm != "AES-256-GCM" || !strings.HasPrefix(envelope.Ciphertext, encryptedPrefix) {
		return restoreLifecycleRecord{}, fmt.Errorf("restore lifecycle journal is not protected")
	}
	plaintext, err := decryptString(a.cfg.AESKey, envelope.Ciphertext)
	if err != nil {
		return restoreLifecycleRecord{}, err
	}
	var record restoreLifecycleRecord
	if err := json.Unmarshal([]byte(plaintext), &record); err != nil {
		return restoreLifecycleRecord{}, err
	}
	if record.ActorUserID == "" || record.RequestID == "" || filepath.Base(path) != record.ID+".json" {
		return restoreLifecycleRecord{}, fmt.Errorf("restore lifecycle journal identity is invalid")
	}
	return record, nil
}

func (a *App) removeRestoreLifecycleRecord(id string) error {
	path, err := a.restoreLifecyclePath(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (a *App) restoreAuditExists(ctx context.Context, executor sqlx.ExtContext, actorID, action, requestID string) (bool, error) {
	var count int
	if err := sqlx.GetContext(ctx, executor, &count, `SELECT COUNT(*) FROM audit_logs WHERE actor_user_id=$1 AND action_type=$2 AND request_id=$3`, actorID, action, requestID); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (a *App) ensureRestoreAudit(ctx context.Context, executor sqlx.ExtContext, actorID string, record restoreLifecycleRecord, action string, metadata map[string]interface{}) error {
	if actorID == "" {
		return fmt.Errorf("restore audit acting user is required")
	}
	exists, err := a.restoreAuditExists(ctx, executor, actorID, action, record.RequestID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return a.insertAuditRecordWithExecutor(ctx, executor, actorID, action, "", record.RequestID, record.SourceIP, metadata)
}

func validRestoreStatus(status string) bool {
	switch status {
	case restoreStatusRequested, restoreStatusCompleted, restoreStatusFailed, restoreStatusInterrupted:
		return true
	default:
		return false
	}
}

func (a *App) persistRestoreLifecycleRecord(ctx context.Context, executor sqlx.ExtContext, record restoreLifecycleRecord) error {
	if record.ActorUserID == "" {
		return fmt.Errorf("restore lifecycle acting user is required")
	}
	if !validRestoreStatus(record.Status) {
		return fmt.Errorf("invalid restore lifecycle status %q", record.Status)
	}
	targetPath, _ := record.RequestedMetadata["database_dump_path"].(string)
	if _, err := executor.ExecContext(ctx, `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'restore',$2,$3,$4,NOW()) ON CONFLICT(id) DO UPDATE SET status=excluded.status,target_path=excluded.target_path,created_by=COALESCE(backup_jobs.created_by,excluded.created_by)`, record.ID, record.Status, targetPath, record.ActorUserID); err != nil {
		return err
	}
	if err := a.ensureRestoreAudit(ctx, executor, record.ActorUserID, record, "BACKUP_RESTORE_REQUESTED", record.RequestedMetadata); err != nil {
		return err
	}
	if record.Status != restoreStatusRequested {
		if record.Action == "" {
			return fmt.Errorf("terminal restore lifecycle action is required")
		}
		outcomeActor := record.OutcomeActorUserID
		if outcomeActor == "" {
			outcomeActor = record.ActorUserID
		}
		if err := a.ensureRestoreAudit(ctx, executor, outcomeActor, record, record.Action, record.Metadata); err != nil {
			return err
		}
	}
	return nil
}

func cloneRestoreMetadata(source map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(source)+3)
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func interruptedRestoreRecord(record restoreLifecycleRecord) restoreLifecycleRecord {
	record.Status = restoreStatusInterrupted
	record.Action = "BACKUP_RESTORE_INTERRUPTED"
	record.OutcomeActorUserID = systemPrincipalID
	record.Metadata = cloneRestoreMetadata(record.RequestedMetadata)
	record.Metadata["outcome"] = "unknown"
	record.Metadata["error"] = "restore process ended before a durable platform outcome was recorded"
	record.Metadata["reconciliation"] = "operator_verification_required"
	record.UpdatedAt = time.Now().UTC()
	return record
}

func (a *App) reconcileRestoreLifecycle(ctx context.Context) error {
	entries, err := os.ReadDir(a.restoreLifecycleDirectory())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(a.restoreLifecycleDirectory(), entry.Name())
		record, err := a.readRestoreLifecycleRecord(path)
		if err != nil {
			return fmt.Errorf("invalid restore lifecycle journal %s: %w", entry.Name(), err)
		}
		if record.Status == restoreStatusRequested {
			record = interruptedRestoreRecord(record)
			if err := a.writeRestoreLifecycleRecord(record); err != nil {
				return err
			}
		}
		if !validRestoreStatus(record.Status) || record.Status == restoreStatusRequested {
			return fmt.Errorf("invalid restore lifecycle status %q", record.Status)
		}
		tx, err := a.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		if err := a.persistRestoreLifecycleRecord(ctx, tx, record); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		if record.Status == restoreStatusCompleted || record.Status == restoreStatusFailed {
			if err := a.removeRestoreLifecycleRecord(record.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
