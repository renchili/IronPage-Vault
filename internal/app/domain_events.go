package app

import (
	"context"
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
	metadataCipher, err := sealAuditMetadata(secret, metadata)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(protectedMetadata{Algorithm: "AES-256-GCM", Ciphertext: metadataCipher})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func (a *App) insertAuditRecord(ctx context.Context, actorID, action, documentID, requestID, sourceIP string, metadata map[string]interface{}) error {
	sourceIPCipher, err := sealAuditSourceIP(a.cfg.AESKey, sourceIP)
	if err != nil {
		return err
	}
	metadataCipher, err := sealAuditMetadata(a.cfg.AESKey, metadata)
	if err != nil {
		return err
	}
	if documentID == "" {
		_, err = a.db.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,action_type,request_id,source_ip,source_ip_ciphertext,metadata,metadata_ciphertext,created_at) VALUES($1,NULLIF($2,''),$3,$4,'',$5,'{}'::jsonb,$6,NOW())`, makeIdentifier("aud"), actorID, action, requestID, sourceIPCipher, metadataCipher)
		return err
	}
	_, err = a.db.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,source_ip_ciphertext,metadata,metadata_ciphertext,created_at) VALUES($1,NULLIF($2,''),$3,$4,$5,'',$6,'{}'::jsonb,$7,NOW())`, makeIdentifier("aud"), actorID, documentID, action, requestID, sourceIPCipher, metadataCipher)
	return err
}

func (a *App) audit(c echo.Context, actorID, action, documentID string, metadata map[string]interface{}) {
	_ = a.insertAuditRecord(c.Request().Context(), actorID, action, documentID, currentRequestID(c), c.RealIP(), metadata)
}

func (a *App) notifyUser(c echo.Context, userID, templateKey, message string, documentID string) {
	_ = a.createNotification(c, userID, documentID, templateKey, message)
}
