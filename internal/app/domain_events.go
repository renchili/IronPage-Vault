package app

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

type protectedMetadata struct {
	Algorithm  string `json:"algorithm"`
	Ciphertext string `json:"ciphertext"`
}

func encryptedAuditMetadata(secret string, metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return `{}`, nil
	}
	plain, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	ciphertext, err := encryptString(secret, string(plain))
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(protectedMetadata{Algorithm: "AES-256-GCM", Ciphertext: ciphertext})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func (a *App) audit(c echo.Context, actorID, action, documentID string, metadata map[string]interface{}) {
	payload, err := encryptedAuditMetadata(a.cfg.AESKey, metadata)
	if err != nil {
		return
	}
	if documentID == "" {
		_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,$4,$5,$6::jsonb,NOW())`, makeIdentifier("aud"), actorID, action, currentRequestID(c), c.RealIP(), payload)
		return
	}
	_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,$4,$5,$6,$7::jsonb,NOW())`, makeIdentifier("aud"), actorID, documentID, action, currentRequestID(c), c.RealIP(), payload)
}

func (a *App) notifyUser(c echo.Context, userID, templateKey, message string, documentID string) {
	_ = a.createNotification(c, userID, documentID, templateKey, message)
}
