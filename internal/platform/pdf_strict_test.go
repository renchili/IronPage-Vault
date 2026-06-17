package platform

import "testing"

func TestStrictPDFEntrypointsDoNotReturnFallbackModes(t *testing.T) {
	if !toolExists("pdftoppm") || !pythonHasRasterRedactionDeps() {
		_, err := RewritePDFWithRedactionsStrict("/missing/input.pdf", "/missing/output.pdf", nil)
		if err == nil {
			t.Fatalf("strict redaction must fail when strict dependencies or input are unavailable")
		}
	}
	if !pythonHasPDFDrawingDeps() {
		_, err := RewritePDFWithBatesStrict("/missing/input.pdf", "/missing/output.pdf", BatesOptions{Prefix: "T", StartNumber: 1})
		if err == nil {
			t.Fatalf("strict Bates must fail when strict dependencies or input are unavailable")
		}
	}
}
