# md-pdf

A simple utility that converts a Markdown file into `.pdf`. It is built exactly like [md-docx](../md-docx), except that instead of a `.docx` it assembles a PDF through the [`github.com/go-pdf/fpdf`](https://github.com/go-pdf/fpdf) library (the only external dependency — generating PDF bytecode by hand on top of the stdlib is too laborious and fragile).

To support Cyrillic it uses the DejaVu Sans Condensed font embedded via `go:embed` — the standard built-in PDF fonts (Helvetica/Courier) do not support Cyrillic. The font is baked into the binary, so no network access is required at runtime.

## Supported syntax

1. Headings: `#` … `######` (levels 1–6).
2. Bold: `**text**` or `__text__`.
3. Italic: `*text*` or `_text_` (inside headings `_` is not treated as markup).
4. Bold italic: `***text***`.
5. Code blocks: `` ```...``` `` — rendered in the same font (there is no monospace variant with Cyrillic in the bundle) with a grey fill; long lines wrap to the page width and tabs are replaced with spaces.

Everything else is treated as a regular text paragraph. Lines with no blank line between them are merged into a single paragraph with a line break inside; a blank line in the source produces a visible gap between paragraphs.

## Build

Requires Go 1.26+.

```bash
make build
```

The binary appears in the repository root — `./md-pdf`.

## Usage

```bash
./md-pdf --in input.md --out output.pdf
```

- `--in` (optional) — path to the source Markdown file. If omitted, reads from stdin.
- `--out` (optional) — path to the output `.pdf`. If omitted, writes to stdout.
- `--pages` (optional) — which pages to include, for example `1,3-5`. By default — every page except page 0.
- `--heads` (optional) — a heading filter, for example `h2:result,h3:resume,summary`. By default there is no filtering.
- `--root-head-hide` — hides the headings matched by `--heads`, keeping their content.
- `--version` / `-v` — prints the version and exits.

stdin/stdout are supported as well:

```bash
cat input.md | ./md-pdf > output.pdf
```

### Frontmatter, pages and the heading filter

The semantics are exactly the same as in [md-docx](../md-docx/README.md#frontmatter-and-pages):

- If the file starts with a `---` line, everything up to the next `---` is frontmatter (page 0) and does not make it into the result unless requested explicitly through `--pages=0`.
- The other `---` separator lines (outside code blocks) split the document into pages 1, 2, 3, … By default every page except page 0 is emitted; specific pages are selected with `--pages=1,3-5`.
- Between every two adjacent selected pages a real page break is inserted into the PDF.
- `--heads=h2:result,resume` keeps only the content under the given headings (and their nested content); nesting is computed across the selected document as a whole, through page breaks. A name without a level prefix (`resume`) matches a heading at any level.

## Clean

```bash
make clean
```

Removes the built binary.

## Development

Markdown parsing, heading filtering and the rendering contract live in the shared library [md-libs](https://github.com/dimkarp93/md-libs) — the same library is used by [md-docx](https://github.com/dimkarp93/md-docx). What remains in this repository is only the PDF rendering (fpdf + the embedded DejaVu fonts) and the flag parsing.

The dependencies, md-libs included, are vendored: `make build` and the tests take them from `vendor/` only (the Makefile exports `GOWORK=off` and `GOFLAGS=-mod=vendor`), nothing has to be set up and no network is needed:

```bash
make build
```

To change the library and the CLI at the same time, publish a new version of md-libs first and then update the dependency here (see below).

### Tests

```sh
make test                  # renderer tests and end-to-end CLI tests
make test-v                # the same, with test names
make test-run T=TestCLIVersion
make cover                 # coverage
make check                 # vet + test
```

The shared core (parsing, filters, pipeline) is tested in [md-libs](https://github.com/dimkarp93/md-libs); only what is specific to this CLI is checked here. `make test-all` in md-libs runs everything at once.

Upgrading to a new version of md-libs:

```bash
GOWORK=off go get github.com/dimkarp93/md-libs@v0.1.1
make vendor
```

`make vendor-check` verifies that `vendor/` matches `go.mod`.

## License

[MIT](LICENSE)

The bundled DejaVu Sans Condensed fonts (`cmd/md-pdf/fonts/`) are distributed under the [DejaVu Fonts License](https://dejavu-fonts.github.io/License.html).
