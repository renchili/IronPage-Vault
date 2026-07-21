package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	loginAttemptLimit  = 5
	loginAttemptWindow = 15 * time.Minute
	accountLockPeriod  = 15 * time.Minute
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) login(c echo.Context) error {
	ctx := c.Request().Context()
	var req loginRequest
	if err := c.Bind(&req); err != nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return apiErr(c, http.StatusBadRequest, "INVALID_LOGIN_REQUEST", "username and password are required")
	}

	var u User
	usernameKey := piiLookupKey(a.cfg.AESKey, req.Username)
	err := a.db.GetContext(ctx, &u, `SELECT id, username, username_ciphertext, display_name, display_name_ciphertext, role, password_hash, failed_attempts, locked_until FROM users WHERE username=$1 OR username=$2`, usernameKey, req.Username)
	if err == sql.ErrNoRows {
		return apiErr(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
	}
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "DB_READ_ERROR", "could not read user")
	}

	now := time.Now().UTC()
	if u.LockedUntil != nil && now.Before(*u.LockedUntil) {
		return apiErr(c, http.StatusLocked, "ACCOUNT_LOCKED", "account is temporarily locked")
	}

	storedHash, err := openPasswordHash(a.cfg.AESKey, u.PasswordHash)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "DB_READ_ERROR", "could not read user")
	}
	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)) != nil {
		locked, err := a.recordFailedLogin(c, u.ID, now)
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "LOGIN_ATTEMPT_WRITE_ERROR", "could not record failed login")
		}
		if locked {
			return apiErr(c, http.StatusLocked, "ACCOUNT_LOCKED", "account is temporarily locked")
		}
		return apiErr(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
	}

	if err := openUserPII(a.cfg.AESKey, &u); err != nil {
		return apiErr(c, http.StatusInternalServerError, "DB_READ_ERROR", "could not read user")
	}
	jti := makeIdentifier("jti")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      u.ID,
		"username": u.Username,
		"role":     u.Role,
		"jti":      jti,
		"iat":      now.Unix(),
		"exp":      now.Add(a.cfg.SessionTTL).Unix(),
	})
	signed, err := token.SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TOKEN_SIGN_ERROR", "could not sign token")
	}
	if err := a.completeSuccessfulLogin(c, u.ID, jti, now, u.Username); err != nil {
		return apiErr(c, http.StatusInternalServerError, "LOGIN_STATE_WRITE_ERROR", "could not persist successful login")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"token":              signed,
		"token_type":         "Bearer",
		"expires_in_seconds": int(a.cfg.SessionTTL.Seconds()),
		"user": map[string]interface{}{
			"id":           u.ID,
			"username":     u.Username,
			"display_name": u.DisplayName,
			"role":         u.Role,
		},
	})
}

func (a *App) recordFailedLogin(c echo.Context, userID string, attemptedAt time.Time) (bool, error) {
	ctx := c.Request().Context()
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, err
	}
	defer rollbackQuietly(tx)

	var lockedUserID string
	if err := tx.GetContext(ctx, &lockedUserID, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, userID); err != nil {
		return false, err
	}
	cutoff := attemptedAt.Add(-loginAttemptWindow)
	if _, err := tx.ExecContext(ctx, `DELETE FROM login_attempts WHERE user_id=$1 AND attempted_at < $2`, userID, cutoff); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO login_attempts(user_id,attempted_at) VALUES($1,$2)`, userID, attemptedAt); err != nil {
		return false, err
	}
	var attempts int
	if err := tx.GetContext(ctx, &attempts, `SELECT COUNT(*) FROM login_attempts WHERE user_id=$1 AND attempted_at >= $2`, userID, cutoff); err != nil {
		return false, err
	}
	locked := attempts >= loginAttemptLimit
	var lockedUntil interface{}
	if locked {
		lockedUntil = attemptedAt.Add(accountLockPeriod)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET failed_attempts=$2, locked_until=$3, updated_at=$4 WHERE id=$1`, userID, attempts, lockedUntil, attemptedAt); err != nil {
		return false, err
	}
	if err := a.auditWithExecutor(c, tx, userID, "LOGIN_FAILED", "", map[string]interface{}{"attempts_in_window": attempts, "locked": locked}); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return locked, nil
}

