package app

import (
	"testing"
	"time"
)

func TestBackupIntervalParsing(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "empty", value: "", want: 0},
		{name: "invalid", value: "not-a-duration", want: 0},
		{name: "zero", value: "0s", want: 0},
		{name: "negative", value: "-1s", want: 0},
		{name: "valid", value: "15m", want: 15 * time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := backupInterval(tc.value); got != tc.want {
				t.Fatalf("backupInterval(%q) = %s, want %s", tc.value, got, tc.want)
			}
		})
	}
}
