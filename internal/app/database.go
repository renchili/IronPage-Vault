package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func OpenDatabase(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

func RunMigrations(db *sqlx.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(raw)); err != nil {
			return fmt.Errorf("migration %s: %w", path, err)
		}
	}
	return nil
}

// EnsureInitialUsers creates acceptance fixtures only in acceptance mode. In
// normal mode it creates a single externally configured Admin only when the
// user table is empty.
func EnsureInitialUsers(ctx context.Context, db *sqlx.DB, cfg Config) error {
	if cfg.AcceptanceMode {
		return EnsureSeedUsers(ctx, db, cfg)
	}

	var count int
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if strings.TrimSpace(cfg.BootstrapAdminUser) == "" || strings.TrimSpace(cfg.BootstrapAdminPassword) == "" {
		return fmt.Errorf("empty user store requires BOOTSTRAP_ADMIN_USERNAME and BOOTSTRAP_ADMIN_PASSWORD")
	}
	return insertLocalUser(ctx, db, cfg, cfg.BootstrapAdminUser, "Initial Admin", RoleAdmin, cfg.BootstrapAdminPassword)
}

// EnsureSeedUsers creates local fixture identities only for an explicitly
// enabled acceptance environment.
func EnsureSeedUsers(ctx context.Context, db *sqlx.DB, cfg Config) error {
	if !cfg.AcceptanceMode {
		return fmt.Errorf("seed users require acceptance mode")
	}
	seeds := []struct {
		username string
		display  string
		role     string
		password string
	}{
		{"admin", "Acceptance Admin", RoleAdmin, cfg.SeedAdminPassword},
		{"editor", "Acceptance Editor", RoleEditor, cfg.SeedEditorPassword},
		{"reviewer", "Acceptance Reviewer", RoleReviewer, cfg.SeedReviewerPassword},
	}
	for _, seed := range seeds {
		usernameKey := piiLookupKey(cfg.AESKey, seed.username)
		var count int
		if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users WHERE username=$1 OR username=$2`, usernameKey, seed.username); err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := insertLocalUser(ctx, db, cfg, seed.username, seed.display, seed.role, seed.password); err != nil {
			return err
		}
	}
	return nil
}

func insertLocalUser(ctx context.Context, db *sqlx.DB, cfg Config, username, display, role, password string) error {
	id := makeIdentifier("usr")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	storedHash, err := sealPasswordHash(cfg.AESKey, hash)
	if err != nil {
		return err
	}
	usernameKey := piiLookupKey(cfg.AESKey, username)
	usernameCipher, err := sealPII(cfg.AESKey, username)
	if err != nil {
		return err
	}
	displayCipher, err := sealPII(cfg.AESKey, display)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO users(id, username, username_ciphertext, display_name, display_name_ciphertext, role, password_hash, created_at) VALUES($1,$2,$3,'',$4,$5,$6,NOW())`, id, usernameKey, usernameCipher, displayCipher, role, storedHash)
	return err
}
