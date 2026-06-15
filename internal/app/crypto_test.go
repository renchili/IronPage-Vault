package app

import "testing"

func TestEncryptDecryptString(t *testing.T) {
    secret := "local-test-secret"
    plain := "sensitive coordinate reason"
    cipherText, err := encryptString(secret, plain)
    if err != nil { t.Fatalf("encryptString error: %v", err) }
    if cipherText == plain { t.Fatalf("ciphertext should not equal plaintext") }
    if len(cipherText) < len(encryptedPrefix) || cipherText[:len(encryptedPrefix)] != encryptedPrefix { t.Fatalf("ciphertext missing prefix") }
    got, err := decryptString(secret, cipherText)
    if err != nil { t.Fatalf("decryptString error: %v", err) }
    if got != plain { t.Fatalf("decryptString=%q want %q", got, plain) }
}

func TestDecryptPlainStringPassThrough(t *testing.T) {
    got, err := decryptString("secret", "plain")
    if err != nil { t.Fatalf("plain decrypt returned error: %v", err) }
    if got != "plain" { t.Fatalf("plain decrypt=%q", got) }
}
