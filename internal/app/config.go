package app

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	productMaxUploadBytes int64 = 200 * 1024 * 1024
	productMaxPDFPages          = 500
	productMaxBatchFiles        = 250
	productMaxVersions          = 50
)

// Config contains runtime settings for the local IronPage Vault service.
type Config struct {
	HTTPAddr               string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	JWTSecret              string
	AESKey                 string
	StorageDir             string
	BackupDir              string
	MigrationsDir          string
	PublicDir              string
	SessionTTL             time.Duration
	RequestMaxAge          time.Duration
	MaxUploadBytes         int64
	MaxPDFPages            int
	MaxBatchFiles          int
	MaxVersions            int
	DefaultPageSize        int
	MaxPageSize            int
	AcceptanceMode         bool
	BootstrapAdminUser     string
	BootstrapAdminPassword string
	SeedAdminPassword      string
	SeedEditorPassword     string
	SeedReviewerPassword   string
}

// LoadConfig reads deployment-owned runtime configuration. Product limits are
// fixed by the project specification; local identity, ports, paths, secrets,
// and credentials have no application fallback.
func LoadConfig() Config {
	return Config{
		HTTPAddr:               env("HTTP_ADDR", ""),
		DBPort:                 env("DB_PORT", ""),
		DBUser:                 env("DB_USER", ""),
		DBPassword:             env("DB_PASSWORD", ""),
		DBName:                 env("DB_NAME", ""),
		JWTSecret:              env("JWT_SECRET", ""),
		AESKey:                 env("AES_KEY", ""),
		StorageDir:             env("STORAGE_DIR", ""),
		BackupDir:              env("BACKUP_DIR", ""),
		MigrationsDir:          env("MIGRATIONS_DIR", ""),
		PublicDir:              env("PUBLIC_DIR", ""),
		SessionTTL:             8 * time.Hour,
		RequestMaxAge:          60 * time.Second,
		MaxUploadBytes:         productMaxUploadBytes,
		MaxPDFPages:            productMaxPDFPages,
		MaxBatchFiles:          productMaxBatchFiles,
		MaxVersions:            productMaxVersions,
		DefaultPageSize:        25,
		MaxPageSize:            100,
		AcceptanceMode:         envBool("ACCEPTANCE_MODE", false),
		BootstrapAdminUser:     env("BOOTSTRAP_ADMIN_USERNAME", ""),
		BootstrapAdminPassword: env("BOOTSTRAP_ADMIN_PASSWORD", ""),
		SeedAdminPassword:      env("SEED_ADMIN_PASSWORD", ""),
		SeedEditorPassword:     env("SEED_EDITOR_PASSWORD", ""),
		SeedReviewerPassword:   env("SEED_REVIEWER_PASSWORD", ""),
	}
}

// Validate rejects insecure or incomplete runtime configuration before the
// service creates directories, connects to PostgreSQL, or serves HTTP routes.
func (c Config) Validate() error {
	for _, item := range []struct {
		name  string
		value string
	}{
		{"HTTP_ADDR", c.HTTPAddr},
		{"DB_PORT", c.DBPort},
		{"DB_USER", c.DBUser},
		{"DB_NAME", c.DBName},
		{"STORAGE_DIR", c.StorageDir},
		{"BACKUP_DIR", c.BackupDir},
		{"MIGRATIONS_DIR", c.MigrationsDir},
		{"PUBLIC_DIR", c.PublicDir},
	} {
		if err := requireValue(item.name, item.value); err != nil {
			return err
		}
	}

	_, httpPort, err := net.SplitHostPort(c.HTTPAddr)
	if err != nil {
		return fmt.Errorf("HTTP_ADDR must include a valid host and port: %w", err)
	}
	if err := validatePort("HTTP_ADDR", httpPort); err != nil {
		return err
	}
	if err := validatePort("DB_PORT", c.DBPort); err != nil {
		return err
	}

	for _, item := range []struct {
		name string
		path string
	}{
		{"STORAGE_DIR", c.StorageDir},
		{"BACKUP_DIR", c.BackupDir},
		{"MIGRATIONS_DIR", c.MigrationsDir},
		{"PUBLIC_DIR", c.PublicDir},
	} {
		if !filepath.IsAbs(item.path) {
			return fmt.Errorf("%s must be an absolute path", item.name)
		}
	}

	if err := requireSecret("DB_PASSWORD", c.DBPassword, 16); err != nil {
		return err
	}
	if err := requireSecret("JWT_SECRET", c.JWTSecret, 32); err != nil {
		return err
	}
	if err := requireSecret("AES_KEY", c.AESKey, 32); err != nil {
		return err
	}
	if c.MaxUploadBytes != productMaxUploadBytes {
		return fmt.Errorf("MAX_UPLOAD_BYTES is fixed at %d", productMaxUploadBytes)
	}
	if c.MaxPDFPages != productMaxPDFPages {
		return fmt.Errorf("MAX_PDF_PAGES is fixed at %d", productMaxPDFPages)
	}
	if c.MaxBatchFiles != productMaxBatchFiles {
		return fmt.Errorf("MAX_BATCH_FILES is fixed at %d", productMaxBatchFiles)
	}
	if c.MaxVersions != productMaxVersions {
		return fmt.Errorf("MAX_VERSIONS is fixed at %d", productMaxVersions)
	}

	seedValues := []struct {
		name  string
		value string
	}{
		{"SEED_ADMIN_PASSWORD", c.SeedAdminPassword},
		{"SEED_EDITOR_PASSWORD", c.SeedEditorPassword},
		{"SEED_REVIEWER_PASSWORD", c.SeedReviewerPassword},
	}
	bootstrapUser := strings.TrimSpace(c.BootstrapAdminUser)
	bootstrapPassword := strings.TrimSpace(c.BootstrapAdminPassword)

	if c.AcceptanceMode {
		if bootstrapUser != "" || bootstrapPassword != "" {
			return fmt.Errorf("bootstrap admin values are not allowed in acceptance mode")
		}
		for _, seed := range seedValues {
			if err := requireBcryptSecret(seed.name, seed.value, 12); err != nil {
				return fmt.Errorf("acceptance mode: %w", err)
			}
		}
		return nil
	}

	for _, seed := range seedValues {
		if strings.TrimSpace(seed.value) != "" {
			return fmt.Errorf("%s requires ACCEPTANCE_MODE=true", seed.name)
		}
	}
	if (bootstrapUser == "") != (bootstrapPassword == "") {
		return fmt.Errorf("BOOTSTRAP_ADMIN_USERNAME and BOOTSTRAP_ADMIN_PASSWORD must be supplied together")
	}
	if bootstrapPassword != "" {
		if err := requireBcryptSecret("BOOTSTRAP_ADMIN_PASSWORD", bootstrapPassword, 16); err != nil {
			return err
		}
	}
	return nil
}

func (c Config) DSN() string {
	return strings.Join([]string{
		"port=" + c.DBPort,
		"user=" + c.DBUser,
		"password=" + c.DBPassword,
		"dbname=" + c.DBName,
		"sslmode=disable",
	}, " ")
}

func validatePort(name, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s port must be numeric", name)
	}
	if port < 1024 || port > 65535 {
		return fmt.Errorf("%s port must be between 1024 and 65535", name)
	}
	return nil
}

func requireValue(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
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

func requireBcryptSecret(name, value string, minimumLength int) error {
	if err := requireSecret(name, value, minimumLength); err != nil {
		return err
	}
	if len([]byte(strings.TrimSpace(value))) > 72 {
		return fmt.Errorf("%s must not exceed bcrypt's 72-byte limit", name)
	}
	return nil
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
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
