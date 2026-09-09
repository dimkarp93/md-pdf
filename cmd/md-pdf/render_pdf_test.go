package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	mdlib "github.com/dimkarp93/md-libs"
	"github.com/dimkarp93/md-libs/markdown"
	"github.com/dimkarp93/md-libs/render"
)

func TestStyleFor(t *testing.T) {
	tests := []struct {
		bold, italic bool
		want         string
	}{
		{false, false, ""},
		{true, false, "B"},
		{false, true, "I"},
		{true, true, "BI"},
	}
	for _, tt := range tests {
		if got := styleFor(tt.bold, tt.italic); got != tt.want {
			t.Errorf("styleFor(%v, %v) = %q, want %q", tt.bold, tt.italic, got, tt.want)
		}
	}
}

func TestLineHeightGrowsWithFontSize(t *testing.T) {
	if lineHeightMM(bodySizePt) >= lineHeightMM(render.HeadingSizePt(1)) {
		t.Error("a heading line must be taller than a body line")
	}
	if got, want := lineHeightMM(10), 10*lineHeightFactor; got != want {
		t.Errorf("lineHeightMM(10) = %v, want %v", got, want)
	}
}

func TestNewPDFRendererStartsWithOnePage(t *testing.T) {
	r := newPDFRenderer()
	if got := r.pdf.PageCount(); got != 1 {
		t.Errorf("newPDFRenderer() page count = %d, want 1", got)
	}
	if err := r.pdf.Error(); err != nil {
		t.Fatalf("newPDFRenderer() failed to load fonts: %v", err)
	}
}

func TestPageBreakAddsAPage(t *testing.T) {
	r := newPDFRenderer()
	r.Paragraph("first")
	r.PageBreak()
	r.Paragraph("second")

	if got := r.pdf.PageCount(); got != 2 {
		t.Errorf("page count = %d, want 2", got)
	}
	if err := r.pdf.Error(); err != nil {
		t.Fatalf("render error: %v", err)
	}
}

func TestRendersEveryBlockKindWithoutError(t *testing.T) {
	r := newPDFRenderer()
	render.Blocks(r, []markdown.Block{
		{Kind: markdown.Heading, Level: 1, Text: "Заголовок **жирный**"},
		{Kind: markdown.Paragraph, Text: "текст с _курсивом_\nи второй строкой"},
		{Kind: markdown.Code, Lines: []string{"func main() {", "\tprintln(1)", "", "}"}},
		{Kind: markdown.PageBreak},
		{Kind: markdown.Paragraph, Text: "после разрыва"},
	})

	if err := r.pdf.Error(); err != nil {
		t.Fatalf("render error: %v", err)
	}
	if got := r.pdf.PageCount(); got != 2 {
		t.Errorf("page count = %d, want 2", got)
	}

	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatalf("Output() unexpected error: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("output is not a PDF")
	}
}

func TestHeadingUsesTheSharedScale(t *testing.T) {
	for level := 1; level <= 6; level++ {
		r := newPDFRenderer()
		r.Heading(level, "Title")
		size, _ := r.pdf.GetFontSize()
		if want := render.HeadingSizePt(level); size != want {
			t.Errorf("Heading(%d) left font size %v, want %v", level, size, want)
		}
	}
}

func renderToPDF(t *testing.T, src string, opts mdlib.Options) []byte {
	t.Helper()

	blocks, err := mdlib.Prepare(src, opts)
	if err != nil {
		t.Fatalf("Prepare() unexpected error: %v", err)
	}

	r := newPDFRenderer()
	epoch := time.Unix(0, 0).UTC()
	r.pdf.SetCreationDate(epoch)
	r.pdf.SetModificationDate(epoch)
	render.Blocks(r, blocks)

	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatalf("Output() unexpected error: %v", err)
	}
	return buf.Bytes()
}

func TestOutputIsStableAcrossRenders(t *testing.T) {
	src := "# Title\n\nbody\n"
	first := renderToPDF(t, src, mdlib.Options{})
	second := renderToPDF(t, src, mdlib.Options{})

	if len(first) != len(second) {
		t.Errorf("two renders of the same input differ in size: %d vs %d bytes", len(first), len(second))
	}
}

func TestMarkdownToPDFEndToEnd(t *testing.T) {
	src := "---\nfront: matter\n---\n# Title\n\nfirst page\n\n---\n\n## Result\n\nsecond page\n"

	full := renderToPDF(t, src, mdlib.Options{})
	if !bytes.HasPrefix(full, []byte("%PDF-")) {
		t.Fatal("output is not a PDF")
	}

	blocks, err := mdlib.Prepare(src, mdlib.Options{Pages: "2", Heads: "h2:result", HideMatchedHeads: true})
	if err != nil {
		t.Fatalf("Prepare() unexpected error: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Text != "second page" {
		t.Fatalf("filters did not reach the renderer: %#v", blocks)
	}

	filtered := renderToPDF(t, src, mdlib.Options{Pages: "2", Heads: "h2:result", HideMatchedHeads: true})
	if len(filtered) >= len(full) {
		t.Errorf("filtered PDF (%d bytes) is not smaller than the full one (%d bytes)", len(filtered), len(full))
	}
}

func TestCodeExpandsTabs(t *testing.T) {
	r := newPDFRenderer()
	r.Code([]string{"\tindented"})
	if err := r.pdf.Error(); err != nil {
		t.Fatalf("render error: %v", err)
	}

	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatalf("Output() unexpected error: %v", err)
	}
	if strings.Contains(buf.String(), "\tindented") {
		t.Error("tabs must be expanded to spaces before rendering")
	}
}
