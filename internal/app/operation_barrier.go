package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
	db               *sqlx.DB
	restoreAdmission sync.Mutex
	maintenanceGate  sync.RWMutex
	maintenance      atomic.Bool
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

func isRestoreResolutionPath(path string) bool {
	return strings.HasPrefix(path, "/api/admin/backup/restore/") && strings.HasSuffix(path, "/resolve")
}

func isExclusiveOperationPath(path string) bool {
	return path == "/api/admin/backup/run" || path == "/api/admin/backup/restore" || isRestoreResolutionPath(path)
}

// maintenanceMiddleware rejects ordinary traffic while restore owns the
// maintenance gate. Restore requests acquire a non-blocking admission mutex so
// only one request may authenticate and attempt maintenance at a time; an
// unauthenticated request releases admission immediately after auth fails.
func (a *App) maintenanceMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.operations == nil {
			return next(c)
		}
		if isRestoreOperationRequest(c) {
			if !a.operations.restoreAdmission.TryLock() {
				return apiErr(c, http.StatusConflict, "RESTORE_ALREADY_RUNNING", "another restore request is already active")
			}
			defer a.operations.restoreAdmission.Unlock()
			if a.operations.maintenance.Load() {
				return apiErr(c, http.StatusConflict, "RESTORE_ALREADY_RUNNING", "another restore maintenance operation is already active")
			}
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

// restoreMaintenanceMiddleware runs after authentication and Admin role
// validation. It marks maintenance active, drains ordinary requests, obtains
// the exclusive advisory lock and keeps all of those boundaries through the
// restore handler response.
func (a *App) restoreMaintenanceMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.operations == nil {
			return apiErr(c, http.StatusInternalServerError, "RESTORE_BARRIER_ERROR", "restore maintenance barrier is unavailable")
		}
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
}

func (a *App) exclusiveOperationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.operations == nil {
			return apiErr(c, http.StatusInternalServerError, "OPERATION_BARRIER_ERROR", "exclusive operation barrier is unavailable")
		}
		err := a.operations.withExclusiveOperation(c.Request().Context(), func() error {
			return next(c)
		})
		if err != nil && !c.Response().Committed {
			return apiErr(c, http.StatusInternalServerError, "OPERATION_BARRIER_ERROR", "could not establish exclusive operation barrier")
		}
		return err
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
