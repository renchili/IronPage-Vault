package app

import (
	"testing"
	"time"
)

func TestBackupInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "disabled by default", value: "", want: 0},
		{name: "valid interval", value: "24h", want: 24 * time.Hour},
		{name: "short test interval", value: "250ms", want: 250 * time.Millisecond},
		{name: "invalid interval", value: "tomorrow", want: 0},
		{name: "negative interval", value: "-1h", want: 0},
		{name: "zero disables", value: "0s", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := backupInterval(tt.value); got != tt.want {
				t.Fatalf("backupInterval(%q)=%s want %s", tt.value, got, tt.want)
			}
		})
	}
}
