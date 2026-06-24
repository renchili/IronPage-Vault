package app

func sealPasswordHash(secret string, bcryptHash []byte) (string, error) {
	return encryptString(secret, string(bcryptHash))
}

func openPasswordHash(secret string, storedHash string) (string, error) {
	return decryptString(secret, storedHash)
}
