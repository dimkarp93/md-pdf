package main

import (
	"io"
	"strings"

	"github.com/dimkarp93/md-libs/markdown"
	"github.com/dimkarp93/md-libs/render"
	"github.com/go-pdf/fpdf"
)

const (
	fontFamily       = "DejaVu"
	bodySizePt       = 11.0
	codeSizePt       = 10.0
	mmPerPt          = 0.3528
	lineHeightFactor = 0.42
)

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

type pdfRenderer struct {
	pdf *fpdf.Fpdf
}

func newPDFRenderer() *pdfRenderer {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(fontFamily, "", fontRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", fontBold)
	pdf.AddUTF8FontFromBytes(fontFamily, "I", fontItalic)
	pdf.AddUTF8FontFromBytes(fontFamily, "BI", fontBoldItalic)
	pdf.SetMargins(25, 25, 20)
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()
	return &pdfRenderer{pdf: pdf}
}

func (p *pdfRenderer) writeRuns(runs []markdown.Run, sizePt float64, forceBold bool) {
	h := lineHeightMM(sizePt)
	for _, r := range runs {
		p.pdf.SetFont(fontFamily, styleFor(r.Bold || forceBold, r.Italic), sizePt)
		p.pdf.Write(h, r.Text)
	}
}

func (p *pdfRenderer) Heading(level int, text string) {
	size := render.HeadingSizePt(level)
	h := lineHeightMM(size)
	p.pdf.Ln(12 * mmPerPt)
	p.writeRuns(markdown.ParseInline(text, false), size, true)
	p.pdf.Ln(h)
	p.pdf.Ln(6 * mmPerPt)
}

func (p *pdfRenderer) Paragraph(text string) {
	h := lineHeightMM(bodySizePt)
	for _, line := range strings.Split(text, "\n") {
		p.writeRuns(markdown.ParseInline(line, true), bodySizePt, false)
		p.pdf.Ln(h)
	}
	p.pdf.Ln(6 * mmPerPt)
}

func (p *pdfRenderer) Code(lines []string) {
	h := lineHeightMM(codeSizePt)
	for _, line := range lines {
		text := strings.ReplaceAll(line, "\t", "    ")
		if text == "" {
			text = " "
		}
		p.pdf.SetFont(fontFamily, "", codeSizePt)
		p.pdf.SetFillColor(242, 242, 242)
		p.pdf.MultiCell(0, h, text, "", "L", true)
	}
}

func (p *pdfRenderer) PageBreak() {
	p.pdf.AddPage()
}

func (p *pdfRenderer) Output(w io.Writer) error {
	return p.pdf.Output(w)
}
