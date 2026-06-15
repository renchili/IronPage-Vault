package app

import "testing"

func TestBatesVersionPath(t *testing.T) {
	got := batesVersionPath("v1.pdf", 3)
	want := "v1_bates_v3.pdf"
	if got != want {
		t.Fatalf("batesVersionPath=%q want %q", got, want)
	}
}

func TestBatesMarker(t *testing.T) {
	got := batesMarker("IPV-", "-A", 6, 12)
	want := "Bates applied: prefix=IPV-, suffix=-A, zero_padding=6, start=12"
	if got != want {
		t.Fatalf("batesMarker=%q want %q", got, want)
	}
}
