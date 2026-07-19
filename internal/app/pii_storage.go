package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const piiLookupPrefix = "lookup:v1:"

func piiLookupKey(secret string, value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	sum := sha256.Sum256([]byte("ironpage-vault:pii-lookup:v1:" + secret + ":" + normalized))
	return piiLookupPrefix + hex.EncodeToString(sum[:])
}

func sealPII(secret string, plaintext string) (string, error) {
	return encryptString(secret, plaintext)
}

func openPII(secret string, ciphertext string, legacyPlain string) (string, error) {
	if strings.TrimSpace(ciphertext) != "" {
		return decryptString(secret, ciphertext)
	}
	return decryptString(secret, legacyPlain)
}

func openUserPII(secret string, u *User) error {
	username, err := openPII(secret, u.UsernameCiphertext, u.Username)
	if err != nil {
		return err
	}
	displayName, err := openPII(secret, u.DisplayNameCiphertext, u.DisplayName)
	if err != nil {
		return err
	}
	u.Username = username
	u.DisplayName = displayName
	return nil
}

func openDocumentPII(secret string, d *Document) error {
	title, err := openPII(secret, d.TitleCiphertext, d.Title)
	if err != nil {
		return err
	}
	d.Title = title
	return nil
}

func openNotificationPII(secret string, n *notificationResponse) error {
	message, err := openPII(secret, n.MessageCiphertext, n.Message)
	if err != nil {
		return err
	}
	n.Message = message
	return nil
}

func sealAuditSourceIP(secret string, sourceIP string) (string, error) {
	return sealPII(secret, sourceIP)
}

func sealAuditMetadata(secret string, metadata map[string]interface{}) (string, error) {
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	plain, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return sealPII(secret, string(plain))
}

func openAuditPII(secret string, row *auditLogResponse) error {
	sourceIP := row.SourceIP
	if strings.TrimSpace(row.SourceIPCiphertext) != "" {
		plain, err := decryptString(secret, row.SourceIPCiphertext)
		if err != nil {
			return err
		}
		sourceIP = plain
	}
	metadata := string(row.Metadata)
	if strings.TrimSpace(row.MetadataCiphertext) != "" {
		plain, err := decryptString(secret, row.MetadataCiphertext)
		if err != nil {
			return err
		}
		metadata = plain
	}
	if strings.TrimSpace(metadata) == "" {
		metadata = `{}`
	}
	if !json.Valid([]byte(metadata)) {
		return &json.SyntaxError{}
	}
	row.SourceIP = sourceIP
	row.Metadata = json.RawMessage(metadata)
	return nil
}
