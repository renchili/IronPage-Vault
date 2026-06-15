package app

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr          string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	JWTSecret         string
	AESKey            string
	StorageDir        string
	BackupDir         string
	MigrationsDir     string
	PublicDir         string
	SessionTTL        time.Duration
	RequestMaxAge     time.Duration
	MaxUploadBytes    int64
	MaxPDFPages       int
	MaxBatchFiles     int
	MaxVersions       int
	DefaultPageSize   int
	MaxPageSize       int
	SeedAdminPassword string
	SeedEditorPassword string
	SeedReviewerPassword string
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
		DBHost:            env("DB_HOST", "127.0.0.1"),
		DBPort:            env("DB_PORT", "5432"),
		DBUser:            env("DB_USER", "ironpage"),
		DBPassword:        env("DB_PASSWORD", "ironpage"),
		DBName:            env("DB_NAME", "ironpage"),
		JWTSecret:         env("JWT_SECRET", "local-dev-change-me-32-bytes-minimum"),
		AESKey:            env("AES_KEY", "local-dev-aes-key-change-me"),
		StorageDir:        env("STORAGE_DIR", "/var/lib/ironpage/storage"),
		BackupDir:         env("BACKUP_DIR", "/var/lib/ironpage/backups"),
		MigrationsDir:     env("MIGRATIONS_DIR", "migrations"),
		PublicDir:         env("PUBLIC_DIR", "public"),
		SessionTTL:        8 * time.Hour,
		RequestMaxAge:     60 * time.Second,
		MaxUploadBytes:    int64(envInt("MAX_UPLOAD_BYTES", 200*1024*1024)),
		MaxPDFPages:       envInt("MAX_PDF_PAGES", 500),
		MaxBatchFiles:     envInt("MAX_BATCH_FILES", 250),
		MaxVersions:       envInt("MAX_VERSIONS", 50),
		DefaultPageSize:   25,
		MaxPageSize:       100,
		SeedAdminPassword: env("SEED_ADMIN_PASSWORD", "Admin123!"),
		SeedEditorPassword: env("SEED_EDITOR_PASSWORD", "Editor123!"),
		SeedReviewerPassword: env("SEED_REVIEWER_PASSWORD", "Reviewer123!"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
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
