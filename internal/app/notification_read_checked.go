package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (a *App) readNotificationChecked(c echo.Context) error {
	p := principal(c)
	result, err := a.db.ExecContext(c.Request().Context(), `UPDATE notifications SET read_at=NOW() WHERE id=$1 AND user_id=$2`, c.Param("id"), p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_UPDATE_ERROR", "could not mark read")
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_UPDATE_ERROR", "could not verify notification update")
	}
	if changed == 0 {
		return apiErr(c, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "notification not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "read"})
}
