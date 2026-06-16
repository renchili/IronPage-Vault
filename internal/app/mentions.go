package app

import (
	"github.com/labstack/echo/v4"

	"ironpage-vault/internal/core"
)

func extractMentionUsernames(text string) []string {
	return core.ExtractMentionUsernames(text)
}

func (a *App) notifyMentionedUsers(c echo.Context, comment string, documentID string, authorID string) {
	for _, username := range extractMentionUsernames(comment) {
		var user User
		if err := a.db.GetContext(c.Request().Context(), &user, `SELECT id,username,display_name,role,password_hash,failed_attempts,locked_until FROM users WHERE username=$1`, username); err != nil {
			continue
		}
		if user.ID == authorID {
			continue
		}
		a.notifyUser(c, user.ID, "annotation.mention", "You were mentioned in an annotation", documentID)
	}
}
