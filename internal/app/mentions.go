package app

import (
    "strings"
    "unicode"

    "github.com/labstack/echo/v4"
)

func extractMentionUsernames(text string) []string {
    found := map[string]bool{}
    out := []string{}
    fields := strings.Fields(text)
    for _, f := range fields {
        if !strings.HasPrefix(f, "@") || len(f) < 2 { continue }
        name := strings.TrimFunc(strings.TrimPrefix(f, "@"), func(r rune) bool {
            return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.')
        })
        if name == "" || found[name] { continue }
        found[name] = true
        out = append(out, name)
    }
    return out
}

func (a *App) notifyMentionedUsers(c echo.Context, comment string, documentID string, authorID string) {
    for _, username := range extractMentionUsernames(comment) {
        var user User
        if err := a.db.GetContext(c.Request().Context(), &user, `SELECT id,username,display_name,role,password_hash,failed_attempts,locked_until FROM users WHERE username=$1`, username); err != nil { continue }
        if user.ID == authorID { continue }
        a.notifyUser(c, user.ID, "annotation.mention", "You were mentioned in an annotation", documentID)
    }
}
