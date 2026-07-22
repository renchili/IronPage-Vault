package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestMutationBarrierClassifiesUnsafeMethods(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if !requiresMutationBarrier(method) {
			t.Fatalf("method %s must participate in the mutation barrier", method)
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		if requiresMutationBarrier(method) {
			t.Fatalf("method %s must remain read-only", method)
		}
	}
}

func TestBackupAndRestoreUseExclusiveOperationPaths(t *testing.T) {
	for _, path := range []string{"/api/admin/backup/run", "/api/admin/backup/restore"} {
		if !isExclusiveOperationPath(path) {
			t.Fatalf("path %s must acquire the exclusive operation barrier", path)
		}
	}
	if isExclusiveOperationPath("/api/documents") {
		t.Fatal("ordinary document mutation must use the shared mutation barrier")
	}
}

func TestMaintenanceRejectsOrdinaryRequestsButAllowsRestoreOwner(t *testing.T) {
	e := echo.New()
	operations := &operationCoordinator{}
	operations.maintenance.Store(true)
	a := &App{operations: operations}

	request := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	called := false
	if err := a.maintenanceMiddleware(func(c echo.Context) error {
		called = true
		return nil
	})(context); err != nil {
		t.Fatalf("maintenance response: %v", err)
	}
	if called || recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("ordinary request was not rejected during maintenance: called=%v status=%d", called, recorder.Code)
	}

	restoreRequest := httptest.NewRequest(http.MethodPost, "/api/admin/backup/restore", nil)
	restoreRecorder := httptest.NewRecorder()
	restoreContext := e.NewContext(restoreRequest, restoreRecorder)
	called = false
	if err := a.maintenanceMiddleware(func(c echo.Context) error {
		called = true
		return nil
	})(restoreContext); err != nil {
		t.Fatalf("restore maintenance owner: %v", err)
	}
	if !called {
		t.Fatal("restore owner request must enter the handler that owns maintenance")
	}
}

func TestRequestedRestoreBecomesInterruptedNotFailed(t *testing.T) {
	record := interruptedRestoreRecord(restoreLifecycleRecord{
		ID:                "rst_test",
		Status:            restoreStatusRequested,
		ActorUserID:       "usr_admin",
		RequestID:         "req_test",
		RequestedMetadata: map[string]interface{}{"restore_id": "rst_test"},
	})
	if record.Status != restoreStatusInterrupted || record.Status == restoreStatusFailed {
		t.Fatalf("requested restore reconciled to %q", record.Status)
	}
	if record.Action != "BACKUP_RESTORE_INTERRUPTED" || record.OutcomeActorUserID != systemPrincipalID {
		t.Fatalf("interrupted restore audit attribution is incomplete: %#v", record)
	}
	if record.Metadata["outcome"] != "unknown" {
		t.Fatalf("interrupted restore must preserve unknown outcome: %#v", record.Metadata)
	}
}
