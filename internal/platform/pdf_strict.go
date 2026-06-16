package platform

import "fmt"

func RewritePDFWithRedactionsStrict(input string, output string, regions []RedactionRegion) (PDFProcessResult, error) {
	if !toolExists("pdftoppm") {
		return PDFProcessResult{}, fmt.Errorf("strict redaction requires pdftoppm")
	}
	if !pythonHasRasterRedactionDeps() {
		return PDFProcessResult{}, fmt.Errorf("strict redaction requires python PIL and reportlab")
	}
	if err := runPythonRasterRedaction(input, output, regions); err != nil {
		return PDFProcessResult{}, fmt.Errorf("strict raster redaction failed: %w", err)
	}
	return PDFProcessResult{Mode: "raster_redaction_burnin", Details: "PDF pages were rasterized, redaction rectangles burned into pixels, and rebuilt into a new PDF artifact"}, nil
}

func RewritePDFWithBatesStrict(input string, output string, opts BatesOptions) (PDFProcessResult, error) {
	if !pythonHasPDFDrawingDeps() {
		return PDFProcessResult{}, fmt.Errorf("strict Bates numbering requires python pypdf and reportlab")
	}
	if err := runPythonBatesOverlay(input, output, opts); err != nil {
		return PDFProcessResult{}, fmt.Errorf("strict Bates overlay failed: %w", err)
	}
	return PDFProcessResult{Mode: "visible_bates_overlay", Details: "Bates labels were drawn onto rewritten PDF pages"}, nil
}
