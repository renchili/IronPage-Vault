package app

import (
	"testing"
	"time"
)

func TestBackupInterval(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"", 0},
		{"24h", 24 * time.Hour},
		{"250ms", 250 * time.Millisecond},
		{"bad", 0},
		{"-1h", 0},
		{"0s", 0},
	}
	for _, test := range tests {
		if got := backupInterval(test.input); got != test.want {
			t.Fatalf("backupInterval(%q)=%s want %s", test.input, got, test.want)
		}
	}
}
