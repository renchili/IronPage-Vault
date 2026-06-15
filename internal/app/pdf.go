package app

import "ironpage-vault/internal/platform"

type PDFInfo = platform.PDFInfo

func InspectPDF(path string, maxBytes int64, maxPages int) (PDFInfo, error) {
	return platform.InspectPDF(path, maxBytes, maxPages)
}

func appendPDFTransformMarker(raw []byte, marker string) []byte {
	return platform.AppendPDFTransformMarker(raw, marker)
}

func ApplyAppendOnlyPDFTransform(src, dst, marker string) error {
	return platform.ApplyAppendOnlyPDFTransform(src, dst, marker)
}
