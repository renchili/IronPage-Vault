package app

import (
    "os"
    "strings"
    "testing"
)

func TestAuditLogsRouteRequiresAdminRole(t *testing.T) {
    raw, err := os.ReadFile("server.go")
    if err != nil { t.Fatal(err) }
    src := string(raw)
    want := `api.GET("/audit-logs", a.auditLogsFiltered, requireRole(RoleAdmin))`
    if !strings.Contains(src, want) {
        t.Fatalf("audit logs route must require admin role")
    }
}
