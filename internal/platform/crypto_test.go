package platform

import "testing"

func TestEncryptDecryptString(t *testing.T) {
	secret := "local-test-secret"
	plain := "sensitive coordinate reason"
	cipherText, err := EncryptString(secret, plain)
	if err != nil {
		t.Fatalf("EncryptString error: %v", err)
	}
	if cipherText == plain {
		t.Fatalf("ciphertext should not equal plaintext")
	}
	if len(cipherText) < len(EncryptedPrefix) || cipherText[:len(EncryptedPrefix)] != EncryptedPrefix {
		t.Fatalf("ciphertext missing prefix")
	}
	got, err := DecryptString(secret, cipherText)
	if err != nil {
		t.Fatalf("DecryptString error: %v", err)
	}
	if got != plain {
		t.Fatalf("DecryptString=%q want %q", got, plain)
	}
}

func TestDecryptPlainStringPassThrough(t *testing.T) {
	got, err := DecryptString("secret", "plain")
	if err != nil {
		t.Fatalf("plain decrypt returned error: %v", err)
	}
	if got != "plain" {
		t.Fatalf("plain decrypt=%q", got)
	}
}
