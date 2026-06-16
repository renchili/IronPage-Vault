package service

import "ironpage-vault/internal/platform"

func ApplyRedactionBurnIn(input string, output string, regions []platform.RedactionRegion) (platform.PDFProcessResult, error) {
	return platform.RewritePDFWithRedactions(input, output, regions)
}

func ApplyBatesNumbering(input string, output string, opts platform.BatesOptions) (platform.PDFProcessResult, error) {
	return platform.RewritePDFWithBates(input, output, opts)
}
