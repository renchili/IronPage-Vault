package app

import "ironpage-vault/internal/platform"

func (c Config) PostgresCommandConfig() platform.PostgresCommandConfig {
	return platform.PostgresCommandConfig{
		Port:          c.DBPort,
		User:          c.DBUser,
		Password:      c.DBPassword,
		Database:      c.DBName,
		CredentialDir: c.BackupDir,
	}
}
