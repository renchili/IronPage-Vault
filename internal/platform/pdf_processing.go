package platform

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
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

type TextBlock struct {
	Page int     `json:"page"`
	XMin float64 `json:"xmin"`
	YMin float64 `json:"ymin"`
	XMax float64 `json:"xmax"`
	YMax float64 `json:"ymax"`
	Text string  `json:"text"`
}

func RewritePDFWithRedactions(input string, output string, regions []RedactionRegion) (PDFProcessResult, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0750); err != nil {
		return PDFProcessResult{}, err
	}
	if pythonHasRasterRedactionDeps() && toolExists("pdftoppm") {
		if err := runPythonRasterRedaction(input, output, regions); err == nil {
			return PDFProcessResult{Mode: "raster_redaction_burnin", Details: "PDF pages were rasterized, redaction rectangles burned into pixels, and rebuilt into a new PDF artifact"}, nil
		}
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

func ExtractPDFTextBlocks(path string) ([]TextBlock, string, error) {
	if _, err := exec.LookPath("pdftotext"); err == nil {
		out, err := exec.Command("pdftotext", "-bbox", path, "-").Output()
		if err == nil {
			blocks := parseBBOXTextBlocks(out)
			if len(blocks) > 0 {
				return blocks, "pdftotext_bbox", nil
			}
		}
	}
	text, mode, err := ExtractPDFText(path)
	if err != nil {
		return nil, mode, err
	}
	lines := strings.Split(text, "\n")
	blocks := make([]TextBlock, 0, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		blocks = append(blocks, TextBlock{Page: 1, XMin: 0, YMin: float64(i * 12), XMax: 0, YMax: float64(i*12 + 12), Text: line})
	}
	return blocks, mode, nil
}

func parseBBOXTextBlocks(raw []byte) []TextBlock {
	type word struct {
		XMin float64 `xml:"xMin,attr"`
		YMin float64 `xml:"yMin,attr"`
		XMax float64 `xml:"xMax,attr"`
		YMax float64 `xml:"yMax,attr"`
		Text string  `xml:",chardata"`
	}
	type page struct {
		Number int    `xml:"number,attr"`
		Words  []word `xml:"word"`
	}
	type doc struct {
		Pages []page `xml:"body>doc>page"`
	}
	var d doc
	if err := xml.Unmarshal(raw, &d); err != nil {
		return nil
	}
	var blocks []TextBlock
	for _, p := range d.Pages {
		for _, w := range p.Words {
			t := strings.TrimSpace(w.Text)
			if t == "" {
				continue
			}
			blocks = append(blocks, TextBlock{Page: p.Number, XMin: w.XMin, YMin: w.YMin, XMax: w.XMax, YMax: w.YMax, Text: t})
		}
	}
	return blocks
}

func DiffTextBlocks(left []TextBlock, right []TextBlock) (added []TextBlock, removed []TextBlock, modified []TextBlock) {
	leftByPosition := map[string]TextBlock{}
	rightByPosition := map[string]TextBlock{}
	for _, block := range left {
		leftByPosition[textBlockPositionKey(block)] = block
	}
	for _, block := range right {
		rightByPosition[textBlockPositionKey(block)] = block
	}

	for position, rightBlock := range rightByPosition {
		leftBlock, exists := leftByPosition[position]
		if !exists {
			added = append(added, rightBlock)
			continue
		}
		if leftBlock.Text != rightBlock.Text {
			modified = append(modified, rightBlock)
		}
	}
	for position, leftBlock := range leftByPosition {
		if _, exists := rightByPosition[position]; !exists {
			removed = append(removed, leftBlock)
		}
	}
	return added, removed, modified
}

func textBlockPositionKey(block TextBlock) string {
	return fmt.Sprintf("%d|%.2f|%.2f|%.2f|%.2f", block.Page, block.XMin, block.YMin, block.XMax, block.YMax)
}

func textBlockKey(block TextBlock) string {
	return textBlockPositionKey(block) + "|" + block.Text
}

func FormatBatesLabel(opts BatesOptions, pageIndex int) string {
	n := opts.StartNumber + pageIndex
	body := strconv.Itoa(n)
	if opts.ZeroPadding > 0 {
		body = fmt.Sprintf("%0*d", opts.ZeroPadding, n)
	}
	return strings.TrimSpace(opts.Prefix + body + opts.Suffix)
}

func toolExists(name string) bool { _, err := exec.LookPath(name); return err == nil }

func pythonHasPDFDrawingDeps() bool {
	return exec.Command("python3", "-c", "import pypdf, reportlab").Run() == nil
}

func pythonHasRasterRedactionDeps() bool {
	return exec.Command("python3", "-c", "import PIL, reportlab").Run() == nil
}

func runPythonRasterRedaction(input string, output string, regions []RedactionRegion) error {
	raw, _ := json.Marshal(regions)
	tmp := filepath.Join(os.TempDir(), "ironpage_redact_"+strconv.FormatInt(int64(os.Getpid()), 10))
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0700); err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	prefix := filepath.Join(tmp, "page")
	if err := exec.Command("pdftoppm", "-r", "144", "-png", input, prefix).Run(); err != nil {
		return err
	}
	code := `
import glob, json, os, sys
from PIL import Image, ImageDraw
from reportlab.pdfgen import canvas
from reportlab.lib.utils import ImageReader
tmp, outp, raw = sys.argv[1], sys.argv[2], sys.argv[3]
regions = json.loads(raw)
pages = sorted(glob.glob(os.path.join(tmp, "page-*.png")))
c = None
for idx, path in enumerate(pages, start=1):
    im = Image.open(path).convert("RGB")
    draw = ImageDraw.Draw(im)
    sx = im.width / 612.0
    sy = im.height / 792.0
    for r in regions:
        if int(r.get("page", 0)) == idx:
            x = float(r.get("x", 0)) * sx; y = float(r.get("y", 0)) * sy
            w = float(r.get("width", 0)) * sx; h = float(r.get("height", 0)) * sy
            draw.rectangle([x, y, x+w, y+h], fill="black")
    redacted_path = os.path.join(tmp, f"redacted-{idx}.png")
    im.save(redacted_path)
    if c is None: c = canvas.Canvas(outp, pagesize=(im.width, im.height))
    else: c.setPageSize((im.width, im.height))
    c.drawImage(ImageReader(redacted_path), 0, 0, width=im.width, height=im.height)
    c.showPage()
if c is None: raise SystemExit("no pages rendered")
c.save()
`
	return exec.Command("python3", "-c", code, tmp, output, string(raw)).Run()
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
            c.rect(x, y, w, h, fill=1, stroke=0)
    c.save()
    packet.seek(0)
    overlay = PdfReader(packet).pages[0]
    page.merge_page(overlay)
    writer.add_page(page)
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
inp, outp, raw = sys.argv[1], sys.argv[2], sys.argv[3]
opts = json.loads(raw)
reader = PdfReader(inp)
writer = PdfWriter()
for idx, page in enumerate(reader.pages):
    width = float(page.mediabox.width)
    height = float(page.mediabox.height)
    n = int(opts.get("start_number", 1)) + idx
    pad = int(opts.get("zero_padding", 0))
    body = str(n).zfill(pad) if pad > 0 else str(n)
    label = (opts.get("prefix", "") + body + opts.get("suffix", "")).strip()
    packet = io.BytesIO()
    c = canvas.Canvas(packet, pagesize=(width, height))
    c.setFont("Helvetica", 9)
    c.drawRightString(width - 24, 18, label)
    c.save()
    packet.seek(0)
    overlay = PdfReader(packet).pages[0]
    page.merge_page(overlay)
    writer.add_page(page)
with open(outp, "wb") as f:
    writer.write(f)
`
	return exec.Command("python3", "-c", code, input, output, string(raw)).Run()
}

func copyWithProcessManifest(input string, output string, mode string, details map[string]interface{}) (PDFProcessResult, error) {
	raw, err := os.ReadFile(input)
	if err != nil {
		return PDFProcessResult{}, err
	}
	if err := os.WriteFile(output, raw, 0640); err != nil {
		return PDFProcessResult{}, err
	}
	return PDFProcessResult{Mode: mode, Details: "fallback retained original PDF because strict processing dependencies were unavailable"}, nil
}
