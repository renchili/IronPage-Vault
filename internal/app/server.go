package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "ironpage-vault/docs/swagger"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type App struct {
	cfg        Config
	db         *sqlx.DB
	operations *operationCoordinator
}

func MustRun(cfg Config) {
	if err := Run(cfg); err != nil {
		log.Fatalf("ironpage startup failed: %v", err)
	}
}

func Run(cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid runtime configuration: %w", err)
	}
	if err := os.MkdirAll(cfg.StorageDir, 0750); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.BackupDir, 0750); err != nil {
		return err
	}
	db, err := OpenDatabase(cfg)
	if err != nil {
		return err
	}
	if err := RunMigrations(db, cfg.MigrationsDir); err != nil {
		return err
	}
	if err := EnsureAuditSourceIPLookups(context.Background(), db, cfg.AESKey); err != nil {
		return err
	}
	if err := EnsureRuntimeConfiguration(context.Background(), db, cfg); err != nil {
		return err
	}
	if err := EnsureInitialUsers(context.Background(), db, cfg); err != nil {
		return err
	}
	if err := EnsureSystemPrincipal(context.Background(), db, cfg); err != nil {
		return fmt.Errorf("system principal initialization failed: %w", err)
	}
	a := &App{cfg: cfg, db: db, operations: newOperationCoordinator(db)}
	if err := a.reconcileRestoreLifecycle(context.Background()); err != nil {
		return fmt.Errorf("restore lifecycle reconciliation failed: %w", err)
	}
	a.startBackupScheduler()
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = apiHTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(a.requestIDMiddleware)
	e.Use(a.maintenanceMiddleware)
	e.Use(a.mutationBarrierMiddleware)
	e.GET("/healthz", a.health)
	if cfg.AcceptanceMode {
		e.Static("/ui", cfg.PublicDir)
	}
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.POST("/api/auth/login", a.login)

	api := e.Group("/api", a.authMiddleware)
	api.POST("/auth/logout", a.logout)
	api.GET("/auth/me", a.me)

	admin := api.Group("/admin", requireRole(RoleAdmin))
	admin.POST("/users", a.createUser)
	admin.GET("/users", a.listUsers)
	admin.GET("/config", a.listConfig)
	admin.PATCH("/config/:key", a.patchConfig)
	admin.GET("/workflow-statuses", a.workflowStatuses)
	admin.PUT("/workflow-statuses", a.replaceWorkflowStatuses)
	admin.GET("/notification-templates", a.notificationTemplates)
	admin.PATCH("/notification-templates/:key", a.patchNotificationTemplate)
	admin.POST("/backup/run", a.runBackupMetadataSnapshot)
	admin.GET("/backup/jobs", a.backupJobs)
	admin.POST("/backup/restore", a.restoreBackup, a.restoreMaintenanceMiddleware)
	admin.POST("/backup/restore/:id/resolve", a.resolveInterruptedRestore, a.exclusiveOperationMiddleware)

	docs := api.Group("/documents")
	docs.GET("", a.listDocuments)
	docs.POST("", a.uploadDocument, requireRole(RoleEditor))
	docs.POST("/batch", a.batchUploadDocuments, requireRole(RoleEditor))
	docs.POST("/compare", a.compareVersions)
	docs.GET("/:id", a.getDocument)
	docs.GET("/:id/file", a.downloadDocument)
	docs.GET("/:id/versions", a.listVersions)
	docs.POST("/:id/rollback", a.rollbackVersion, requireRole(RoleEditor))
	docs.POST("/:id/finalize", a.finalizeDocument, requireRole(RoleEditor))
	docs.POST("/:id/workflow/transition", a.transitionDocument)
	docs.POST("/:id/redactions", a.proposeRedaction, requireRole(RoleEditor))
	docs.GET("/:id/redactions", a.listRedactions)
	docs.POST("/:id/redactions/:redaction_id/confirm", a.confirmRedaction, requireRole(RoleEditor))
	docs.POST("/:id/annotations", a.createAnnotation, requireRole(RoleReviewer))
	docs.GET("/:id/annotations", a.listAnnotations)
	docs.POST("/:id/bates", a.applyBatesVersion, requireRole(RoleEditor))

	api.PATCH("/annotations/:id/disposition", a.updateAnnotationDisposition, requireRole(RoleReviewer))
	api.GET("/audit-logs", a.auditLogsFiltered, requireRole(RoleAdmin))
	api.GET("/notifications", a.notifications)
	api.POST("/notifications/:id/read", a.readNotificationChecked)
	return e.Start(cfg.HTTPAddr)
}

func (a *App) health(c echo.Context) error {
	if err := a.db.PingContext(c.Request().Context()); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{"status": "db_unavailable"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok", "time": time.Now().UTC()})
}

func (a *App) requestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Request().Header.Get("X-Request-ID")
		if id == "" {
			id = makeIdentifier("req")
		}
		c.Set("request_id", id)
		c.Response().Header().Set("X-Request-ID", id)
		return next(c)
	}
}

func requireRole(roles ...string) echo.MiddlewareFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			p, ok := c.Get("principal").(Principal)
			if !ok || !allowed[p.Role] {
				return c.JSON(http.StatusForbidden, map[string]interface{}{"error": map[string]interface{}{"code": "FORBIDDEN", "message": "role is not allowed", "details": map[string]interface{}{}, "request_id": currentRequestID(c), "timestamp": time.Now().UTC().Format(time.RFC3339)}})
			}
			return next(c)
		}
	}
}

func principal(c echo.Context) Principal {
	p, _ := c.Get("principal").(Principal)
	return p
}

func atoiDefault(s string, d int) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return d
	}
	return n
}
