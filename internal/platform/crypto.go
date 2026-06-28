package platform

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// EncryptedPrefix marks values sealed by the current local encryption format.
const EncryptedPrefix = "enc:v1:"

// AESKeyFromSecret derives a fixed-size AES key from a local deployment secret.
//
// The caller remains responsible for sourcing the secret from configuration or a
// local secret provider rather than hard-coding production material.
func AESKeyFromSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// EncryptString seals plaintext using AES-GCM and prefixes the stored value.
//
// The nonce is generated from crypto/rand and prepended to the ciphertext so the
// result is self-contained for local database storage.
func EncryptString(secret string, plaintext string) (string, error) {
	block, err := aes.NewCipher(AESKeyFromSecret(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return EncryptedPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptString opens values produced by EncryptString.
//
// Values without EncryptedPrefix are returned unchanged to support migration of
// legacy plaintext rows while handlers are moved to ciphertext-only storage.
func DecryptString(secret string, ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, EncryptedPrefix) {
		return ciphertext, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, EncryptedPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(AESKeyFromSecret(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce := raw[:gcm.NonceSize()]
	body := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
