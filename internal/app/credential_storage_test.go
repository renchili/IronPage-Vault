package app

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashIsSealedAndBcryptCompatible(t *testing.T) {
	rawHash, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate bcrypt hash: %v", err)
	}

	stored, err := sealPasswordHash("test-aes-key", rawHash)
	if err != nil {
		t.Fatalf("seal password hash: %v", err)
	}
	if !strings.HasPrefix(stored, encryptedPrefix) {
		t.Fatalf("stored hash must use encrypted prefix, got %q", stored)
	}
	if strings.Contains(stored, string(rawHash)) {
		t.Fatalf("stored hash must not contain plaintext bcrypt verifier")
	}

	opened, err := openPasswordHash("test-aes-key", stored)
	if err != nil {
		t.Fatalf("open password hash: %v", err)
	}
	if opened != string(rawHash) {
		t.Fatalf("opened hash mismatch")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(opened), []byte("Admin123!")); err != nil {
		t.Fatalf("opened hash must remain bcrypt-compatible: %v", err)
	}
}

func TestPasswordHashLegacyFallback(t *testing.T) {
	rawHash, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate bcrypt hash: %v", err)
	}

	opened, err := openPasswordHash("test-aes-key", string(rawHash))
	if err != nil {
		t.Fatalf("legacy plaintext bcrypt hash should still open: %v", err)
	}
	if opened != string(rawHash) {
		t.Fatalf("legacy plaintext bcrypt hash should be returned unchanged")
	}
}
