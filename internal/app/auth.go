package app

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return apiErr(c, http.StatusBadRequest, "INVALID_LOGIN_REQUEST", "username and password are required")
	}
	var u User
	err := a.db.GetContext(c.Request().Context(), &u, `SELECT id, username, display_name, role, password_hash, failed_attempts, locked_until FROM users WHERE username=$1`, req.Username)
	if err == sql.ErrNoRows {
		return apiErr(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
	}
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "DB_READ_ERROR", "could not read user")
	}
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return apiErr(c, http.StatusLocked, "ACCOUNT_LOCKED", "account is temporarily locked")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		a.db.ExecContext(c.Request().Context(), `UPDATE users SET failed_attempts=failed_attempts+1, locked_until=CASE WHEN failed_attempts+1 >= 5 THEN NOW()+INTERVAL '15 minutes' ELSE locked_until END WHERE id=$1`, u.ID)
		return apiErr(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
	}
	a.db.ExecContext(c.Request().Context(), `UPDATE users SET failed_attempts=0, locked_until=NULL WHERE id=$1`, u.ID)
	jti := makeIdentifier("jti")
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": u.ID, "username": u.Username, "role": u.Role, "jti": jti, "iat": now.Unix(), "exp": now.Add(a.cfg.SessionTTL).Unix()})
	signed, err := token.SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TOKEN_SIGN_ERROR", "could not sign token")
	}
	_, err = a.db.ExecContext(c.Request().Context(), `INSERT INTO sessions(jti,user_id,last_seen_at,expires_at,created_at) VALUES($1,$2,NOW(),NOW()+INTERVAL '8 hours',NOW())`, jti, u.ID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "SESSION_CREATE_ERROR", "could not create session")
	}
	a.audit(c, u.ID, "LOGIN", "", map[string]interface{}{"username": u.Username})
	return c.JSON(http.StatusOK, map[string]interface{}{"token": signed, "token_type": "Bearer", "expires_in_seconds": int(a.cfg.SessionTTL.Seconds()), "user": map[string]interface{}{"id": u.ID, "username": u.Username, "display_name": u.DisplayName, "role": u.Role}})
}

func (a *App) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			return apiErr(c, http.StatusUnauthorized, "AUTH_REQUIRED", "bearer token is required")
		}
		ts := c.Request().Header.Get("X-Request-Timestamp")
		if ts == "" {
			return apiErr(c, http.StatusBadRequest, "REQUEST_TIMESTAMP_REQUIRED", "X-Request-Timestamp is required")
		}
		parsed, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return apiErr(c, http.StatusBadRequest, "REQUEST_TIMESTAMP_INVALID", "timestamp must be RFC3339")
		}
		if time.Since(parsed) > a.cfg.RequestMaxAge || parsed.Sub(time.Now()) > a.cfg.RequestMaxAge {
			return apiErr(c, http.StatusUnauthorized, "REQUEST_EXPIRED", "request timestamp is outside the allowed window")
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(strings.TrimPrefix(raw, "Bearer "), claims, func(t *jwt.Token) (interface{}, error) { return []byte(a.cfg.JWTSecret), nil })
		if err != nil || !token.Valid {
			return apiErr(c, http.StatusUnauthorized, "TOKEN_INVALID", "token is invalid")
		}
		jti, _ := claims["jti"].(string)
		sub, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)
		username, _ := claims["username"].(string)
		if jti == "" || sub == "" {
			return apiErr(c, http.StatusUnauthorized, "TOKEN_INVALID", "token claims are incomplete")
		}
		var blacklisted int
		a.db.GetContext(c.Request().Context(), &blacklisted, `SELECT COUNT(*) FROM jwt_blacklist WHERE jti=$1`, jti)
		if blacklisted > 0 {
			return apiErr(c, http.StatusUnauthorized, "TOKEN_REVOKED", "token has been revoked")
		}
		reqID := currentRequestID(c)
		res, err := a.db.ExecContext(c.Request().Context(), `INSERT INTO request_replay_guard(request_id,jti,seen_at) VALUES($1,$2,NOW()) ON CONFLICT DO NOTHING`, reqID, jti)
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "REPLAY_GUARD_ERROR", "could not record request id")
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return apiErr(c, http.StatusConflict, "REPLAY_DETECTED", "request id was already used")
		}
		var active int
		a.db.GetContext(c.Request().Context(), &active, `SELECT COUNT(*) FROM sessions WHERE jti=$1 AND revoked_at IS NULL AND last_seen_at > NOW()-INTERVAL '8 hours'`, jti)
		if active == 0 {
			return apiErr(c, http.StatusUnauthorized, "SESSION_EXPIRED", "session is expired or revoked")
		}
		a.db.ExecContext(c.Request().Context(), `UPDATE sessions SET last_seen_at=NOW() WHERE jti=$1`, jti)
		c.Set("principal", Principal{UserID: sub, Username: username, Role: role, JTI: jti})
		return next(c)
	}
}

func (a *App) logout(c echo.Context) error {
	p := principal(c)
	a.db.ExecContext(c.Request().Context(), `INSERT INTO jwt_blacklist(jti,created_at) VALUES($1,NOW()) ON CONFLICT DO NOTHING`, p.JTI)
	a.db.ExecContext(c.Request().Context(), `UPDATE sessions SET revoked_at=NOW() WHERE jti=$1`, p.JTI)
	a.audit(c, p.UserID, "LOGOUT", "", map[string]interface{}{})
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "logged_out"})
}

func (a *App) me(c echo.Context) error {
	p := principal(c)
	return c.JSON(http.StatusOK, map[string]interface{}{"id": p.UserID, "username": p.Username, "role": p.Role})
}

func apiErr(c echo.Context, status int, code, msg string) error {
	return c.JSON(status, map[string]interface{}{"error": map[string]interface{}{"code": code, "message": msg, "details": map[string]interface{}{}, "request_id": currentRequestID(c), "timestamp": time.Now().UTC().Format(time.RFC3339)}})
}
