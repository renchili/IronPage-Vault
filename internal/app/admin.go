package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type configEntryResponse struct {
	Key       string    `db:"key" json:"key"`
	Value     string    `db:"value" json:"value"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type workflowStatusResponse struct {
	Name     string `db:"name" json:"name"`
	Position int    `db:"position" json:"position"`
	Mutable  bool   `db:"mutable" json:"mutable"`
}

type notificationTemplateResponse struct {
	ID      string `db:"id" json:"id"`
	Key     string `db:"key" json:"key"`
	Subject string `db:"subject" json:"subject"`
	Body    string `db:"body" json:"body"`
}

type auditLogResponse struct {
	ID                 string          `db:"id" json:"id"`
	ActorUserID        *string         `db:"actor_user_id" json:"actor_user_id,omitempty"`
	DocumentID         *string         `db:"document_id" json:"document_id,omitempty"`
	ActionType         string          `db:"action_type" json:"action_type"`
	RequestID          string          `db:"request_id" json:"request_id"`
	SourceIP           string          `db:"source_ip" json:"source_ip"`
	SourceIPCiphertext string          `db:"source_ip_ciphertext" json:"-"`
	Metadata           json.RawMessage `db:"metadata" json:"metadata"`
	MetadataCiphertext string          `db:"metadata_ciphertext" json:"-"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
}

type notificationResponse struct {
	ID                string     `db:"id" json:"id"`
	DocumentID        *string    `db:"document_id" json:"document_id,omitempty"`
	TemplateKey       string     `db:"template_key" json:"template_key"`
	Message           string     `db:"message" json:"message"`
	MessageCiphertext string     `db:"message_ciphertext" json:"-"`
	ReadAt            *time.Time `db:"read_at" json:"read_at,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
}

type backupJobResponse struct {
	ID         string    `db:"id" json:"id"`
	Kind       string    `db:"kind" json:"kind"`
	Status     string    `db:"status" json:"status"`
	TargetPath string    `db:"target_path" json:"target_path"`
	CreatedBy  *string   `db:"created_by" json:"created_by,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

func (a *App) createUser(c echo.Context) error {
	p := principal(c)
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Password    string `json:"password"`
	}
	if err := c.Bind(&req); err != nil || req.Username == "" || req.Password == "" || req.Role == "" {
		return apiErr(c, http.StatusBadRequest, "INVALID_USER_REQUEST", "username, role and password are required")
	}
	if req.Role != RoleAdmin && req.Role != RoleEditor && req.Role != RoleReviewer {
		return apiErr(c, http.StatusBadRequest, "INVALID_ROLE", "role must be Admin, Editor, or Reviewer")
	}
	if !IsValidUserSecret(req.Password) {
		return apiErr(c, http.StatusBadRequest, "WEAK_PASSWORD", "password must be at least 8 characters and include a digit and special character")
	}
	usernameKey := piiLookupKey(a.cfg.AESKey, req.Username)
	var existing int
	if err := a.db.GetContext(c.Request().Context(), &existing, `SELECT COUNT(*) FROM users WHERE username=$1 OR username=$2`, usernameKey, req.Username); err != nil {
		return apiErr(c, http.StatusInternalServerError, "USER_QUERY_ERROR", "could not check user")
	}
	if existing > 0 {
		return apiErr(c, http.StatusConflict, "USER_CREATE_ERROR", "could not create user")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "PASSWORD_HASH_ERROR", "could not hash password")
	}
	storedHash, err := sealPasswordHash(a.cfg.AESKey, hash)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "PASSWORD_HASH_ERROR", "could not protect password hash")
	}
	usernameCipher, err := sealPII(a.cfg.AESKey, req.Username)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt username")
	}
	displayCipher, err := sealPII(a.cfg.AESKey, req.DisplayName)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "ENCRYPTION_ERROR", "could not encrypt display name")
	}
	id := makeIdentifier("usr")
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO users(id,username,username_ciphertext,display_name,display_name_ciphertext,role,password_hash,created_at) VALUES($1,$2,$3,'',$4,$5,$6,NOW())`, id, usernameKey, usernameCipher, displayCipher, req.Role, storedHash); err != nil {
		return apiErr(c, http.StatusConflict, "USER_CREATE_ERROR", "could not create user")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "USER_CREATE", "", map[string]interface{}{"created_user_id": id, "role": req.Role}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record user creation audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit user creation")
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "username": req.Username, "display_name": req.DisplayName, "role": req.Role})
}

func (a *App) listUsers(c echo.Context) error {
	rows := []User{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,username,username_ciphertext,display_name,display_name_ciphertext,role,password_hash,failed_attempts,locked_until FROM users ORDER BY created_at`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "USER_QUERY_ERROR", "could not list users")
	}
	for i := range rows {
		if err := openUserPII(a.cfg.AESKey, &rows[i]); err != nil {
			return apiErr(c, http.StatusInternalServerError, "USER_QUERY_ERROR", "could not read user")
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) listConfig(c echo.Context) error {
	rows := []configEntryResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT key,value,updated_at FROM config_entries ORDER BY key`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "CONFIG_QUERY_ERROR", "could not list config")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) patchConfig(c echo.Context) error {
	p := principal(c)
	var req struct {
		Value string `json:"value"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_CONFIG_REQUEST", "value is required")
	}
	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO config_entries(key,value,updated_by,updated_at) VALUES($1,$2,$3,NOW()) ON CONFLICT(key) DO UPDATE SET value=excluded.value,updated_by=excluded.updated_by,updated_at=NOW()`, c.Param("key"), req.Value, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "CONFIG_UPDATE_ERROR", "could not update config")
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "CONFIG_UPDATE", "", map[string]interface{}{"key": c.Param("key")}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record configuration audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit configuration update")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"key": c.Param("key"), "value": req.Value})
}

func (a *App) workflowStatuses(c echo.Context) error {
	rows := []workflowStatusResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT name,position,mutable FROM workflow_status_definitions ORDER BY position`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_QUERY_ERROR", "could not list statuses")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) notificationTemplates(c echo.Context) error {
	rows := []notificationTemplateResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,key,subject,body FROM notification_templates ORDER BY key`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "TEMPLATE_QUERY_ERROR", "could not list templates")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) notifications(c echo.Context) error {
	p := principal(c)
	rows := []notificationResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,template_key,message,message_ciphertext,read_at,created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100`, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_QUERY_ERROR", "could not list notifications")
	}
	for i := range rows {
		if err := openNotificationPII(a.cfg.AESKey, &rows[i]); err != nil {
			return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_QUERY_ERROR", "could not read notification")
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) backupJobs(c echo.Context) error {
	rows := []backupJobResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,kind,status,target_path,created_by,created_at FROM backup_jobs ORDER BY created_at DESC`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_QUERY_ERROR", "could not list backups")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}
