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
		{name: "below minimum", value: "59s", want: 0},
		{name: "minimum", value: "1m", want: time.Minute},
		{name: "normal", value: "24h", want: 24 * time.Hour},
		{name: "maximum", value: "168h", want: 7 * 24 * time.Hour},
		{name: "above maximum", value: "169h", want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := backupInterval(tc.value); got != tc.want {
				t.Fatalf("backupInterval(%q) = %s, want %s", tc.value, got, tc.want)
			}
		})
	}
}
