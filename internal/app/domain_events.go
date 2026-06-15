package app

import "github.com/labstack/echo/v4"

func (a *App) audit(c echo.Context, actorID, action, documentID string, metadata map[string]interface{}) {
    _ = metadata
    if documentID == "" {
        _, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,$4,$5,'{}'::jsonb,NOW())`, makeIdentifier("aud"), actorID, action, currentRequestID(c), c.RealIP())
        return
    }
    _, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,$3,$4,$5,$6,'{}'::jsonb,NOW())`, makeIdentifier("aud"), actorID, documentID, action, currentRequestID(c), c.RealIP())
}

func (a *App) notifyUser(c echo.Context, userID, templateKey, message string, documentID string) {
    _ = a.createNotification(c, userID, documentID, templateKey, message)
}
