package app

import "github.com/labstack/echo/v4"

func (a *App) audit(c echo.Context, actorID, action, documentID string, metadata map[string]interface{}) {}

func (a *App) notifyUser(c echo.Context, userID, templateKey, message string, documentID string) {}
