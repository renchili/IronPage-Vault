package app

import "github.com/labstack/echo/v4"

func currentRequestID(c echo.Context) string {
    if v := c.Get("request_id"); v != nil {
        if s, ok := v.(string); ok && s != "" {
            return s
        }
    }
    if s := c.Request().Header.Get("X-Request-ID"); s != "" {
        return s
    }
    return "req_unknown"
}
