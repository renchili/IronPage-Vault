package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (a *App) readNotificationChecked(c echo.Context) error {
	p := principal(c)
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()

	var documentID string
	if err := tx.GetContext(c.Request().Context(), &documentID, `SELECT COALESCE(document_id,'') FROM notifications WHERE id=$1 AND user_id=$2 FOR UPDATE`, c.Param("id"), p.UserID); err != nil {
		return apiErr(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "notification not found")
	}
	result, err := tx.ExecContext(c.Request().Context(), `UPDATE notifications SET read_at=NOW() WHERE id=$1 AND user_id=$2`, c.Param("id"), p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_UPDATE_ERROR", "could not mark read")
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_UPDATE_ERROR", "could not verify notification update")
	}
	if changed != 1 {
		return apiErr(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "notification not found")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "NOTIFICATION_READ", documentID, map[string]interface{}{"notification_id": c.Param("id")}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record notification acknowledgement")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit notification acknowledgement")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "read"})
}
