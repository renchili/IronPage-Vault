package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatBatesLabel(t *testing.T) {
	got := FormatBatesLabel(BatesOptions{Prefix: "ABC-", Suffix: "-X", ZeroPadding: 4, StartNumber: 7}, 2)
	if got != "ABC-0009-X" {
		t.Fatalf("label=%q", got)
	}
}

func TestPDFFallbackCreatesDistinctArtifact(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.pdf")
	out := filepath.Join(dir, "out.pdf")
	if err := os.WriteFile(in, []byte("%PDF-1.4\n% test\n"), 0640); err != nil {
		t.Fatal(err)
	}
	result, err := copyWithProcessManifest(in, out, "test_mode", map[string]interface{}{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mode != "test_mode" {
		t.Fatalf("mode=%q", result.Mode)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) <= len("%PDF-1.4\n% test\n") {
		t.Fatalf("expected rewritten artifact to contain process manifest")
	}
}
