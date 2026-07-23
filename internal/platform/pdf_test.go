package platform

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func buildClassicTestPDF(pageCount int, compressedPayload string) []byte {
	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	writeObject := func(number int, body string) {
		offsets = append(offsets, document.Len())
		fmt.Fprintf(&document, "%d 0 obj\n%s\nendobj\n", number, body)
	}

	writeObject(1, "<< /Type /Catalog /Pages 2 0 R >>")
	kids := make([]string, 0, pageCount)
	for index := 0; index < pageCount; index++ {
		kids = append(kids, fmt.Sprintf("%d 0 R", index+3))
	}
	writeObject(2, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), pageCount))
	for index := 0; index < pageCount; index++ {
		writeObject(index+3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>")
	}
	if compressedPayload != "" {
		var compressed bytes.Buffer
		writer := zlib.NewWriter(&compressed)
		_, _ = writer.Write([]byte(compressedPayload))
		_ = writer.Close()
		number := pageCount + 3
		offsets = append(offsets, document.Len())
		fmt.Fprintf(&document, "%d 0 obj\n<< /Length %d /Filter /FlateDecode >>\nstream\n", number, compressed.Len())
		document.Write(compressed.Bytes())
		document.WriteString("\nendstream\nendobj\n")
	}

	xrefOffset := document.Len()
	fmt.Fprintf(&document, "xref\n0 %d\n", len(offsets))
	document.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&document, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&document, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xrefOffset)
	return document.Bytes()
}

func writeXRefEntry(target *bytes.Buffer, kind byte, field2 uint32, field3 uint16) {
	target.WriteByte(kind)
	_ = binary.Write(target, binary.BigEndian, field2)
	_ = binary.Write(target, binary.BigEndian, field3)
}

func buildObjectStreamTestPDF() []byte {
	var document bytes.Buffer
	document.WriteString("%PDF-1.5\n%\xe2\xe3\xcf\xd3\n")
	offsets := map[int]int{}
	writeObject := func(number int, body string) {
		offsets[number] = document.Len()
		fmt.Fprintf(&document, "%d 0 obj\n%s\nendobj\n", number, body)
	}

	writeObject(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObject(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")

	objectStreamBody := []byte("3 0 << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>")
	var compressedObjects bytes.Buffer
	objectWriter := zlib.NewWriter(&compressedObjects)
	_, _ = objectWriter.Write(objectStreamBody)
	_ = objectWriter.Close()
	offsets[4] = document.Len()
	fmt.Fprintf(&document, "4 0 obj\n<< /Type /ObjStm /N 1 /First 4 /Length %d /Filter /FlateDecode >>\nstream\n", compressedObjects.Len())
	document.Write(compressedObjects.Bytes())
	document.WriteString("\nendstream\nendobj\n")

	offsets[5] = document.Len()
	var xref bytes.Buffer
	writeXRefEntry(&xref, 0, 0, 65535)
	writeXRefEntry(&xref, 1, uint32(offsets[1]), 0)
	writeXRefEntry(&xref, 1, uint32(offsets[2]), 0)
	writeXRefEntry(&xref, 2, 4, 0)
	writeXRefEntry(&xref, 1, uint32(offsets[4]), 0)
	writeXRefEntry(&xref, 1, uint32(offsets[5]), 0)
	var compressedXRef bytes.Buffer
	xrefWriter := zlib.NewWriter(&compressedXRef)
	_, _ = xrefWriter.Write(xref.Bytes())
	_ = xrefWriter.Close()
	fmt.Fprintf(&document, "5 0 obj\n<< /Type /XRef /Size 6 /Root 1 0 R /W [1 4 2] /Index [0 6] /Length %d /Filter /FlateDecode >>\nstream\n", compressedXRef.Len())
	document.Write(compressedXRef.Bytes())
	document.WriteString("\nendstream\nendobj\n")
	fmt.Fprintf(&document, "startxref\n%d\n%%%%EOF\n", offsets[5])
	return document.Bytes()
}

func writePDF(t *testing.T, raw []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.pdf")
	if err := os.WriteFile(path, raw, 0640); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInspectPDFPageBoundaries(t *testing.T) {
	for _, test := range []struct {
		pages   int
		wantErr bool
	}{
		{0, true},
		{1, false},
		{499, false},
		{500, false},
		{501, true},
	} {
		path := writePDF(t, buildClassicTestPDF(test.pages, ""))
		info, err := InspectPDF(path, 16*1024*1024, 500)
		if test.wantErr {
			if err == nil {
				t.Fatalf("expected %d-page PDF to be rejected", test.pages)
			}
			continue
		}
		if err != nil {
			t.Fatalf("InspectPDF(%d pages): %v", test.pages, err)
		}
		if info.PageCount != test.pages {
			t.Fatalf("page count = %d, want %d", info.PageCount, test.pages)
		}
	}
}

func TestInspectPDFIgnoresPagesRootAndCompressedStreamTokens(t *testing.T) {
	payload := strings.Repeat("/Type /Page /Type /Pages ", 600)
	path := writePDF(t, buildClassicTestPDF(1, payload))
	info, err := InspectPDF(path, 1024*1024, 500)
	if err != nil {
		t.Fatalf("InspectPDF returned error: %v", err)
	}
	if info.PageCount != 1 {
		t.Fatalf("page count = %d, want 1", info.PageCount)
	}
}

func TestInspectPDFReadsPageFromCompressedObjectStream(t *testing.T) {
	path := writePDF(t, buildObjectStreamTestPDF())
	info, err := InspectPDF(path, 1024*1024, 500)
	if err != nil {
		t.Fatalf("InspectPDF object stream: %v", err)
	}
	if info.PageCount != 1 {
		t.Fatalf("object-stream page count = %d, want 1", info.PageCount)
	}
}

func TestInspectPDFRejectsMalformedPDF(t *testing.T) {
	path := writePDF(t, []byte("%PDF-1.4\nnot a valid object graph\n%%EOF\n"))
	if _, err := InspectPDF(path, 1024, 10); err == nil {
		t.Fatal("expected malformed PDF to be rejected")
	}
}

func TestInspectPDFRejectsNonPDF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.txt")
	if err := os.WriteFile(path, []byte("not a pdf"), 0640); err != nil {
		t.Fatal(err)
	}
	if _, err := InspectPDF(path, 1024, 10); err == nil {
		t.Fatal("expected non-PDF file to be rejected")
	}
}

func TestInspectPDFRejectsOversizedFile(t *testing.T) {
	path := writePDF(t, buildClassicTestPDF(1, ""))
	if _, err := InspectPDF(path, 5, 10); err == nil {
		t.Fatal("expected oversized PDF to be rejected")
	}
}

func TestAppendPDFTransformMarker(t *testing.T) {
	raw := []byte("%PDF-1.4\n%%EOF")
	out := AppendPDFTransformMarker(raw, "review marker")
	if !bytes.Contains(out, []byte("IronPage Vault transform: review marker")) {
		t.Fatal("expected transform marker")
	}
	if !bytes.HasPrefix(out, raw) {
		t.Fatal("expected append-only transform")
	}
}
