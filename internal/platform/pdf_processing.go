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
	if pythonHasPDFDrawingDeps() {
		if err := runPythonRedactionOverlay(input, output, regions); err == nil {
			return PDFProcessResult{Mode: "visible_redaction_overlay", Details: "redaction rectangles were drawn onto rewritten PDF pages"}, nil
		}
	}
	return copyWithProcessManifest(input, output, "redaction_fallback", map[string]interface{}{"regions": regions})
}

func RewritePDFWithBates(input string, output string, opts BatesOptions) (PDFProcessResult, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0750); err != nil {
		return PDFProcessResult{}, err
	}
	if pythonHasPDFDrawingDeps() {
		if err := runPythonBatesOverlay(input, output, opts); err == nil {
			return PDFProcessResult{Mode: "visible_bates_overlay", Details: "Bates labels were drawn onto rewritten PDF pages"}, nil
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

func pythonHasPDFDrawingDeps() bool {
	return exec.Command("python3", "-c", "import pypdf, reportlab").Run() == nil
}

func runPythonRedactionOverlay(input string, output string, regions []RedactionRegion) error {
	raw, _ := json.Marshal(regions)
	code := `
import io, json, sys
from pypdf import PdfReader, PdfWriter
from reportlab.pdfgen import canvas
from reportlab.lib.colors import black
inp, outp, raw = sys.argv[1], sys.argv[2], sys.argv[3]
regions = json.loads(raw)
reader = PdfReader(inp)
writer = PdfWriter()
for idx, page in enumerate(reader.pages, start=1):
    width = float(page.mediabox.width)
    height = float(page.mediabox.height)
    packet = io.BytesIO()
    c = canvas.Canvas(packet, pagesize=(width, height))
    c.setFillColor(black)
    for r in regions:
        if int(r.get("page", 0)) == idx:
            x = float(r.get("x", 0)); y = float(r.get("y", 0))
            w = float(r.get("width", 0)); h = float(r.get("height", 0))
            c.rect(x, height - y - h, w, h, stroke=0, fill=1)
    c.save()
    packet.seek(0)
    overlay = PdfReader(packet)
    if overlay.pages:
        page.merge_page(overlay.pages[0])
    writer.add_page(page)
writer.add_metadata({"/IronPageProcessMode":"visible_redaction_overlay"})
with open(outp, "wb") as f:
    writer.write(f)
`
	return exec.Command("python3", "-c", code, input, output, string(raw)).Run()
}

func runPythonBatesOverlay(input string, output string, opts BatesOptions) error {
	raw, _ := json.Marshal(opts)
	code := `
import io, json, sys
from pypdf import PdfReader, PdfWriter
from reportlab.pdfgen import canvas
from reportlab.lib.colors import black
inp, outp, raw = sys.argv[1], sys.argv[2], sys.argv[3]
opts = json.loads(raw)
reader = PdfReader(inp)
writer = PdfWriter()
prefix = opts.get("prefix", "")
suffix = opts.get("suffix", "")
pad = int(opts.get("zero_padding", 0) or 0)
start = int(opts.get("start_number", 1) or 1)
for idx, page in enumerate(reader.pages):
    width = float(page.mediabox.width)
    height = float(page.mediabox.height)
    number = start + idx
    body = str(number).zfill(pad) if pad > 0 else str(number)
    label = f"{prefix}{body}{suffix}"
    packet = io.BytesIO()
    c = canvas.Canvas(packet, pagesize=(width, height))
    c.setFillColor(black)
    c.setFont("Helvetica", 9)
    c.drawRightString(width - 36, 24, label)
    c.save()
    packet.seek(0)
    overlay = PdfReader(packet)
    if overlay.pages:
        page.merge_page(overlay.pages[0])
    writer.add_page(page)
writer.add_metadata({"/IronPageProcessMode":"visible_bates_overlay"})
with open(outp, "wb") as f:
    writer.write(f)
`
	return exec.Command("python3", "-c", code, input, output, string(raw)).Run()
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
