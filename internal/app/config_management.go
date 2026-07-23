package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	paginationDefaultKey      = "pagination.default_page_size"
	paginationMaxKey          = "pagination.max_page_size"
	backupVolumeKey           = "backup.local_volume"
	backupScheduleEnabledKey  = "backup.schedule_enabled"
	backupScheduleIntervalKey = "backup.interval"
	minimumBackupInterval     = time.Minute
	maximumBackupInterval     = 7 * 24 * time.Hour
)

var (
	errDeploymentOwnedConfig = errors.New("configuration key is deployment-owned")
	errUnsupportedConfigKey  = errors.New("configuration key is not Admin-managed")
)

func validateBackupScheduleValue(key, rawValue string) (string, error) {
	switch key {
	case backupScheduleEnabledKey:
		value, err := strconv.ParseBool(strings.TrimSpace(rawValue))
		if err != nil {
			return "", fmt.Errorf("backup.schedule_enabled must be true or false")
		}
		return strconv.FormatBool(value), nil
	case backupScheduleIntervalKey:
		value, err := time.ParseDuration(strings.TrimSpace(rawValue))
		if err != nil {
			return "", fmt.Errorf("backup.interval must be a Go duration such as 24h")
		}
		if value < minimumBackupInterval || value > maximumBackupInterval {
			return "", fmt.Errorf("backup.interval must be between %s and %s", minimumBackupInterval, maximumBackupInterval)
		}
		return value.String(), nil
	default:
		return "", errUnsupportedConfigKey
	}
}

func validateAdminConfigUpdate(ctx context.Context, tx *sqlx.Tx, cfg Config, key, rawValue string) (string, error) {
	if key == backupVolumeKey {
		return "", errDeploymentOwnedConfig
	}
	if key == backupScheduleEnabledKey || key == backupScheduleIntervalKey {
		return validateBackupScheduleValue(key, rawValue)
	}
	if key != paginationDefaultKey && key != paginationMaxKey {
		return "", errUnsupportedConfigKey
	}
	value, err := strconv.Atoi(strings.TrimSpace(rawValue))
	if err != nil {
		return "", fmt.Errorf("pagination configuration must be an integer")
	}
	rows := []paginationConfigRow{}
	if err := tx.SelectContext(ctx, &rows, `SELECT key,value FROM config_entries WHERE key IN ($1,$2) FOR UPDATE`, paginationDefaultKey, paginationMaxKey); err != nil {
		return "", err
	}
	defaultSize := cfg.DefaultPageSize
	maxSize := cfg.MaxPageSize
	for _, row := range rows {
		parsed, err := strconv.Atoi(row.Value)
		if err != nil {
			return "", fmt.Errorf("stored configuration %s must be an integer", row.Key)
		}
		switch row.Key {
		case paginationDefaultKey:
			defaultSize = parsed
		case paginationMaxKey:
			maxSize = parsed
		}
	}
	if key == paginationDefaultKey {
		defaultSize = value
	} else {
		maxSize = value
	}
	if err := validatePaginationValues(defaultSize, maxSize); err != nil {
		return "", err
	}
	return strconv.Itoa(value), nil
}
