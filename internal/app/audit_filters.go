package app

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/repository"
)

func (a *App) auditLogsFiltered(c echo.Context) error {
	if principal(c).Role != RoleAdmin {
		return apiErr(c, http.StatusForbidden, "AUDIT_ADMIN_REQUIRED", "only admin may list audit logs")
	}
	page, size := parsePage(c, a.cfg)
	filters := repository.AuditLogFilters{
		ActorUserID: c.QueryParam("actor_user_id"),
		DocumentID:  c.QueryParam("document_id"),
		ActionType:  c.QueryParam("action_type"),
		RequestID:   c.QueryParam("request_id"),
		SourceIP:    c.QueryParam("source_ip"),
		CreatedFrom: c.QueryParam("created_from"),
		CreatedTo:   c.QueryParam("created_to"),
	}
	rows, err := repository.New(a.db).ListAuditLogs(c.Request().Context(), filters, size, (page-1)*size)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_QUERY_ERROR", "could not list audit logs")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":      rows,
		"page":      page,
		"page_size": size,
		"filters": map[string]string{
			"actor_user_id": c.QueryParam("actor_user_id"),
			"document_id":   c.QueryParam("document_id"),
			"action_type":   c.QueryParam("action_type"),
			"request_id":    c.QueryParam("request_id"),
			"source_ip":     c.QueryParam("source_ip"),
			"created_from":  c.QueryParam("created_from"),
			"created_to":    c.QueryParam("created_to"),
		},
	})
}
