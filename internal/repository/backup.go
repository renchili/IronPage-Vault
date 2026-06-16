package repository

import "context"

type BackupTableCount struct {
	Table string `json:"table"`
	Count int    `json:"count"`
}

func BackupSnapshotTables() []string {
	return []string{"users", "documents", "document_versions", "audit_logs", "notifications", "backup_jobs"}
}

func (r Repository) CountBackupTables(ctx context.Context) ([]BackupTableCount, error) {
	counts := []BackupTableCount{}
	for _, table := range BackupSnapshotTables() {
		var n int
		if err := r.DB.GetContext(ctx, &n, "SELECT COUNT(*) FROM "+table); err != nil {
			return nil, err
		}
		counts = append(counts, BackupTableCount{Table: table, Count: n})
	}
	return counts, nil
}

func (r Repository) InsertBackupJob(ctx context.Context, id string, targetPath string, userID string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'metadata_snapshot','Completed',$2,$3,NOW())`, id, targetPath, userID)
	return err
}
