package qr

import "github.com/farizfadian/go-qrcode/internal/render"

// cornerFunc builds one finder pattern. px and py are the pixel position of the
// pattern's top-left module and s is the module size, so the figure spans 7s.
type cornerFunc func(px, py, s float64, r CornerRadius) render.Path

var cornerFuncs = map[CornerType]cornerFunc{
	CornerSquare: cornerSquare,
}

// cornerPath returns the path for one finder pattern, falling back to a plain
// square for any shape not yet implemented.
func cornerPath(t CornerType, px, py, s float64, r CornerRadius) render.Path {
	f, ok := cornerFuncs[t]
	if !ok {
		f = cornerSquare
	}
	return f(px, py, s, r)
}

// ring returns a one-module-thick frame as an outer contour plus a reversed
// inner contour, which the non-zero fill rule turns into a hole.
//
// The reference library strokes a centreline figure with lineWidth = s.
// Offsetting a rounded rectangle outward by d grows its side by 2d and its
// radius by d, so a centreline side of 6s becomes a 7s outer contour and a 5s
// inner one. The radius rule applies only where centreRadius is positive: at
// zero the centreline has a real corner join and the canvas default lineJoin is
// miter, so the outline stays sharp and no radius appears.
func ring(px, py, s, centreRadius float64) render.Path {
	outer, inner := 0.0, 0.0
	if centreRadius > 0 {
		outer = centreRadius + s/2
		inner = centreRadius - s/2
		if inner < 0 {
			inner = 0
		}
	}
	out := render.RoundRect(px, py, 7*s, 7*s, outer, outer, outer, outer)
	hole := render.RoundRect(px+s, py+s, 5*s, 5*s, inner, inner, inner, inner).Reverse()
	return out.Append(hole)
}

// cornerSquare draws the unstyled finder pattern: a sharp-cornered ring around
// a sharp-cornered core.
func cornerSquare(px, py, s float64, _ CornerRadius) render.Path {
	core := render.RoundRect(px+2*s, py+2*s, 3*s, 3*s, 0, 0, 0, 0)
	return ring(px, py, s, 0).Append(core)
}
