package main

import _ "embed"

//go:embed fonts/DejaVuSansCondensed.ttf
var fontRegular []byte

//go:embed fonts/DejaVuSansCondensed-Bold.ttf
var fontBold []byte

//go:embed fonts/DejaVuSansCondensed-Oblique.ttf
var fontItalic []byte

//go:embed fonts/DejaVuSansCondensed-BoldOblique.ttf
var fontBoldItalic []byte
