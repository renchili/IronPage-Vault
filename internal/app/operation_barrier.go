package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

const (
	operationAdvisoryLockID     int64 = 52833707454590
	maintenanceOwnerContextKey       = "restore_maintenance_owner"
)

var errMaintenanceActive = errors.New("maintenance operation is active")

type operationCoordinator struct {
	db              *sqlx.DB
	maintenanceGate sync.RWMutex
	maintenance     atomic.Bool
}

func newOperationCoordinator(db *sqlx.DB) *operationCoordinator {
	return &operationCoordinator{db: db}
}

func (o *operationCoordinator) withAdvisoryLock(ctx context.Context, lockSQL, unlockSQL string, fn func() error) error {
	if o == nil || o.db == nil {
		return fmt.Errorf("operation coordinator database is unavailable")
	}
	conn, err := o.db.Connx(ctx)
	if err != nil {
		return err
	}
	locked := false
	defer func() {
		if locked {
			_ = conn.Close()
		}
	}()
	if _, err := conn.ExecContext(ctx, lockSQL, operationAdvisoryLockID); err != nil {
		_ = conn.Close()
		return err
	}
	locked = true
	operationErr := fn()
	releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, unlockErr := conn.ExecContext(releaseCtx, unlockSQL, operationAdvisoryLockID)
	closeErr := conn.Close()
	locked = false
	return errors.Join(operationErr, unlockErr, closeErr)
}

func (o *operationCoordinator) withSharedMutation(ctx context.Context, fn func() error) error {
	return o.withAdvisoryLock(ctx, `SELECT pg_advisory_lock_shared($1)`, `SELECT pg_advisory_unlock_shared($1)`, fn)
}

func (o *operationCoordinator) withExclusiveOperation(ctx context.Context, fn func() error) error {
	return o.withAdvisoryLock(ctx, `SELECT pg_advisory_lock($1)`, `SELECT pg_advisory_unlock($1)`, fn)
}

func (o *operationCoordinator) withMaintenanceOperation(ctx context.Context, fn func() error) error {
	if o == nil {
		return fmt.Errorf("operation coordinator is unavailable")
	}
	if !o.maintenance.CompareAndSwap(false, true) {
		return errMaintenanceActive
	}
	defer o.maintenance.Store(false)
	o.maintenanceGate.Lock()
	defer o.maintenanceGate.Unlock()
	return o.withExclusiveOperation(ctx, fn)
}

func requiresMutationBarrier(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isRestoreOperationRequest(c echo.Context) bool {
	return c.Request().Method == http.MethodPost && c.Request().URL.Path == "/api/admin/backup/restore"
}

func isExclusiveOperationPath(path string) bool {
	return path == "/api/admin/backup/run" || path == "/api/admin/backup/restore"
}

func (a *App) maintenanceMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.operations == nil {
			return next(c)
		}
		if isRestoreOperationRequest(c) {
			err := a.operations.withMaintenanceOperation(c.Request().Context(), func() error {
				c.Set(maintenanceOwnerContextKey, true)
				return next(c)
			})
			if errors.Is(err, errMaintenanceActive) {
				return apiErr(c, http.StatusConflict, "RESTORE_ALREADY_RUNNING", "another restore maintenance operation is already active")
			}
			if err != nil && !c.Response().Committed {
				return apiErr(c, http.StatusInternalServerError, "RESTORE_BARRIER_ERROR", "could not establish exclusive restore maintenance")
			}
			return err
		}
		if a.operations.maintenance.Load() {
			return apiErr(c, http.StatusServiceUnavailable, "MAINTENANCE_MODE", "restore maintenance is in progress")
		}
		a.operations.maintenanceGate.RLock()
		defer a.operations.maintenanceGate.RUnlock()
		if a.operations.maintenance.Load() {
			return apiErr(c, http.StatusServiceUnavailable, "MAINTENANCE_MODE", "restore maintenance is in progress")
		}
		return next(c)
	}
}

func (a *App) mutationBarrierMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.operations == nil || !requiresMutationBarrier(c.Request().Method) || isExclusiveOperationPath(c.Request().URL.Path) {
			return next(c)
		}
		return a.operations.withSharedMutation(c.Request().Context(), func() error {
			return next(c)
		})
	}
}
