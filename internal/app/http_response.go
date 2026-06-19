package app

import (
	"github.com/labstack/echo/v4"
)

type Payload struct {
	Data interface{} `json:"data,omitempty"`
}

func sendJSON(c echo.Context, code int, value interface{}) error {
	return c.JSON(code, Payload{Data: value})
}
