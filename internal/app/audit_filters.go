package app

import (
    "fmt"
    "net/http"
    "strings"

    "github.com/labstack/echo/v4"
)

func (a *App) auditLogsFiltered(c echo.Context) error {
    page, size := parsePage(c, a.cfg)
    clauses := []string{"1=1"}
    args := []interface{}{}
    add := func(expr string, value string) {
        if strings.TrimSpace(value) == "" { return }
        args = append(args, value)
        clauses = append(clauses, fmt.Sprintf(expr, len(args)))
    }
    add("actor_user_id=$%d", c.QueryParam("actor_user_id"))
    add("document_id=$%d", c.QueryParam("document_id"))
    add("action_type=$%d", c.QueryParam("action_type"))
    add("request_id=$%d", c.QueryParam("request_id"))
    add("source_ip=$%d", c.QueryParam("source_ip"))
    if v := strings.TrimSpace(c.QueryParam("created_from")); v != "" {
        args = append(args, v)
        clauses = append(clauses, fmt.Sprintf("created_at >= $%d::timestamptz", len(args)))
    }
    if v := strings.TrimSpace(c.QueryParam("created_to")); v != "" {
        args = append(args, v)
        clauses = append(clauses, fmt.Sprintf("created_at <= $%d::timestamptz", len(args)))
    }
    args = append(args, size, (page-1)*size)
    query := fmt.Sprintf(`SELECT id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, strings.Join(clauses, " AND "), len(args)-1, len(args))
    rows := []map[string]interface{}{}
    if err := a.db.SelectContext(c.Request().Context(), &rows, query, args...); err != nil { return apiErr(c, http.StatusInternalServerError, "AUDIT_QUERY_ERROR", "could not list audit logs") }
    return c.JSON(http.StatusOK, map[string]interface{}{"data":rows,"page":page,"page_size":size,"filters":map[string]string{"actor_user_id":c.QueryParam("actor_user_id"),"document_id":c.QueryParam("document_id"),"action_type":c.QueryParam("action_type"),"request_id":c.QueryParam("request_id"),"source_ip":c.QueryParam("source_ip"),"created_from":c.QueryParam("created_from"),"created_to":c.QueryParam("created_to")}})
}
