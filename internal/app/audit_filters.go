package app

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/repository"
)

func validateAuditTimeFilter(value string) bool {
	if strings.TrimSpace(value) == "" {
		return true
	}
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func (a *App) auditLogsFiltered(c echo.Context) error {
	if principal(c).Role != RoleAdmin {
		return apiErr(c, http.StatusForbidden, "AUDIT_ADMIN_REQUIRED", "only admin may list audit logs")
	}
	if !validateAuditTimeFilter(c.QueryParam("created_from")) || !validateAuditTimeFilter(c.QueryParam("created_to")) {
		return apiErr(c, http.StatusBadRequest, "INVALID_AUDIT_TIME_FILTER", "created_from and created_to must be RFC3339 timestamps")
	}
	page, size, err := a.configuredPage(c)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "PAGINATION_CONFIG_ERROR", "could not read pagination configuration")
	}
	sourceIP := strings.TrimSpace(c.QueryParam("source_ip"))
	filters := repository.AuditLogFilters{
		ActorUserID: c.QueryParam("actor_user_id"),
		DocumentID:  c.QueryParam("document_id"),
		ActionType:  c.QueryParam("action_type"),
		RequestID:   c.QueryParam("request_id"),
		CreatedFrom: c.QueryParam("created_from"),
		CreatedTo:   c.QueryParam("created_to"),
	}
	if sourceIP != "" {
		filters.SourceIPLookup = piiLookupKey(a.cfg.AESKey, sourceIP)
	}
	rows, err := repository.New(a.db).ListAuditLogs(c.Request().Context(), filters, size, (page-1)*size)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_QUERY_ERROR", "could not list audit logs")
	}
	data := make([]auditLogResponse, 0, len(rows))
	for _, row := range rows {
		response := auditLogResponse{
			ID:                 row.ID,
			ActorUserID:        row.ActorUserID,
			DocumentID:         row.DocumentID,
			ActionType:         row.ActionType,
			RequestID:          row.RequestID,
			SourceIP:           row.SourceIP,
			SourceIPCiphertext: row.SourceIPCiphertext,
			Metadata:           row.Metadata,
			MetadataCiphertext: row.MetadataCiphertext,
			CreatedAt:          row.CreatedAt,
		}
		if err := openAuditPII(a.cfg.AESKey, &response); err != nil {
			return apiErr(c, http.StatusInternalServerError, "AUDIT_QUERY_ERROR", "could not read audit log")
		}
		data = append(data, response)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":      data,
		"page":      page,
		"page_size": size,
		"filters": map[string]string{
			"actor_user_id": c.QueryParam("actor_user_id"),
			"document_id":   c.QueryParam("document_id"),
			"action_type":   c.QueryParam("action_type"),
			"request_id":    c.QueryParam("request_id"),
			"source_ip":     sourceIP,
			"created_from":  c.QueryParam("created_from"),
			"created_to":    c.QueryParam("created_to"),
		},
	})
}
