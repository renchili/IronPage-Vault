package service

import "ironpage-vault/internal/platform"

func ApplyRedactionBurnIn(input string, output string, regions []platform.RedactionRegion) (platform.PDFProcessResult, error) {
	return platform.RewritePDFWithRedactionsStrict(input, output, regions)
}

func ApplyBatesNumbering(input string, output string, opts platform.BatesOptions) (platform.PDFProcessResult, error) {
	return platform.RewritePDFWithBatesStrict(input, output, opts)
}
