package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

const absoluteMaxPageSize = 100

type paginationConfigRow struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

func (a *App) paginationLimits(ctx context.Context, executor sqlx.ExtContext) (int, int, error) {
	defaultSize := a.cfg.DefaultPageSize
	maxSize := a.cfg.MaxPageSize
	rows := []paginationConfigRow{}
	if err := sqlx.SelectContext(ctx, executor, &rows, `SELECT key,value FROM config_entries WHERE key IN ('pagination.default_page_size','pagination.max_page_size')`); err != nil {
		return 0, 0, err
	}
	for _, row := range rows {
		value, err := strconv.Atoi(row.Value)
		if err != nil {
			return 0, 0, fmt.Errorf("configuration %s must be an integer", row.Key)
		}
		switch row.Key {
		case "pagination.default_page_size":
			defaultSize = value
		case "pagination.max_page_size":
			maxSize = value
		}
	}
	if defaultSize < 1 || maxSize < 1 || maxSize > absoluteMaxPageSize || defaultSize > maxSize {
		return 0, 0, fmt.Errorf("pagination configuration must satisfy 1 <= default <= max <= %d", absoluteMaxPageSize)
	}
	return defaultSize, maxSize, nil
}

func (a *App) configuredPage(c echo.Context) (int, int, error) {
	defaultSize, maxSize, err := a.paginationLimits(c.Request().Context(), a.db)
	if err != nil {
		return 0, 0, err
	}
	page := atoiDefault(c.QueryParam("page"), 1)
	size := atoiDefault(c.QueryParam("page_size"), defaultSize)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultSize
	}
	if size > maxSize {
		size = maxSize
	}
	return page, size, nil
}
