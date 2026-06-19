package app

import "ironpage-vault/internal/platform"

const encryptedPrefix = platform.EncryptedPrefix

func aesKeyFromSecret(secret string) []byte { return platform.AESKeyFromSecret(secret) }
func encryptString(secret string, plaintext string) (string, error) {
	return platform.EncryptString(secret, plaintext)
}
func decryptString(secret string, ciphertext string) (string, error) {
	return platform.DecryptString(secret, ciphertext)
}
