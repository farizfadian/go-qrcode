package qr

import (
	"math"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// cornerFunc builds one finder pattern. px and py are the pixel position of the
// pattern's top-left module and s is the module size, so the figure spans 7s.
type cornerFunc func(px, py, s float64, r CornerRadius) render.Path

var cornerFuncs = map[CornerType]cornerFunc{
	CornerSquare:        cornerSquare,
	CornerRounded:       cornerRounded,
	CornerCircle:        cornerCircle,
	CornerRoundedCircle: cornerRoundedCircle,
	CornerCircleRounded: cornerCircleRounded,
	CornerCircleStar:    cornerCircleStar,
	CornerCircleDiamond: cornerCircleDiamond,
}

// cornerPath returns the path for one finder pattern, falling back to a plain
// square for any shape not yet implemented. New rejects unimplemented shapes
// before this is reached, so the fallback is defence rather than behaviour.
func cornerPath(t CornerType, px, py, s float64, r CornerRadius) render.Path {
	f, ok := cornerFuncs[t]
	if !ok {
		f = cornerSquare
	}
	return f(px, py, s, r)
}

// resolve fills in the defaults the reference library uses: half a module for
// the outer radius, a quarter for the inner.
func (r CornerRadius) resolve(s float64) (outer, inner float64) {
	outer, inner = r.Outer, r.Inner
	if outer == 0 {
		outer = s / 2
	}
	if inner == 0 {
		inner = s / 4
	}
	return outer, inner
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

// circleRing is the same one-module frame drawn as a circle: the stroked
// centreline of radius 3s becomes an outer circle of 3.5s and an inner one of
// 2.5s.
func circleRing(px, py, s float64) render.Path {
	cx, cy := px+3.5*s, py+3.5*s
	return render.Circle(cx, cy, 3.5*s).
		Append(render.Circle(cx, cy, 2.5*s).Reverse())
}

// coreRect is the solid 3x3 centre. The reference strokes and fills it, so the
// union is a figure of side 3s whose radius is half a module larger than the
// caller asked for.
func coreRect(px, py, s, radius float64) render.Path {
	return render.RoundRect(px+2*s, py+2*s, 3*s, 3*s, radius, radius, radius, radius)
}

// coreDisc is the solid 3x3 centre drawn as a circle.
func coreDisc(px, py, s float64) render.Path {
	return render.Circle(px+3.5*s, py+3.5*s, 1.5*s)
}

// coreDiamond is a square rotated 45 degrees and inscribed in the 3x3 centre,
// so its diagonal is exactly three modules.
//
// This deviates from the reference library, deliberately. The reference rotates
// a full 3s square, giving a diagonal of 3*sqrt(2) = 4.24 modules that nearly
// fills the ring's 5-module gap. A reader looks for dark:light:dark:light:dark
// in the ratio 1:1:3:1:1 through the finder's centre; the reference geometry
// measures 1:0.38:4.24:0.38:1 and is not recognised as a finder pattern at all.
// Inscribing the diamond restores exactly 1:1:3:1:1, and it was measured: the
// reference proportions fail to decode, these succeed.
func coreDiamond(px, py, s float64) render.Path {
	side := 3 * s / math.Sqrt2
	cx, cy := px+3.5*s, py+3.5*s
	return render.RoundRect(cx-side/2, cy-side/2, side, side, 0, 0, 0, 0).
		Rotate(math.Pi/4, cx, cy)
}

// coreStar pulls each edge of the 3x3 centre inward with a quadratic curve,
// leaving a four-pointed star. Its control points sit a quarter of the way in
// from each edge rather than at the centre, so the concavity is gentler than
// the dot shape of the same name.
func coreStar(px, py, s float64) render.Path {
	cx, cy := px+3.5*s, py+3.5*s
	h := 1.5 * s // half the 3s span
	q := 0.75 * s

	var b render.Builder
	b.MoveTo(cx-h, cy-h)
	b.QuadTo(cx, cy-q, cx+h, cy-h)
	b.QuadTo(cx+q, cy, cx+h, cy+h)
	b.QuadTo(cx, cy+q, cx-h, cy+h)
	b.QuadTo(cx-q, cy, cx-h, cy-h)
	b.Close()
	return b.Path()
}

// cornerSquare draws the unstyled finder pattern: a sharp-cornered ring around
// a sharp-cornered core.
func cornerSquare(px, py, s float64, _ CornerRadius) render.Path {
	return ring(px, py, s, 0).Append(coreRect(px, py, s, 0))
}

// cornerRounded softens both the ring and the core, using the caller's radii or
// the reference defaults.
func cornerRounded(px, py, s float64, r CornerRadius) render.Path {
	outer, inner := r.resolve(s)
	return ring(px, py, s, outer).Append(coreRect(px, py, s, inner+s/2))
}

func cornerCircle(px, py, s float64, _ CornerRadius) render.Path {
	return circleRing(px, py, s).Append(coreDisc(px, py, s))
}

func cornerRoundedCircle(px, py, s float64, r CornerRadius) render.Path {
	outer, _ := r.resolve(s)
	return ring(px, py, s, outer).Append(coreDisc(px, py, s))
}

func cornerCircleRounded(px, py, s float64, r CornerRadius) render.Path {
	_, inner := r.resolve(s)
	return circleRing(px, py, s).Append(coreRect(px, py, s, inner+s/2))
}

func cornerCircleStar(px, py, s float64, _ CornerRadius) render.Path {
	return circleRing(px, py, s).Append(coreStar(px, py, s))
}

func cornerCircleDiamond(px, py, s float64, _ CornerRadius) render.Path {
	return circleRing(px, py, s).Append(coreDiamond(px, py, s))
}
