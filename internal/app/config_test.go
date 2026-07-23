package app

import (
	"strings"
	"testing"
)

func validRuntimeConfig() Config {
	return Config{
		HTTPAddr:       ":28080",
		DBPort:         "25432",
		DBUser:         "ironpage_test",
		DBPassword:     strings.Repeat("d", 16),
		DBName:         "ironpage_test",
		JWTSecret:      strings.Repeat("j", 32),
		AESKey:         strings.Repeat("a", 32),
		StorageDir:     "/runtime/storage",
		BackupDir:      "/runtime/backups",
		MigrationsDir:  "/runtime/app/migrations",
		PublicDir:      "/runtime/app/public",
		MaxUploadBytes: productMaxUploadBytes,
		MaxPDFPages:    productMaxPDFPages,
		MaxBatchFiles:  productMaxBatchFiles,
		MaxVersions:    productMaxVersions,
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

func TestConfigValidateRejectsInvalidNetworkConfiguration(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "malformed HTTP address", edit: func(cfg *Config) { cfg.HTTPAddr = "28080" }, want: "HTTP_ADDR must include a valid host and port"},
		{name: "non-numeric HTTP port", edit: func(cfg *Config) { cfg.HTTPAddr = ":http" }, want: "HTTP_ADDR port must be numeric"},
		{name: "privileged HTTP port", edit: func(cfg *Config) { cfg.HTTPAddr = ":80" }, want: "HTTP_ADDR port must be between 1024 and 65535"},
		{name: "non-numeric database port", edit: func(cfg *Config) { cfg.DBPort = "postgres" }, want: "DB_PORT port must be numeric"},
		{name: "out-of-range database port", edit: func(cfg *Config) { cfg.DBPort = "70000" }, want: "DB_PORT port must be between 1024 and 65535"},
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

func TestConfigValidateRequiresAbsoluteRuntimePaths(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "storage", edit: func(cfg *Config) { cfg.StorageDir = "storage" }, want: "STORAGE_DIR must be an absolute path"},
		{name: "backup", edit: func(cfg *Config) { cfg.BackupDir = "backups" }, want: "BACKUP_DIR must be an absolute path"},
		{name: "migrations", edit: func(cfg *Config) { cfg.MigrationsDir = "migrations" }, want: "MIGRATIONS_DIR must be an absolute path"},
		{name: "public", edit: func(cfg *Config) { cfg.PublicDir = "public" }, want: "PUBLIC_DIR must be an absolute path"},
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

func TestConfigValidateRejectsProductLimitOverrides(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "upload bytes", edit: func(cfg *Config) { cfg.MaxUploadBytes-- }, want: "MAX_UPLOAD_BYTES is fixed"},
		{name: "pdf pages", edit: func(cfg *Config) { cfg.MaxPDFPages-- }, want: "MAX_PDF_PAGES is fixed"},
		{name: "batch files", edit: func(cfg *Config) { cfg.MaxBatchFiles-- }, want: "MAX_BATCH_FILES is fixed"},
		{name: "versions", edit: func(cfg *Config) { cfg.MaxVersions-- }, want: "MAX_VERSIONS is fixed"},
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

func TestConfigValidateProductionRejectsSeedPasswords(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.SeedAdminPassword = strings.Repeat("p", 12)
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "requires ACCEPTANCE_MODE=true") {
		t.Fatalf("Validate() error = %v, want acceptance-mode error", err)
	}
}

func TestConfigValidateBootstrapPair(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.BootstrapAdminUser = "initial-admin"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "must be supplied together") {
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
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "72-byte limit") {
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
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "not allowed in acceptance mode") {
		t.Fatalf("Validate() error = %v, want bootstrap rejection", err)
	}
}

func TestConfigValidateAcceptanceRequiresAllSeedPasswords(t *testing.T) {
	cfg := validRuntimeConfig()
	cfg.AcceptanceMode = true
	cfg.SeedAdminPassword = strings.Repeat("p", 12)
	cfg.SeedEditorPassword = strings.Repeat("q", 12)
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "SEED_REVIEWER_PASSWORD is required") {
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
		"HTTP_ADDR", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"JWT_SECRET", "AES_KEY", "STORAGE_DIR", "BACKUP_DIR",
		"MIGRATIONS_DIR", "PUBLIC_DIR", "BOOTSTRAP_ADMIN_USERNAME",
		"BOOTSTRAP_ADMIN_PASSWORD", "SEED_ADMIN_PASSWORD",
		"SEED_EDITOR_PASSWORD", "SEED_REVIEWER_PASSWORD",
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

func TestLoadConfigIgnoresProductLimitEnvironmentOverrides(t *testing.T) {
	t.Setenv("MAX_UPLOAD_BYTES", "1")
	t.Setenv("MAX_PDF_PAGES", "1")
	t.Setenv("MAX_BATCH_FILES", "1")
	t.Setenv("MAX_VERSIONS", "1")
	cfg := LoadConfig()
	if cfg.MaxUploadBytes != productMaxUploadBytes || cfg.MaxPDFPages != productMaxPDFPages || cfg.MaxBatchFiles != productMaxBatchFiles || cfg.MaxVersions != productMaxVersions {
		t.Fatalf("LoadConfig() allowed product limit environment overrides: %#v", cfg)
	}
}
