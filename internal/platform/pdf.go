package platform

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// PDFInfo describes locally inspected PDF metadata used by document intake.
type PDFInfo struct {
	PageCount int
	Size      int64
	SHA256    string
}

// InspectPDF validates a local PDF file and returns lightweight intake metadata.
//
// This helper intentionally stays local and deterministic: it checks the PDF
// header, size limit, approximate page count, and file digest without calling a
// remote processor.
func InspectPDF(path string, maxBytes int64, maxPages int) (PDFInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return PDFInfo{}, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return PDFInfo{}, err
	}
	if st.Size() > maxBytes {
		return PDFInfo{}, fmt.Errorf("pdf exceeds max size")
	}

	head := make([]byte, 5)
	if _, err := io.ReadFull(f, head); err != nil {
		return PDFInfo{}, err
	}
	if string(head) != "%PDF-" {
		return PDFInfo{}, fmt.Errorf("not a pdf file")
	}

	if _, err := f.Seek(0, 0); err != nil {
		return PDFInfo{}, err
	}
	raw, err := io.ReadAll(f)
	if err != nil {
		return PDFInfo{}, err
	}

	pages := bytes.Count(raw, []byte("/Type /Page"))
	if pages == 0 {
		pages = 1
	}
	if pages > maxPages {
		return PDFInfo{}, fmt.Errorf("pdf exceeds max page count")
	}

	if _, err := f.Seek(0, 0); err != nil {
		return PDFInfo{}, err
	}
	sum, err := FileDigest(f)
	if err != nil {
		return PDFInfo{}, err
	}

	return PDFInfo{
		PageCount: pages,
		Size:      st.Size(),
		SHA256:    sum,
	}, nil
}

// AppendPDFMetadataMarker returns raw PDF bytes with a local transform marker.
func AppendPDFMetadataMarker(raw []byte, marker string) []byte {
	return append(raw, []byte("\n% IronPage Vault transform: "+marker+"\n")...)
}

// AppendPDFTransformMarker is a compatibility wrapper for append-only transforms.
func AppendPDFTransformMarker(raw []byte, marker string) []byte {
	return AppendPDFMetadataMarker(raw, marker)
}

// AppendPDFMetadataMarkerFile writes a marked PDF copy from src to dst.
func AppendPDFMetadataMarkerFile(src, dst, marker string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(string(raw), "%PDF-") {
		return fmt.Errorf("source is not pdf")
	}
	raw = AppendPDFMetadataMarker(raw, marker)
	return os.WriteFile(dst, raw, 0640)
}

// ApplyAppendOnlyPDFTransform applies the repository's append-only PDF transform.
func ApplyAppendOnlyPDFTransform(src, dst, marker string) error {
	return AppendPDFMetadataMarkerFile(src, dst, marker)
}