func (a *App) completeSuccessfulLogin(c echo.Context, userID, jti string, now time.Time, username string) error {
	ctx := c.Request().Context()
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackQuietly(tx)
	if _, err := tx.ExecContext(ctx, `DELETE FROM login_attempts WHERE user_id=$1`, userID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `UPDATE users SET failed_attempts=0, locked_until=NULL, updated_at=$2 WHERE id=$1`, userID, now)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to reset one user, reset %d", rows)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO sessions(jti,user_id,last_seen_at,expires_at,created_at) VALUES($1,$2,$3,$4,$3)`, jti, userID, now, now.Add(a.cfg.SessionTTL)); err != nil {
		return err
	}
	if err := a.auditWithExecutor(c, tx, userID, "LOGIN", "", map[string]interface{}{"username": username, "jti": jti}); err != nil {
		return err
	}
	return tx.Commit()
}

func rollbackQuietly(tx *sqlx.Tx) {
	_ = tx.Rollback()
}

func (a *App) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
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
		token, err := jwt.ParseWithClaims(strings.TrimPrefix(raw, "Bearer "), claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method %s", t.Method.Alg())
			}
			return []byte(a.cfg.JWTSecret), nil
		})
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
		if err := a.db.GetContext(ctx, &blacklisted, `SELECT COUNT(*) FROM jwt_blacklist WHERE jti=$1`, jti); err != nil {
			return apiErr(c, http.StatusInternalServerError, "AUTH_STATE_READ_ERROR", "could not read token state")
		}
		if blacklisted > 0 {
			return apiErr(c, http.StatusUnauthorized, "TOKEN_REVOKED", "token has been revoked")
		}
		reqID := currentRequestID(c)
		res, err := a.db.ExecContext(ctx, `INSERT INTO request_replay_guard(request_id,jti,seen_at) VALUES($1,$2,NOW()) ON CONFLICT DO NOTHING`, reqID, jti)
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "REPLAY_GUARD_ERROR", "could not record request id")
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "REPLAY_GUARD_ERROR", "could not verify request id")
		}
		if rows == 0 {
			return apiErr(c, http.StatusConflict, "REPLAY_DETECTED", "request id was already used")
		}
		now := time.Now().UTC()
		res, err = a.db.ExecContext(ctx, `UPDATE sessions SET last_seen_at=$2 WHERE jti=$1 AND revoked_at IS NULL AND last_seen_at > $3 AND (expires_at IS NULL OR expires_at > $2)`, jti, now, now.Add(-a.cfg.SessionTTL))
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "SESSION_UPDATE_ERROR", "could not update session activity")
		}
		rows, err = res.RowsAffected()
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "SESSION_UPDATE_ERROR", "could not verify session activity")
		}
		if rows == 0 {
			return apiErr(c, http.StatusUnauthorized, "SESSION_EXPIRED", "session is expired or revoked")
		}
		c.Set("principal", Principal{UserID: sub, Username: username, Role: role, JTI: jti})
		return next(c)
	}
}

func (a *App) logout(c echo.Context) error {
	ctx := c.Request().Context()
	p := principal(c)
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "LOGOUT_WRITE_ERROR", "could not start logout")
	}
	defer rollbackQuietly(tx)
	if _, err := tx.ExecContext(ctx, `INSERT INTO jwt_blacklist(jti,created_at) VALUES($1,NOW()) ON CONFLICT DO NOTHING`, p.JTI); err != nil {
		return apiErr(c, http.StatusInternalServerError, "LOGOUT_WRITE_ERROR", "could not revoke token")
	}
	res, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=NOW() WHERE jti=$1 AND revoked_at IS NULL`, p.JTI)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "LOGOUT_WRITE_ERROR", "could not revoke session")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "LOGOUT_WRITE_ERROR", "could not verify session revocation")
	}
	if rows != 1 {
		return apiErr(c, http.StatusInternalServerError, "LOGOUT_WRITE_ERROR", "session revocation did not affect one active session")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "LOGOUT", "", map[string]interface{}{"jti": p.JTI}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "LOGOUT_WRITE_ERROR", "could not record logout audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "LOGOUT_WRITE_ERROR", "could not complete logout")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "logged_out"})
}

func (a *App) me(c echo.Context) error {
	p := principal(c)
	return c.JSON(http.StatusOK, map[string]interface{}{"id": p.UserID, "username": p.Username, "role": p.Role})
}

func apiErr(c echo.Context, status int, code, msg string) error {
	return c.JSON(status, map[string]interface{}{"error": map[string]interface{}{"code": code, "message": msg, "details": map[string]interface{}{}, "request_id": currentRequestID(c), "timestamp": time.Now().UTC().Format(time.RFC3339)}})
}
