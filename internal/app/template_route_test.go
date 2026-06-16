package app

import (
	"os"
	"strings"
	"testing"
)

func TestNotificationTemplatePatchRouteIsRegistered(t *testing.T) {
	raw, err := os.ReadFile("server.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	want := `admin.PATCH("/notification-templates/:key", a.patchNotificationTemplate)`
	if !strings.Contains(src, want) {
		t.Fatalf("notification template patch route must remain registered")
	}
}
