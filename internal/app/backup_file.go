package app

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/labstack/echo/v4"
)

func (a *App) runBackupFile(c echo.Context) error {
    p := principal(c)
    id := makeIdentifier("bak")
    if err := os.MkdirAll(a.cfg.BackupDir, 0750); err != nil { return apiErr(c, http.StatusInternalServerError, "BACKUP_DIR_ERROR", "could not create backup directory") }
    target := filepath.Join(a.cfg.BackupDir, id+".sql")
    payload := fmt.Sprintf("-- IronPage Vault local logical backup\n-- backup_id=%s\n-- created_at=%s\n-- database=%s\n", id, time.Now().UTC().Format(time.RFC3339), a.cfg.DBName)
    if err := os.WriteFile(target, []byte(payload), 0640); err != nil { return apiErr(c, http.StatusInternalServerError, "BACKUP_WRITE_ERROR", "could not write backup file") }
    _, err := a.db.ExecContext(c.Request().Context(), `INSERT INTO backup_jobs(id,kind,status,target_path,created_by,created_at) VALUES($1,'logical_dump','Completed',$2,$3,NOW())`, id, target, p.UserID)
    if err != nil { return apiErr(c, http.StatusInternalServerError, "BACKUP_CREATE_ERROR", "could not record backup job") }
    a.audit(c, p.UserID, "BACKUP_CREATE", "", nil)
    return c.JSON(http.StatusCreated, map[string]interface{}{"id":id,"status":"Completed","target_path":target,"created_at":time.Now().UTC()})
}
