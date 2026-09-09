package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dimkarp93/install-libs/buildinfo"
	mdlib "github.com/dimkarp93/md-libs"
	"github.com/dimkarp93/md-libs/render"
)

var (
	version  string
	origin   string
	upstream string
	commit   string
	channel  string
)

func build() buildinfo.Info {
	return buildinfo.Info{
		Version:  version,
		Origin:   origin,
		Upstream: upstream,
		Commit:   commit,
		Channel:  channel,
	}
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.BoolVar(showVersion, "v", false, "print version and exit")
	showOrigin := flag.Bool("origin", false, "print the repository the binary was built from and exit")
	showBuildInfo := flag.Bool("buildinfo", false, "print build metadata and exit")
	inPath := flag.String("in", "", "input markdown file (default: stdin)")
	outPath := flag.String("out", "", "output pdf file (default: stdout)")
	pagesFlag := flag.String("pages", "", "pages to include, e.g. 1,3-5 (default: all pages except 0)")
	headsFlag := flag.String("heads", "", "heading filters, e.g. h2:result,h3:resume")
	rootHeadHide := flag.Bool("root-head-hide", false, "hide headings matched by --heads, keep their content")
	flag.Parse()

	switch {
	case *showVersion:
		fmt.Println(build().VersionString())
		return
	case *showOrigin:
		fmt.Println(build().OriginString())
		return
	case *showBuildInfo:
		build().Print()
		return
	}

	var data []byte
	var err error
	if *inPath == "" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(*inPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	blocks, err := mdlib.Prepare(string(data), mdlib.Options{
		Pages:            *pagesFlag,
		Heads:            *headsFlag,
		HideMatchedHeads: *rootHeadHide,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	r := newPDFRenderer()
	render.Blocks(r, blocks)

	out := io.Writer(os.Stdout)
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", *outPath, err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	if err := r.Output(out); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}
