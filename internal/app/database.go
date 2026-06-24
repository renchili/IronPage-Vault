package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

func EnsureSeedUsers(ctx context.Context, db *sqlx.DB, cfg Config) error {
	seeds := []struct {
		Username string
		Display  string
		Role     string
		Password string
	}{
		{"admin", "Default Admin", "Admin", cfg.SeedAdminPassword},
		{"editor", "Default Editor", "Editor", cfg.SeedEditorPassword},
		{"reviewer", "Default Reviewer", "Reviewer", cfg.SeedReviewerPassword},
	}
	for _, s := range seeds {
		var count int
		usernameKey := piiLookupKey(cfg.AESKey, s.Username)
		if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users WHERE username=$1 OR username=$2`, usernameKey, s.Username); err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		id := makeIdentifier("usr")
		hash, err := bcrypt.GenerateFromPassword([]byte(s.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		storedHash, err := sealPasswordHash(cfg.AESKey, hash)
		if err != nil {
			return err
		}
		usernameCipher, err := sealPII(cfg.AESKey, s.Username)
		if err != nil {
			return err
		}
		displayCipher, err := sealPII(cfg.AESKey, s.Display)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `INSERT INTO users(id, username, username_ciphertext, display_name, display_name_ciphertext, role, password_hash, created_at) VALUES($1,$2,$3,'',$4,$5,$6,NOW())`, id, usernameKey, usernameCipher, displayCipher, s.Role, storedHash)
		if err != nil {
			return err
		}
	}
	return nil
}
