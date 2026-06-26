package platform

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// FileDigest streams r and returns its SHA-256 digest as lowercase hex.
func FileDigest(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
