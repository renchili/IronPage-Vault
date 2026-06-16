package service

import (
	"time"

	"ironpage-vault/internal/platform"
	"ironpage-vault/internal/repository"
)

func NewBackupSnapshot(id string, dbName string, counts []repository.BackupTableCount) platform.BackupMetadataSnapshot {
	out := make([]platform.BackupTableCount, 0, len(counts))
	for _, c := range counts {
		out = append(out, platform.BackupTableCount{Table: c.Table, Count: c.Count})
	}
	return platform.NewBackupMetadataSnapshot(id, dbName, out, time.Now().UTC())
}
