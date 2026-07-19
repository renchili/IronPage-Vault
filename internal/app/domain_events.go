package app

import (
	"context"
	"encoding/json"

	"github.com/jmoiron/sqlx"
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

func (a *App) insertAuditRecordWithExecutor(ctx context.Context, executor sqlx.ExtContext, actorID, action, documentID, requestID, sourceIP string, metadata map[string]interface{}) error {
	sourceIPCipher, err := sealAuditSourceIP(a.cfg.AESKey, sourceIP)
	if err != nil {
		return err
	}
	metadataCipher, err := sealAuditMetadata(a.cfg.AESKey, metadata)
	if err != nil {
		return err
	}
	sourceIPLookup := piiLookupKey(a.cfg.AESKey, sourceIP)
	_, err = executor.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,source_ip_lookup,source_ip_ciphertext,metadata,metadata_ciphertext,created_at) VALUES($1,NULLIF($2,''),NULLIF($3,''),$4,$5,'',$6,$7,'{}'::jsonb,$8,NOW())`, makeIdentifier("aud"), actorID, documentID, action, requestID, sourceIPLookup, sourceIPCipher, metadataCipher)
	return err
}

func (a *App) insertAuditRecord(ctx context.Context, actorID, action, documentID, requestID, sourceIP string, metadata map[string]interface{}) error {
	return a.insertAuditRecordWithExecutor(ctx, a.db, actorID, action, documentID, requestID, sourceIP, metadata)
}

func (a *App) audit(c echo.Context, actorID, action, documentID string, metadata map[string]interface{}) error {
	return a.insertAuditRecord(c.Request().Context(), actorID, action, documentID, currentRequestID(c), c.RealIP(), metadata)
}

func (a *App) auditWithExecutor(c echo.Context, executor sqlx.ExtContext, actorID, action, documentID string, metadata map[string]interface{}) error {
	return a.insertAuditRecordWithExecutor(c.Request().Context(), executor, actorID, action, documentID, currentRequestID(c), c.RealIP(), metadata)
}

func (a *App) notifyUser(c echo.Context, userID, templateKey, message string, documentID string) error {
	return a.createNotification(c, userID, documentID, templateKey, message)
}

func (a *App) notifyUserWithExecutor(c echo.Context, executor sqlx.ExtContext, userID, templateKey, message string, documentID string) error {
	return a.createNotificationWithExecutor(c.Request().Context(), executor, userID, documentID, templateKey, message)
}
