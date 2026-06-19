package platform

import (
	"strings"
	"testing"
)

func TestFileDigest(t *testing.T) {
	got, err := FileDigest(strings.NewReader("abc"))
	if err != nil {
		t.Fatalf("FileDigest error: %v", err)
	}
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("FileDigest=%q want %q", got, want)
	}
}
