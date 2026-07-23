package app

import (
	"context"
	"errors"
	"testing"
)

func TestPaginationConfigurationBounds(t *testing.T) {
	for _, values := range []struct {
		defaultSize int
		maxSize     int
	}{
		{0, 100},
		{25, 0},
		{101, 101},
		{50, 25},
	} {
		if err := validatePaginationValues(values.defaultSize, values.maxSize); err == nil {
			t.Fatalf("invalid pagination values accepted: default=%d max=%d", values.defaultSize, values.maxSize)
		}
	}
	if err := validatePaginationValues(25, 100); err != nil {
		t.Fatalf("valid pagination values rejected: %v", err)
	}
}

func TestConfigurationOwnershipRejectsDeploymentAndUnknownKeys(t *testing.T) {
	if _, err := validateAdminConfigUpdate(context.Background(), nil, Config{}, backupVolumeKey, "/tmp/other"); !errors.Is(err, errDeploymentOwnedConfig) {
		t.Fatalf("deployment-owned key error = %v", err)
	}
	if _, err := validateAdminConfigUpdate(context.Background(), nil, Config{}, "unknown.key", "1"); !errors.Is(err, errUnsupportedConfigKey) {
		t.Fatalf("unknown key error = %v", err)
	}
}

func TestAdminManagedBackupConfigurationKeys(t *testing.T) {
	for _, test := range []struct {
		key   string
		value string
	}{
		{backupScheduleEnabledKey, "true"},
		{backupScheduleEnabledKey, "false"},
		{backupScheduleIntervalKey, "1m"},
		{backupScheduleIntervalKey, "24h"},
		{backupScheduleIntervalKey, "168h"},
	} {
		if _, err := validateAdminConfigUpdate(context.Background(), nil, Config{}, test.key, test.value); err != nil {
			t.Fatalf("Admin-managed backup config rejected: %s=%q: %v", test.key, test.value, err)
		}
	}
}

func TestAdminBackupConfigurationRejectsInvalidValues(t *testing.T) {
	for _, test := range []struct {
		key   string
		value string
	}{
		{backupScheduleEnabledKey, "yes"},
		{backupScheduleIntervalKey, "59s"},
		{backupScheduleIntervalKey, "169h"},
		{backupScheduleIntervalKey, "invalid"},
	} {
		if _, err := validateAdminConfigUpdate(context.Background(), nil, Config{}, test.key, test.value); err == nil {
			t.Fatalf("invalid Admin backup config accepted: %s=%q", test.key, test.value)
		}
	}
}

func TestPaginationConfigurationRejectsNonIntegerBeforePersistence(t *testing.T) {
	if _, err := validateAdminConfigUpdate(context.Background(), nil, Config{}, paginationDefaultKey, "not-an-integer"); err == nil {
		t.Fatal("non-integer pagination value was accepted")
	}
}

func TestMaximumPageOffsetDoesNotOverflowInt(t *testing.T) {
	offset := (maxSafePageNumber - 1) * absoluteMaxPageSize
	if offset < 0 {
		t.Fatalf("maximum pagination offset overflowed: %d", offset)
	}
}
