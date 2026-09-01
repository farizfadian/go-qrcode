package render

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"strings"
)

// ErrColorSyntax reports a colour string that is not hex. CSS colour names and
// functional notation such as rgb() resolve only inside a browser; no renderer
// here can measure such a colour, so they are rejected rather than passed on.
var ErrColorSyntax = errors.New("render: not a hex colour")

// NormalizeColor adds a leading '#' to a bare hex colour of 3, 4, 6 or 8
// digits. Anything else, including CSS names and functional notation, is
// returned unchanged, so ParseColor's rejection names the original input.
func NormalizeColor(s string) string {
	if s == "" || s[0] == '#' || !isHexRun(s) {
		return s
	}
	switch len(s) {
	case 3, 4, 6, 8:
		return "#" + s
	}
	return s
}

// ParseColor returns the premultiplied RGBA value of a hex colour, with or
// without a leading '#', in 3, 4, 6 or 8 digit form. It returns ErrColorSyntax
// for anything it cannot resolve.
func ParseColor(s string) (color.RGBA, error) {
	h := strings.TrimPrefix(NormalizeColor(s), "#")
	if !isHexRun(h) {
		return color.RGBA{}, fmt.Errorf("%w: %q", ErrColorSyntax, s)
	}
	var r, g, b, a uint8 = 0, 0, 0, 0xff
	switch len(h) {
	case 3:
		r, g, b = dup(h[0]), dup(h[1]), dup(h[2])
	case 4:
		r, g, b, a = dup(h[0]), dup(h[1]), dup(h[2]), dup(h[3])
	case 6:
		r, g, b = hexByte(h[0], h[1]), hexByte(h[2], h[3]), hexByte(h[4], h[5])
	case 8:
		r, g, b, a = hexByte(h[0], h[1]), hexByte(h[2], h[3]),
			hexByte(h[4], h[5]), hexByte(h[6], h[7])
	default:
		return color.RGBA{}, fmt.Errorf("%w: %q", ErrColorSyntax, s)
	}
	// color.RGBA is premultiplied; hex notation carries straight alpha.
	m := uint32(a)
	return color.RGBA{
		R: uint8(uint32(r) * m / 0xff),
		G: uint8(uint32(g) * m / 0xff),
		B: uint8(uint32(b) * m / 0xff),
		A: a,
	}, nil
}

// Luminance returns the WCAG relative luminance of a colour, from 0 for black
// to 1 for white. Alpha is ignored; composite first if that matters.
func Luminance(c color.RGBA) float64 {
	return 0.2126*linear(c.R) + 0.7152*linear(c.G) + 0.0722*linear(c.B)
}

// ContrastRatio returns the WCAG contrast ratio between two colours, from 1 for
// identical to 21 for black against white. The order of the arguments does not
// matter.
func ContrastRatio(a, b color.RGBA) float64 {
	la, lb := Luminance(a), Luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// linear converts one sRGB channel to its linear-light value.
func linear(v uint8) float64 {
	c := float64(v) / 255
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// Over composites src onto dst, both straight-alpha-free RGBA values, and
// returns the opaque result. It is how a transparent background is judged: the
// code will sit on something, and white is the usual something.
func Over(src, dst color.RGBA) color.RGBA {
	if src.A == 0xff {
		return src
	}
	a := float64(src.A) / 255
	blend := func(s, d uint8) uint8 {
		// src is premultiplied, so its colour channels already carry the alpha.
		return uint8(float64(s) + float64(d)*(1-a))
	}
	return color.RGBA{blend(src.R, dst.R), blend(src.G, dst.G), blend(src.B, dst.B), 0xff}
}

func isHexRun(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if hexVal(s[i]) < 0 {
			return false
		}
	}
	return true
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func hexByte(hi, lo byte) uint8 { return uint8(hexVal(hi)<<4 | hexVal(lo)) }

// dup expands one hex digit into a byte, so 'f' becomes 0xff.
func dup(c byte) uint8 {
	v := uint8(hexVal(c))
	return v<<4 | v
}
