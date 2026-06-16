package app

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/store"
)

func (a *App) auditLogsFiltered(c echo.Context) error {
	page, size := parsePage(c, a.cfg)
	filters := store.AuditLogFilters{
		ActorUserID: c.QueryParam("actor_user_id"),
		DocumentID:  c.QueryParam("document_id"),
		ActionType:  c.QueryParam("action_type"),
		RequestID:   c.QueryParam("request_id"),
		SourceIP:    c.QueryParam("source_ip"),
		CreatedFrom: c.QueryParam("created_from"),
		CreatedTo:   c.QueryParam("created_to"),
	}
	query, args := store.BuildAuditLogListQuery(filters, size, (page-1)*size)
	rows := []map[string]interface{}{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, query, args...); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_QUERY_ERROR", "could not list audit logs")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows, "page": page, "page_size": size, "filters": map[string]string{"actor_user_id": c.QueryParam("actor_user_id"), "document_id": c.QueryParam("document_id"), "action_type": c.QueryParam("action_type"), "request_id": c.QueryParam("request_id"), "source_ip": c.QueryParam("source_ip"), "created_from": c.QueryParam("created_from"), "created_to": c.QueryParam("created_to")}})
}
