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

func TestBackupRestoreAndResolutionUseExclusiveOperationPaths(t *testing.T) {
	for _, path := range []string{
		"/api/admin/backup/run",
		"/api/admin/backup/restore",
		"/api/admin/backup/restore/rst_example/resolve",
	} {
		if !isExclusiveOperationPath(path) {
			t.Fatalf("path %s must acquire the exclusive operation barrier", path)
		}
	}
	if isExclusiveOperationPath("/api/documents") {
		t.Fatal("ordinary document mutation must use the shared mutation barrier")
	}
	if isRestoreResolutionPath("/api/admin/backup/restore/resolve") {
		t.Fatal("restore resolution path must contain a restore id")
	}
}

func TestRestoreRequestClassificationIsExact(t *testing.T) {
	e := echo.New()
	for _, test := range []struct {
		method string
		path   string
		want   bool
	}{
		{http.MethodPost, "/api/admin/backup/restore", true},
		{http.MethodGet, "/api/admin/backup/restore", false},
		{http.MethodPost, "/api/admin/backup/restore/rst_1/resolve", false},
		{http.MethodPost, "/api/admin/backup/run", false},
	} {
		request := httptest.NewRequest(test.method, test.path, nil)
		context := e.NewContext(request, httptest.NewRecorder())
		if got := isRestoreOperationRequest(context); got != test.want {
			t.Fatalf("isRestoreOperationRequest(%s %s) = %v, want %v", test.method, test.path, got, test.want)
		}
	}
}

func TestMaintenanceRejectsOrdinaryAndConcurrentRestoreRequests(t *testing.T) {
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
		t.Fatalf("concurrent restore response: %v", err)
	}
	if called || restoreRecorder.Code != http.StatusConflict {
		t.Fatalf("concurrent restore was not rejected: called=%v status=%d", called, restoreRecorder.Code)
	}
}

func TestRestoreAdmissionRejectsSecondAuthenticationPath(t *testing.T) {
	e := echo.New()
	operations := &operationCoordinator{}
	operations.restoreAdmission.Lock()
	defer operations.restoreAdmission.Unlock()
	a := &App{operations: operations}

	request := httptest.NewRequest(http.MethodPost, "/api/admin/backup/restore", nil)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	called := false
	if err := a.maintenanceMiddleware(func(c echo.Context) error {
		called = true
		return nil
	})(context); err != nil {
		t.Fatalf("restore admission response: %v", err)
	}
	if called || recorder.Code != http.StatusConflict {
		t.Fatalf("second restore entered authentication: called=%v status=%d", called, recorder.Code)
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
