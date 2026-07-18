package app

import (
	"strings"
	"testing"
)

func validRuntimeConfig() Config {
	return Config{
		HTTPAddr:      "0.0.0.0:28080",
		DBPort:        "25432",
		DBUser:        "ironpage_test",
		DBPassword:    strings.Repeat("d", 16),
		DBName:        "ironpage_test",
		JWTSecret:     strings.Repeat("j", 32),
		AESKey:        strings.Repeat("a", 32),
		StorageDir:    "/runtime/storage",
		BackupDir:     "/runtime/backups",
		MigrationsDir: "/runtime/app/migrations",
		PublicDir:     "/runtime/app/public",
	}
}

func TestConfigValidateRequiresRuntimeValues(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "http address", edit: func(cfg *Config) { cfg.HTTPAddr = "" }, want: "HTTP_ADDR is required"},
		{name: "database port", edit: func(cfg *Config) { cfg.DBPort = "" }, want: "DB_PORT is required"},
		{name: "database user", edit: func(cfg *Config) { cfg.DBUser = "" }, want: "DB_USER is required"},
		{name: "database name", edit: func(cfg *Config) { cfg.DBName = "" }, want: "DB_NAME is required"},
		{name: "storage directory", edit: func(cfg *Config) { cfg.StorageDir = "" }, want: "STORAGE_DIR is required"},
		{name: "backup directory", edit: func(cfg *Config) { cfg.BackupDir = "" }, want: "BACKUP_DIR is required"},
		{name: "migrations directory", edit: func(cfg *Config) { cfg.MigrationsDir = "" }, want: "MIGRATIONS_DIR is required"},
		{name: "public directory", edit: func(cfg *Config) { cfg.PublicDir = "" }, want: "PUBLIC_DIR is required"},
		{name: "database password", edit: func(cfg *Config) { cfg.DBPassword = "" }, want: "DB_PASSWORD is required"},
		{name: "jwt secret", edit: func(cfg *Config) { cfg.JWTSecret = "" }, want: "JWT_SECRET is required"},
		{name: "aes key", edit: func(cfg *Config) { cfg.AESKey = "" }, want: "AES_KEY is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validRuntimeConfig()
			tt.edit(&cfg)
			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestConfigValidateRejectsNonNumericDatabasePort(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.DBPort = "postgres"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "DB_PORT must be numeric") {
		t.Fatalf("Validate() error = %v, want numeric-port error", err)
	}
}

func TestConfigValidateProductionRejectsSeedPasswords(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.SeedAdminPassword = strings.Repeat("p", 12)

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "requires ACCEPTANCE_MODE=true") {
		t.Fatalf("Validate() error = %v, want acceptance-mode error", err)
	}
}

func TestConfigValidateBootstrapPair(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.BootstrapAdminUser = "initial-admin"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "must be supplied together") {
		t.Fatalf("Validate() error = %v, want bootstrap pair error", err)
	}
}

func TestConfigValidateBootstrapConfiguration(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.BootstrapAdminUser = "initial-admin"
	cfg.BootstrapAdminPassword = strings.Repeat("b", 16)

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestConfigValidateRejectsBootstrapPasswordAboveBcryptLimit(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.BootstrapAdminUser = "initial-admin"
	cfg.BootstrapAdminPassword = strings.Repeat("b", 73)

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "72-byte limit") {
		t.Fatalf("Validate() error = %v, want bcrypt-length error", err)
	}
}

func TestConfigValidateAcceptanceRejectsBootstrapValues(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.AcceptanceMode = true
	cfg.BootstrapAdminUser = "initial-admin"
	cfg.BootstrapAdminPassword = strings.Repeat("b", 16)
	cfg.SeedAdminPassword = strings.Repeat("p", 12)
	cfg.SeedEditorPassword = strings.Repeat("q", 12)
	cfg.SeedReviewerPassword = strings.Repeat("r", 12)

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "not allowed in acceptance mode") {
		t.Fatalf("Validate() error = %v, want bootstrap rejection", err)
	}
}

func TestConfigValidateAcceptanceRequiresAllSeedPasswords(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.AcceptanceMode = true
	cfg.SeedAdminPassword = strings.Repeat("p", 12)
	cfg.SeedEditorPassword = strings.Repeat("q", 12)

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "SEED_REVIEWER_PASSWORD is required") {
		t.Fatalf("Validate() error = %v, want missing reviewer seed password", err)
	}
}

func TestConfigValidateAcceptanceConfiguration(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.AcceptanceMode = true
	cfg.SeedAdminPassword = strings.Repeat("p", 12)
	cfg.SeedEditorPassword = strings.Repeat("q", 12)
	cfg.SeedReviewerPassword = strings.Repeat("r", 12)

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestLoadConfigHasNoLocalRuntimeFallbacks(t *testing.T) {
	for _, key := range []string{
		"HTTP_ADDR",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"JWT_SECRET",
		"AES_KEY",
		"STORAGE_DIR",
		"BACKUP_DIR",
		"MIGRATIONS_DIR",
		"PUBLIC_DIR",
		"BOOTSTRAP_ADMIN_USERNAME",
		"BOOTSTRAP_ADMIN_PASSWORD",
		"SEED_ADMIN_PASSWORD",
		"SEED_EDITOR_PASSWORD",
		"SEED_REVIEWER_PASSWORD",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("ACCEPTANCE_MODE", "false")

	cfg := LoadConfig()
	if cfg.HTTPAddr != "" || cfg.DBPort != "" || cfg.DBUser != "" || cfg.DBName != "" {
		t.Fatalf("LoadConfig() supplied local network or database identity fallbacks")
	}
	if cfg.DBPassword != "" || cfg.JWTSecret != "" || cfg.AESKey != "" {
		t.Fatalf("LoadConfig() supplied a sensitive runtime fallback")
	}
	if cfg.StorageDir != "" || cfg.BackupDir != "" || cfg.MigrationsDir != "" || cfg.PublicDir != "" {
		t.Fatalf("LoadConfig() supplied local filesystem fallbacks")
	}
	if cfg.BootstrapAdminUser != "" || cfg.BootstrapAdminPassword != "" {
		t.Fatalf("LoadConfig() supplied a bootstrap-admin fallback")
	}
	if cfg.SeedAdminPassword != "" || cfg.SeedEditorPassword != "" || cfg.SeedReviewerPassword != "" {
		t.Fatalf("LoadConfig() supplied a seed-user fallback")
	}
}
