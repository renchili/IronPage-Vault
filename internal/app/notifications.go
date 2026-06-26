package app

import (
	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/core"
)

const maxUnreadNotifications = core.MaxUnreadNotifications

// notificationTrimCount delegates unread-cap arithmetic to core.
func notificationTrimCount(unread int, limit int) int {
	return core.NotificationTrimCount(unread, limit)
}

// createNotification stores a local in-app notification for a user.
//
// The unread cap is enforced before insert, and the message is sealed before it
// is persisted so plaintext notification content is not stored in the database.
func (a *App) createNotification(c echo.Context, userID, documentID, templateKey, message string) error {
	var unread int
	if err := a.db.GetContext(c.Request().Context(), &unread, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read_at IS NULL`, userID); err != nil {
		return err
	}
	if trim := notificationTrimCount(unread, maxUnreadNotifications); trim > 0 {
		_, err := a.db.ExecContext(c.Request().Context(), `UPDATE notifications SET read_at=NOW() WHERE id IN (SELECT id FROM notifications WHERE user_id=$1 AND read_at IS NULL ORDER BY created_at ASC LIMIT $2)`, userID, trim)
		if err != nil {
			return err
		}
	}
	messageCipher, err := sealPII(a.cfg.AESKey, message)
	if err != nil {
		return err
	}
	_, err = a.db.ExecContext(c.Request().Context(), `INSERT INTO notifications(id,user_id,document_id,template_key,message,message_ciphertext,created_at) VALUES($1,$2,$3,$4,'',$5,NOW())`, makeIdentifier("not"), userID, documentID, templateKey, messageCipher)
	return err
}
