package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contains runtime settings for the local IronPage Vault service.
type Config struct {
	HTTPAddr             string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	JWTSecret            string
	AESKey               string
	StorageDir           string
	BackupDir            string
	MigrationsDir        string
	PublicDir            string
	SessionTTL           time.Duration
	RequestMaxAge        time.Duration
	MaxUploadBytes       int64
	MaxPDFPages          int
	MaxBatchFiles        int
	MaxVersions          int
	DefaultPageSize      int
	MaxPageSize          int
	AcceptanceMode       bool
	SeedAdminPassword    string
	SeedEditorPassword   string
	SeedReviewerPassword string
}

// LoadConfig reads runtime configuration without providing fallback values for
// passwords, signing material, encryption material, or acceptance identities.
func LoadConfig() Config {
	return Config{
		HTTPAddr:             env("HTTP_ADDR", ":8080"),
		DBHost:               env("DB_HOST", "127.0.0.1"),
		DBPort:               env("DB_PORT", "5432"),
		DBUser:               env("DB_USER", "ironpage"),
		DBPassword:           env("DB_PASSWORD", ""),
		DBName:               env("DB_NAME", "ironpage"),
		JWTSecret:            env("JWT_SECRET", ""),
		AESKey:               env("AES_KEY", ""),
		StorageDir:           env("STORAGE_DIR", "/var/lib/ironpage/storage"),
		BackupDir:            env("BACKUP_DIR", "/var/lib/ironpage/backups"),
		MigrationsDir:        env("MIGRATIONS_DIR", "migrations"),
		PublicDir:            env("PUBLIC_DIR", "public"),
		SessionTTL:           8 * time.Hour,
		RequestMaxAge:        60 * time.Second,
		MaxUploadBytes:       int64(envInt("MAX_UPLOAD_BYTES", 200*1024*1024)),
		MaxPDFPages:          envInt("MAX_PDF_PAGES", 500),
		MaxBatchFiles:        envInt("MAX_BATCH_FILES", 250),
		MaxVersions:          envInt("MAX_VERSIONS", 50),
		DefaultPageSize:      25,
		MaxPageSize:          100,
		AcceptanceMode:       envBool("ACCEPTANCE_MODE", false),
		SeedAdminPassword:    env("SEED_ADMIN_PASSWORD", ""),
		SeedEditorPassword:   env("SEED_EDITOR_PASSWORD", ""),
		SeedReviewerPassword: env("SEED_REVIEWER_PASSWORD", ""),
	}
}

// Validate rejects insecure or incomplete runtime configuration before the
// service creates directories, connects to PostgreSQL, or serves HTTP routes.
func (c Config) Validate() error {
	if err := requireSecret("DB_PASSWORD", c.DBPassword, 16); err != nil {
		return err
	}
	if err := requireSecret("JWT_SECRET", c.JWTSecret, 32); err != nil {
		return err
	}
	if err := requireSecret("AES_KEY", c.AESKey, 32); err != nil {
		return err
	}

	seedValues := []struct {
		name  string
		value string
	}{
		{"SEED_ADMIN_PASSWORD", c.SeedAdminPassword},
		{"SEED_EDITOR_PASSWORD", c.SeedEditorPassword},
		{"SEED_REVIEWER_PASSWORD", c.SeedReviewerPassword},
	}

	if !c.AcceptanceMode {
		for _, seed := range seedValues {
			if strings.TrimSpace(seed.value) != "" {
				return fmt.Errorf("%s requires ACCEPTANCE_MODE=true", seed.name)
			}
		}
		return nil
	}

	for _, seed := range seedValues {
		if err := requireSecret(seed.name, seed.value, 12); err != nil {
			return fmt.Errorf("acceptance mode: %w", err)
		}
	}
	return nil
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func requireSecret(name, value string, minimumLength int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s is required", name)
	}
	if len(trimmed) < minimumLength {
		return fmt.Errorf("%s must be at least %d characters", name, minimumLength)
	}
	return nil
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func envInt(k string, fallback int) int {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envBool(k string, fallback bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
