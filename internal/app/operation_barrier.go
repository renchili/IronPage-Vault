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

const operationAdvisoryLockID int64 = 52833707454590

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
		if !locked {
			_ = conn.Close()
		}
	}()
	if _, err := conn.ExecContext(ctx, lockSQL, operationAdvisoryLockID); err != nil {
		return err
	}
	locked = true
	operationErr := fn()
	releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, unlockErr := conn.ExecContext(releaseCtx, unlockSQL, operationAdvisoryLockID)
	closeErr := conn.Close()
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

func isExclusiveOperationPath(path string) bool {
	return path == "/api/admin/backup/run" || path == "/api/admin/backup/restore"
}

func (a *App) maintenanceMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.operations == nil || c.Request().URL.Path == "/api/admin/backup/restore" {
			return next(c)
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
