package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"image/png"
	"strconv"
	"strings"
)

// SVG serialises sc to standalone SVG markup. It emits only the absolute path
// commands M, L, Q, C and Z, and relies on the default non-zero fill rule so
// reversed subpaths punch holes exactly as they do in the rasteriser.
func SVG(sc Scene) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" `+
		`viewBox="0 0 %d %d" width="%d" height="%d">`,
		sc.Width, sc.Height, sc.Width, sc.Height)

	if sc.Background != nil {
		if _, _, _, a := sc.Background.RGBA(); a > 0 {
			fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="%s"/>`,
				sc.Width, sc.Height, hexOf(sc.Background))
		}
	}

	clipID := 0
	for _, it := range sc.Items {
		switch v := it.(type) {
		case PathItem:
			if v.Path.IsEmpty() {
				continue
			}
			fmt.Fprintf(&b, `<path d="%s" fill="%s"/>`, pathD(v.Path), hexOf(v.Fill))
		case ImageItem:
			data, err := dataURI(v)
			if err != nil {
				return "", err
			}
			attr := ""
			if v.Clip != nil {
				clipID++
				id := "clip" + strconv.Itoa(clipID)
				fmt.Fprintf(&b, `<clipPath id="%s"><path d="%s"/></clipPath>`, id, pathD(*v.Clip))
				attr = fmt.Sprintf(` clip-path="url(#%s)"`, id)
			}
			fmt.Fprintf(&b, `<image x="%s" y="%s" width="%s" height="%s" href="%s"%s/>`,
				num(v.X), num(v.Y), num(v.W), num(v.H), data, attr)
		}
	}
	b.WriteString(`</svg>`)
	return b.String(), nil
}

// pathD renders a path as an SVG d attribute.
func pathD(p Path) string {
	var b strings.Builder
	for _, sp := range p.SubPaths {
		fmt.Fprintf(&b, "M %s %s", num(sp.Start.X), num(sp.Start.Y))
		for _, s := range sp.Segs {
			switch s.Kind {
			case SegLine:
				fmt.Fprintf(&b, " L %s %s", num(s.To.X), num(s.To.Y))
			case SegQuad:
				fmt.Fprintf(&b, " Q %s %s %s %s",
					num(s.C1.X), num(s.C1.Y), num(s.To.X), num(s.To.Y))
			case SegCube:
				fmt.Fprintf(&b, " C %s %s %s %s %s %s",
					num(s.C1.X), num(s.C1.Y), num(s.C2.X), num(s.C2.Y),
					num(s.To.X), num(s.To.Y))
			}
		}
		if sp.Closed {
			b.WriteString(" Z")
		}
	}
	return b.String()
}

// num formats a coordinate with three decimals and no trailing zeros, which
// keeps the markup small while staying well inside the tolerance the parity
// test compares against.
func num(f float64) string {
	s := strconv.FormatFloat(f, 'f', 3, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	return s
}

// hexOf renders a colour as #rrggbb, undoing Go's premultiplication so the hex
// matches what the caller supplied. A fully transparent colour becomes "none".
func hexOf(c color.Color) string {
	r, g, b, a := c.RGBA()
	if a == 0 {
		return "none"
	}
	r, g, b = r*0xffff/a, g*0xffff/a, b*0xffff/a
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func dataURI(it ImageItem) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, it.Img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
