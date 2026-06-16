package platform

import "testing"

func TestFormatBatesLabel(t *testing.T) {
	got := FormatBatesLabel(BatesOptions{Prefix: "ABC-", Suffix: "-X", ZeroPadding: 4, StartNumber: 7}, 2)
	if got != "ABC-0009-X" {
		t.Fatalf("label=%q", got)
	}
}
