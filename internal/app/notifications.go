package app

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/core"
)

const maxUnreadNotifications = core.MaxUnreadNotifications

// notificationTrimCount delegates unread-cap arithmetic to core.
func notificationTrimCount(unread int, limit int) int {
	return core.NotificationTrimCount(unread, limit)
}

// createNotificationWithExecutor stores a local in-app notification using the
// caller's database boundary. Passing a transaction keeps notification trimming
// and insertion atomic with the parent mutation.
func (a *App) createNotificationWithExecutor(ctx context.Context, executor sqlx.ExtContext, userID, documentID, templateKey, message string) error {
	var unread int
	if err := sqlx.GetContext(ctx, executor, &unread, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read_at IS NULL`, userID); err != nil {
		return err
	}
	if trim := notificationTrimCount(unread, maxUnreadNotifications); trim > 0 {
		if _, err := executor.ExecContext(ctx, `UPDATE notifications SET read_at=NOW() WHERE id IN (SELECT id FROM notifications WHERE user_id=$1 AND read_at IS NULL ORDER BY created_at ASC LIMIT $2)`, userID, trim); err != nil {
			return err
		}
	}
	messageCipher, err := sealPII(a.cfg.AESKey, message)
	if err != nil {
		return err
	}
	_, err = executor.ExecContext(ctx, `INSERT INTO notifications(id,user_id,document_id,template_key,message,message_ciphertext,created_at) VALUES($1,$2,NULLIF($3,''),$4,'',$5,NOW())`, makeIdentifier("not"), userID, documentID, templateKey, messageCipher)
	return err
}

func (a *App) createNotification(c echo.Context, userID, documentID, templateKey, message string) error {
	return a.createNotificationWithExecutor(c.Request().Context(), a.db, userID, documentID, templateKey, message)
}
