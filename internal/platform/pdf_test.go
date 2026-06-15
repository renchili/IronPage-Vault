package platform

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectPDFValidMinimalFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.pdf")
	raw := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Page >>\nendobj\n%%EOF\n")

	if err := os.WriteFile(path, raw, 0640); err != nil {
		t.Fatal(err)
	}

	info, err := InspectPDF(path, 1024, 10)
	if err != nil {
		t.Fatalf("InspectPDF returned error: %v", err)
	}
	if info.PageCount < 1 {
		t.Fatalf("expected at least one page, got %d", info.PageCount)
	}
	if info.Size == 0 || info.SHA256 == "" {
		t.Fatalf("expected size and digest to be populated")
	}
}

func TestInspectPDFRejectsNonPDF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.txt")

	if err := os.WriteFile(path, []byte("not a pdf"), 0640); err != nil {
		t.Fatal(err)
	}

	if _, err := InspectPDF(path, 1024, 10); err == nil {
		t.Fatalf("expected non-PDF file to be rejected")
	}
}

func TestInspectPDFRejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.pdf")
	raw := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Page >>\nendobj\n%%EOF\n")

	if err := os.WriteFile(path, raw, 0640); err != nil {
		t.Fatal(err)
	}

	if _, err := InspectPDF(path, 5, 10); err == nil {
		t.Fatalf("expected oversized PDF to be rejected")
	}
}

func TestAppendPDFTransformMarker(t *testing.T) {
	raw := []byte("%PDF-1.4\n%%EOF")
	out := AppendPDFTransformMarker(raw, "review marker")

	if !bytes.Contains(out, []byte("IronPage Vault transform: review marker")) {
		t.Fatalf("expected transform marker")
	}
	if !bytes.HasPrefix(out, raw) {
		t.Fatalf("expected append-only transform")
	}
}
