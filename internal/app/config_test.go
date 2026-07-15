package app

import (
	"strings"
	"testing"
)

func validRuntimeConfig() Config {
	return Config{
		DBPassword: strings.Repeat("d", 16),
		JWTSecret:  strings.Repeat("j", 32),
		AESKey:     strings.Repeat("a", 32),
	}
}

func TestConfigValidateRequiresRuntimeSecrets(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{
			name: "database password",
			edit: func(cfg *Config) { cfg.DBPassword = "" },
			want: "DB_PASSWORD is required",
		},
		{
			name: "jwt secret",
			edit: func(cfg *Config) { cfg.JWTSecret = "" },
			want: "JWT_SECRET is required",
		},
		{
			name: "aes key",
			edit: func(cfg *Config) { cfg.AESKey = "" },
			want: "AES_KEY is required",
		},
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

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "requires ACCEPTANCE_MODE=true") {
		t.Fatalf("Validate() error = %v, want acceptance-mode error", err)
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

func TestLoadConfigHasNoSensitiveFallbacks(t *testing.T) {
	for _, key := range []string{
		"DB_PASSWORD",
		"JWT_SECRET",
		"AES_KEY",
		"SEED_ADMIN_PASSWORD",
		"SEED_EDITOR_PASSWORD",
		"SEED_REVIEWER_PASSWORD",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("ACCEPTANCE_MODE", "false")

	cfg := LoadConfig()
	if cfg.DBPassword != "" || cfg.JWTSecret != "" || cfg.AESKey != "" {
		t.Fatalf("LoadConfig() supplied a sensitive runtime fallback")
	}
	if cfg.SeedAdminPassword != "" || cfg.SeedEditorPassword != "" || cfg.SeedReviewerPassword != "" {
		t.Fatalf("LoadConfig() supplied a seed-user fallback")
	}
}
