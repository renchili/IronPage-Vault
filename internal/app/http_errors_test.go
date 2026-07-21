package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAPIHTTPErrorHandlerUsesErrorEnvelope(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	req.Header.Set("X-Request-ID", "req_contract")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set("request_id", "req_contract")

	apiHTTPErrorHandler(echo.NewHTTPError(http.StatusNotFound), ctx)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rec.Code)
	}
	var body struct {
		Error struct {
			Code      string `json:"code"`
			RequestID string `json:"request_id"`
			Timestamp string `json:"timestamp"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "NOT_FOUND" || body.Error.RequestID != "req_contract" || body.Error.Timestamp == "" {
		t.Fatalf("unexpected envelope: %+v", body.Error)
	}
}

func TestAPIHTTPErrorHandlerDoesNotLeakInternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set("request_id", "req_internal")

	apiHTTPErrorHandler(assertionError{}, ctx)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rec.Code)
	}
	if rec.Body.String() == "" || !contains(rec.Body.String(), "INTERNAL_ERROR") {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

type assertionError struct{}

func (assertionError) Error() string { return "database password must not leak" }

func contains(value, needle string) bool {
	for i := 0; i + len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
