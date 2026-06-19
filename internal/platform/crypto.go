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

const EncryptedPrefix = "enc:v1:"

func AESKeyFromSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

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
