package app

import "testing"

func TestBackupScheduleConfigurationValidation(t *testing.T) {
	for _, test := range []struct {
		key   string
		value string
		want  string
	}{
		{backupScheduleEnabledKey, "true", "true"},
		{backupScheduleEnabledKey, "false", "false"},
		{backupScheduleIntervalKey, "1m", "1m0s"},
		{backupScheduleIntervalKey, "24h", "24h0m0s"},
		{backupScheduleIntervalKey, "168h", "168h0m0s"},
	} {
		got, err := validateBackupScheduleValue(test.key, test.value)
		if err != nil {
			t.Fatalf("validateBackupScheduleValue(%q,%q): %v", test.key, test.value, err)
		}
		if got != test.want {
			t.Fatalf("normalized value = %q, want %q", got, test.want)
		}
	}
	for _, test := range []struct {
		key   string
		value string
	}{
		{backupScheduleEnabledKey, "enabled"},
		{backupScheduleIntervalKey, "59s"},
		{backupScheduleIntervalKey, "169h"},
		{backupScheduleIntervalKey, "not-a-duration"},
	} {
		if _, err := validateBackupScheduleValue(test.key, test.value); err == nil {
			t.Fatalf("invalid backup schedule accepted: %s=%q", test.key, test.value)
		}
	}
}
