package app

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEncryptedAuditMetadata(t *testing.T) {
	payload, err := encryptedAuditMetadata("test-aes-secret", map[string]interface{}{
		"subject_name": "Alice Example",
		"matter_ref":   "MAT-12345",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(payload, "Alice Example") || strings.Contains(payload, "MAT-12345") {
		t.Fatalf("plaintext metadata leaked into payload: %s", payload)
	}

	var protected protectedMetadata
	if err := json.Unmarshal([]byte(payload), &protected); err != nil {
		t.Fatal(err)
	}
	if protected.Algorithm != "AES-256-GCM" {
		t.Fatalf("algorithm=%q", protected.Algorithm)
	}
	plain, err := decryptString("test-aes-secret", protected.Ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plain, "Alice Example") || !strings.Contains(plain, "MAT-12345") {
		t.Fatalf("decrypted metadata=%s", plain)
	}
}

func TestEncryptedAuditMetadataEmpty(t *testing.T) {
	payload, err := encryptedAuditMetadata("test-aes-secret", nil)
	if err != nil {
		t.Fatal(err)
	}
	if payload != "{}" {
		t.Fatalf("payload=%q", payload)
	}
}
