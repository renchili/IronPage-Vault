package app

import "testing"

func TestRedactedVersionPath(t *testing.T) {
	got := redactedVersionPath("case.pdf", 4)
	want := "case_redacted_v4.pdf"
	if got != want {
		t.Fatalf("redactedVersionPath=%q want %q", got, want)
	}
}

func TestRedactionTransformMarker(t *testing.T) {
	if redactionTransformMarker() == "" {
		t.Fatalf("marker must not be empty")
	}
}
