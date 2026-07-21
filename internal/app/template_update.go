package app

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (a *App) patchNotificationTemplate(c echo.Context) error {
	p := principal(c)
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		return apiErr(c, http.StatusBadRequest, "TEMPLATE_KEY_REQUIRED", "template key is required")
	}
	var req struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_TEMPLATE_REQUEST", "invalid template request")
	}
	req.Subject = strings.TrimSpace(req.Subject)
	req.Body = strings.TrimSpace(req.Body)
	if req.Subject == "" || req.Body == "" {
		return apiErr(c, http.StatusBadRequest, "INVALID_TEMPLATE_REQUEST", "subject and body are required")
	}
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(c.Request().Context(), `UPDATE notification_templates SET subject=$1, body=$2 WHERE key=$3`, req.Subject, req.Body, key)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TEMPLATE_UPDATE_ERROR", "could not update notification template")
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TEMPLATE_UPDATE_ERROR", "could not verify notification template update")
	}
	if changed == 0 {
		return apiErr(c, http.StatusNotFound, "TEMPLATE_NOT_FOUND", "notification template not found")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "NOTIFICATION_TEMPLATE_UPDATE", "", map[string]interface{}{"key": key}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record template audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit notification template update")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"key": key, "subject": req.Subject, "body": req.Body})
}
