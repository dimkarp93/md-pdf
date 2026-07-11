package main

import (
	"strings"

	"github.com/go-pdf/fpdf"
)

const (
	fontFamily       = "DejaVu"
	bodySizePt       = 11.0
	codeSizePt       = 10.0
	mmPerPt          = 0.3528
	lineHeightFactor = 0.42
)

func headingSizePt(level int) float64 {
	switch level {
	case 1:
		return 18
	case 2:
		return 16
	case 3:
		return 14
	case 4:
		return 13
	case 5:
		return 12
	default:
		return 11
	}
}

func lineHeightMM(sizePt float64) float64 {
	return sizePt * lineHeightFactor
}

func styleFor(bold, italic bool) string {
	s := ""
	if bold {
		s += "B"
	}
	if italic {
		s += "I"
	}
	return s
}

func newPDF() *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(fontFamily, "", fontRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", fontBold)
	pdf.AddUTF8FontFromBytes(fontFamily, "I", fontItalic)
	pdf.AddUTF8FontFromBytes(fontFamily, "BI", fontBoldItalic)
	pdf.SetMargins(25, 25, 20)
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()
	return pdf
}

func writeRuns(pdf *fpdf.Fpdf, runs []run, sizePt float64, forceBold bool) {
	h := lineHeightMM(sizePt)
	for _, r := range runs {
		pdf.SetFont(fontFamily, styleFor(r.bold || forceBold, r.italic), sizePt)
		pdf.Write(h, r.text)
	}
}

func renderHeadingPDF(pdf *fpdf.Fpdf, level int, text string) {
	size := headingSizePt(level)
	h := lineHeightMM(size)
	pdf.Ln(12 * mmPerPt)
	writeRuns(pdf, parseInline(text, false), size, true)
	pdf.Ln(h)
	pdf.Ln(6 * mmPerPt)
}

func renderParagraphPDF(pdf *fpdf.Fpdf, text string) {
	h := lineHeightMM(bodySizePt)
	for _, line := range strings.Split(text, "\n") {
		writeRuns(pdf, parseInline(line, true), bodySizePt, false)
		pdf.Ln(h)
	}
	pdf.Ln(6 * mmPerPt)
}

func renderCodeLinePDF(pdf *fpdf.Fpdf, line string) {
	text := strings.ReplaceAll(line, "\t", "    ")
	if text == "" {
		text = " "
	}
	h := lineHeightMM(codeSizePt)
	pdf.SetFont(fontFamily, "", codeSizePt)
	pdf.SetFillColor(242, 242, 242)
	pdf.MultiCell(0, h, text, "", "L", true)
}

func renderPDFBody(pdf *fpdf.Fpdf, blocks []block) {
	for _, b := range blocks {
		switch b.kind {
		case blockHeading:
			renderHeadingPDF(pdf, b.level, b.text)
		case blockCode:
			for _, line := range b.lines {
				renderCodeLinePDF(pdf, line)
			}
		case blockPageBreak:
			pdf.AddPage()
		default:
			renderParagraphPDF(pdf, b.text)
		}
	}
}
