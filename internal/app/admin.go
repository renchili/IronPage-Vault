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
	ID          string          `db:"id" json:"id"`
	ActorUserID *string         `db:"actor_user_id" json:"actor_user_id,omitempty"`
	DocumentID  *string         `db:"document_id" json:"document_id,omitempty"`
	ActionType  string          `db:"action_type" json:"action_type"`
	RequestID   string          `db:"request_id" json:"request_id"`
	SourceIP    string          `db:"source_ip" json:"source_ip"`
	Metadata    json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

type notificationResponse struct {
	ID          string     `db:"id" json:"id"`
	DocumentID  *string    `db:"document_id" json:"document_id,omitempty"`
	TemplateKey string     `db:"template_key" json:"template_key"`
	Message     string     `db:"message" json:"message"`
	ReadAt      *time.Time `db:"read_at" json:"read_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
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
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "PASSWORD_HASH_ERROR", "could not hash password")
	}
	id := makeIdentifier("usr")
	_, err = a.db.ExecContext(c.Request().Context(), `INSERT INTO users(id,username,display_name,role,password_hash,created_at) VALUES($1,$2,$3,$4,$5,NOW())`, id, req.Username, req.DisplayName, req.Role, string(hash))
	if err != nil {
		return apiErr(c, http.StatusConflict, "USER_CREATE_ERROR", "could not create user")
	}
	_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,'USER_CREATE',$3,$4,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, currentRequestID(c), c.RealIP())
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "username": req.Username, "role": req.Role})
}

func (a *App) listUsers(c echo.Context) error {
	rows := []User{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,username,display_name,role,password_hash,failed_attempts,locked_until FROM users ORDER BY username`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "USER_QUERY_ERROR", "could not list users")
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
	_, err := a.db.ExecContext(c.Request().Context(), `INSERT INTO config_entries(key,value,updated_by,updated_at) VALUES($1,$2,$3,NOW()) ON CONFLICT(key) DO UPDATE SET value=excluded.value,updated_by=excluded.updated_by,updated_at=NOW()`, c.Param("key"), req.Value, p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "CONFIG_UPDATE_ERROR", "could not update config")
	}
	_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,'CONFIG_UPDATE',$3,$4,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, currentRequestID(c), c.RealIP())
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

func (a *App) auditLogs(c echo.Context) error {
	page, size := parsePage(c, a.cfg)
	rows := []auditLogResponse{}
	actionType := c.QueryParam("action_type")
	if actionType != "" {
		if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at FROM audit_logs WHERE action_type=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, actionType, size, (page-1)*size); err != nil {
			return apiErr(c, http.StatusInternalServerError, "AUDIT_QUERY_ERROR", "could not list audit logs")
		}
	} else {
		if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,actor_user_id,document_id,action_type,request_id,source_ip,metadata,created_at FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, size, (page-1)*size); err != nil {
			return apiErr(c, http.StatusInternalServerError, "AUDIT_QUERY_ERROR", "could not list audit logs")
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows, "page": page, "page_size": size})
}

func (a *App) notifications(c echo.Context) error {
	p := principal(c)
	rows := []notificationResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,document_id,template_key,message,read_at,created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100`, p.UserID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_QUERY_ERROR", "could not list notifications")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

func (a *App) readNotification(c echo.Context) error {
	p := principal(c)
	_, err := a.db.ExecContext(c.Request().Context(), `UPDATE notifications SET read_at=NOW() WHERE id=$1 AND user_id=$2`, c.Param("id"), p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "NOTIFICATION_UPDATE_ERROR", "could not mark read")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "read"})
}

func (a *App) runBackup(c echo.Context) error {
	p := principal(c)
	id := makeIdentifier("bak")
	_, err := a.db.ExecContext(c.Request().Context(), `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'logical_dump','Queued',$2,$3,NOW())`, id, a.cfg.BackupDir+"/"+id+".sql", p.UserID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "could not create backup job")
	}
	_, _ = a.db.ExecContext(c.Request().Context(), `INSERT INTO audit_logs(id,actor_user_id,action_type,request_id,source_ip,metadata,created_at) VALUES($1,$2,'BACKUP_CREATE',$3,$4,'{}'::jsonb,NOW())`, makeIdentifier("aud"), p.UserID, currentRequestID(c), c.RealIP())
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id, "status": "Queued", "created_at": time.Now().UTC()})
}

func (a *App) backupJobs(c echo.Context) error {
	rows := []backupJobResponse{}
	if err := a.db.SelectContext(c.Request().Context(), &rows, `SELECT id,kind,status,target_path,created_by,created_at FROM backup_jobs ORDER BY created_at DESC`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "BACKUP_QUERY_ERROR", "could not list backups")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}
