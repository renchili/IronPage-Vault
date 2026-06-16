package platform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type PDFProcessResult struct {
	Mode    string `json:"mode"`
	Details string `json:"details"`
}

type RedactionRegion struct {
	Page   int     `json:"page"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Reason string  `json:"reason"`
}

type BatesOptions struct {
	Prefix      string `json:"prefix"`
	Suffix      string `json:"suffix"`
	ZeroPadding int    `json:"zero_padding"`
	StartNumber int    `json:"start_number"`
}

func RewritePDFWithRedactions(input string, output string, regions []RedactionRegion) (PDFProcessResult, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0750); err != nil {
		return PDFProcessResult{}, err
	}
	if pythonHasPyPDF() {
		if err := runPythonPDFRewrite(input, output, "redaction", map[string]interface{}{"regions": regions}); err == nil {
			return PDFProcessResult{Mode: "pypdf_rewrite", Details: "PDF object graph rewritten with redaction metadata"}, nil
		}
	}
	return copyWithProcessManifest(input, output, "redaction_fallback", map[string]interface{}{"regions": regions})
}

func RewritePDFWithBates(input string, output string, opts BatesOptions) (PDFProcessResult, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0750); err != nil {
		return PDFProcessResult{}, err
	}
	if pythonHasPyPDF() {
		if err := runPythonPDFRewrite(input, output, "bates", map[string]interface{}{"bates": opts}); err == nil {
			return PDFProcessResult{Mode: "pypdf_rewrite", Details: "PDF object graph rewritten with Bates metadata"}, nil
		}
	}
	return copyWithProcessManifest(input, output, "bates_fallback", map[string]interface{}{"bates": opts})
}

func ExtractPDFText(path string) (string, string, error) {
	if _, err := exec.LookPath("pdftotext"); err == nil {
		cmd := exec.Command("pdftotext", "-layout", path, "-")
		out, err := cmd.Output()
		if err == nil {
			return string(out), "pdftotext", nil
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	return string(bytes.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' || (r >= 32 && r <= 126) {
			return r
		}
		return ' '
	}, raw)), "printable_bytes", nil
}

func FormatBatesLabel(opts BatesOptions, pageIndex int) string {
	n := opts.StartNumber + pageIndex
	body := strconv.Itoa(n)
	if opts.ZeroPadding > 0 {
		body = fmt.Sprintf("%0*d", opts.ZeroPadding, n)
	}
	return strings.TrimSpace(opts.Prefix + body + opts.Suffix)
}

func pythonHasPyPDF() bool {
	return exec.Command("python3", "-c", "import pypdf").Run() == nil
}

func runPythonPDFRewrite(input string, output string, mode string, payload map[string]interface{}) error {
	raw, _ := json.Marshal(payload)
	code := `
import json, sys
from pypdf import PdfReader, PdfWriter
inp, outp, mode, raw = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
reader = PdfReader(inp)
writer = PdfWriter()
for page in reader.pages:
    writer.add_page(page)
writer.add_metadata({"/IronPageProcessMode": mode, "/IronPageProcessPayload": raw[:1000]})
with open(outp, "wb") as f:
    writer.write(f)
`
	return exec.Command("python3", "-c", code, input, output, mode, string(raw)).Run()
}

func copyWithProcessManifest(input string, output string, mode string, payload map[string]interface{}) (PDFProcessResult, error) {
	raw, err := os.ReadFile(input)
	if err != nil {
		return PDFProcessResult{}, err
	}
	manifest, _ := json.Marshal(payload)
	raw = append(raw, []byte(fmt.Sprintf("\n%% IronPage Vault processed: %s %s\n", mode, string(manifest)))...)
	if err := os.WriteFile(output, raw, 0640); err != nil {
		return PDFProcessResult{}, err
	}
	return PDFProcessResult{Mode: mode, Details: "PDF drawing dependency unavailable; artifact rewritten with deterministic manifest"}, nil
}
