package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const cliSample = `---
front: matter
---
# Intro

intro text

---

## Result

keep me

## Other

drop me
`

var cliBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "md-pdf-cli")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	cliBinary = filepath.Join(dir, "md-pdf")
	build := exec.Command("go", "build", "-o", cliBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	os.Exit(m.Run())
}

func runCLI(t *testing.T, stdin string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command(cliBinary, args...)
	cmd.Stdin = strings.NewReader(stdin)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %v failed: %v\nstderr: %s", cliBinary, args, err, stderr.String())
	}
	return out
}

func TestCLIReadsStdinAndWritesPDFToStdout(t *testing.T) {
	out := runCLI(t, cliSample)

	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatal("stdout is not a PDF")
	}
	if n := bytes.Count(out, []byte("/Type /Page\n")); n != 2 {
		t.Errorf("PDF has %d pages, want 2 (page 1 and page 2, page 0 excluded)", n)
	}
}

func TestCLIWritesOutputFile(t *testing.T) {
	in := filepath.Join(t.TempDir(), "in.md")
	out := filepath.Join(t.TempDir(), "out.pdf")
	if err := os.WriteFile(in, []byte(cliSample), 0o644); err != nil {
		t.Fatal(err)
	}

	runCLI(t, "", "--in", in, "--out", out)

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("output file is not a PDF")
	}
}

func TestCLIFlagsChangeTheDocument(t *testing.T) {
	full := runCLI(t, cliSample)
	onePage := runCLI(t, cliSample, "--pages", "2")
	filtered := runCLI(t, cliSample, "--heads", "h2:result", "--root-head-hide")

	if n := bytes.Count(onePage, []byte("/Type /Page\n")); n != 1 {
		t.Errorf("--pages 2 produced %d pages, want 1", n)
	}
	if len(filtered) >= len(full) {
		t.Errorf("--heads did not shrink the document: %d vs %d bytes", len(filtered), len(full))
	}
	if !bytes.HasPrefix(filtered, []byte("%PDF-")) {
		t.Error("filtered output is not a PDF")
	}
}

func TestCLIVersion(t *testing.T) {
	if got := strings.TrimSpace(string(runCLI(t, "", "--version"))); got == "" {
		t.Error("--version printed nothing")
	}
}

func TestCLIRejectsBadPages(t *testing.T) {
	cmd := exec.Command(cliBinary, "--pages", "5-1")
	cmd.Stdin = strings.NewReader(cliSample)

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected a non-zero exit for an invalid range")
	}
	if !strings.Contains(string(out), "5-1") {
		t.Errorf("error message does not mention the bad range: %s", out)
	}
}
