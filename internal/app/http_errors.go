package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func apiHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "internal server error"
	if httpError, ok := err.(*echo.HTTPError); ok {
		status = httpError.Code
		switch status {
		case http.StatusBadRequest:
			code, message = "BAD_REQUEST", "bad request"
		case http.StatusUnauthorized:
			code, message = "UNAUTHORIZED", "authentication is required"
		case http.StatusForbidden:
			code, message = "FORBIDDEN", "access is denied"
		case http.StatusNotFound:
			code, message = "NOT_FOUND", "resource was not found"
		case http.StatusMethodNotAllowed:
			code, message = "METHOD_NOT_ALLOWED", "method is not allowed"
		}
	}
	_ = apiErr(c, status, code, message)
}
