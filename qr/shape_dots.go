package qr

import (
	"math"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// dotFunc builds the figure for one module. It may claim neighbouring modules
// through ShapeContext.Consume, which is how the stripe shapes merge runs.
type dotFunc func(c ShapeContext, x, y int) render.Path

var dotFuncs = map[DotType]dotFunc{
	DotSquare:       dotSquare,
	DotDot:          dotDot,
	DotDotSmall:     dotDotSmall,
	DotTile:         dotTile,
	DotRounded:      dotRounded,
	DotDiamond:      dotDiamond,
	DotStar:         dotStar,
	DotStripe:       dotStripe,
	DotStripeRow:    dotStripeRow,
	DotStripeColumn: dotStripeColumn,
}

// dotPath returns the path for one module, falling back to a plain square for
// any shape not yet implemented. New rejects unimplemented shapes before this
// is reached, so the fallback is defence rather than behaviour.
func dotPath(t DotType, c ShapeContext, x, y int) render.Path {
	f, ok := dotFuncs[t]
	if !ok {
		f = dotSquare
	}
	return f(c, x, y)
}

// centredSquare returns the top-left corner and side of a square occupying
// `fraction` of module (x, y) and centred in it.
func centredSquare(c ShapeContext, x, y int, fraction float64) (ox, oy, side float64) {
	px, py, s := c.Rect(x, y)
	side = fraction * s
	return px + (s-side)/2, py + (s-side)/2, side
}

// circleIn returns a circle filling `fraction` of the module. A rounded
// rectangle whose radius is half its own side is exactly a circle, so the whole
// catalogue needs only one primitive.
func circleIn(c ShapeContext, x, y int, fraction float64) render.Path {
	ox, oy, side := centredSquare(c, x, y, fraction)
	r := side / 2
	return render.RoundRect(ox, oy, side, side, r, r, r, r)
}

// dotSquare fills the module completely. It is the default shape and the one
// with the best scanning margin, so it is also the control every other shape is
// measured against.
func dotSquare(c ShapeContext, x, y int) render.Path {
	px, py, s := c.Rect(x, y)
	c.Consume(x, y)
	return render.RoundRect(px, py, s, s, 0, 0, 0, 0)
}

// dotDot draws a circle of diameter 0.8s.
func dotDot(c ShapeContext, x, y int) render.Path {
	c.Consume(x, y)
	return circleIn(c, x, y, 0.8)
}

// dotDotSmall is dotDot at 0.6s, for a sparser, lighter texture.
func dotDotSmall(c ShapeContext, x, y int) render.Path {
	c.Consume(x, y)
	return circleIn(c, x, y, 0.6)
}

// dotTile fills the module less one pixel, so adjacent modules read as separate
// tiles instead of merging into a solid block.
func dotTile(c ShapeContext, x, y int) render.Path {
	px, py, s := c.Rect(x, y)
	c.Consume(x, y)
	return render.RoundRect(px, py, s-1, s-1, 0, 0, 0, 0)
}

// dotRounded is a square at 0.75s with a radius of a quarter of its own side.
func dotRounded(c ShapeContext, x, y int) render.Path {
	c.Consume(x, y)
	ox, oy, side := centredSquare(c, x, y, 0.75)
	r := side / 4
	return render.RoundRect(ox, oy, side, side, r, r, r, r)
}

// dotDiamond is a square rotated 45 degrees. Its side is scaled by
// 1/sin(45 degrees) so the rotated figure spans exactly one module corner to
// corner rather than overflowing it.
func dotDiamond(c ShapeContext, x, y int) render.Path {
	px, py, s := c.Rect(x, y)
	c.Consume(x, y)
	ox, oy, side := centredSquare(c, x, y, 0.5/math.Sin(math.Pi/4))
	return render.RoundRect(ox, oy, side, side, 0, 0, 0, 0).
		Rotate(math.Pi/4, px+s/2, py+s/2)
}

// dotStar runs four quadratic curves between the module's corners with their
// control point at the centre, pulling each edge inward, then rotates the
// figure 45 degrees so the points face the axes.
//
// Unlike dotDiamond this is not scaled down, so the points reach s/sqrt(2) from
// the centre and overlap slightly into neighbouring modules. That overlap is
// what the reference library draws and what gives the shape its look.
func dotStar(c ShapeContext, x, y int) render.Path {
	px, py, s := c.Rect(x, y)
	c.Consume(x, y)
	cx, cy := px+s/2, py+s/2
	h := s / 2

	var b render.Builder
	b.MoveTo(cx-h, cy-h)
	b.QuadTo(cx, cy, cx+h, cy-h)
	b.QuadTo(cx, cy, cx+h, cy+h)
	b.QuadTo(cx, cy, cx-h, cy+h)
	b.QuadTo(cx, cy, cx-h, cy-h)
	b.Close()
	return b.Path().Rotate(math.Pi/4, cx, cy)
}

// run is a rectangle of modules a stripe shape may merge into one figure,
// measured in modules.
type run struct{ w, h int }

// The greedy orders the reference library uses. The first run that fits wins,
// so listing the longest first is what makes runs merge rather than fragment.
var (
	stripeRuns       = []run{{3, 1}, {1, 3}, {2, 1}, {1, 2}, {1, 1}}
	stripeRowRuns    = []run{{3, 1}, {2, 1}, {1, 1}}
	stripeColumnRuns = []run{{1, 3}, {1, 2}, {1, 1}}
)

func dotStripe(c ShapeContext, x, y int) render.Path {
	return mergeRun(c, x, y, stripeRuns)
}

func dotStripeRow(c ShapeContext, x, y int) render.Path {
	return mergeRun(c, x, y, stripeRowRuns)
}

func dotStripeColumn(c ShapeContext, x, y int) render.Path {
	return mergeRun(c, x, y, stripeColumnRuns)
}

// mergeRun claims the first run in order that is entirely available, and draws
// it as one capsule.
//
// This is why ShapeContext is a context and not a neighbour mask: the shape
// reads coordinates it was not asked about and writes consumption state back,
// so the main loop skips the modules it swallowed. A read-only 3x3 snapshot
// cannot express either half of that.
func mergeRun(c ShapeContext, x, y int, order []run) render.Path {
	for _, r := range order {
		if !runAvailable(c, x, y, r) {
			continue
		}
		for dy := 0; dy < r.h; dy++ {
			for dx := 0; dx < r.w; dx++ {
				c.Consume(x+dx, y+dy)
			}
		}
		return capsule(c, x, y, r)
	}
	// Unreachable in practice: the caller only invokes a dot shape for a module
	// that is available, and every order ends with the 1x1 run. Returning an
	// empty path rather than panicking keeps the no-panic rule intact.
	return render.Path{}
}

// runAvailable reports whether every module of the run may be claimed. It asks
// ShapeContext.Dark, which is false for a module that is out of bounds, light,
// already consumed, part of a finder pattern or hidden by the logo — so a run
// can never grow into a region it must not touch. The reference library checks
// only darkness here, which is how its runs leak across finder patterns.
func runAvailable(c ShapeContext, x, y int, r run) bool {
	for dy := 0; dy < r.h; dy++ {
		for dx := 0; dx < r.w; dx++ {
			if !c.Dark(x+dx, y+dy) {
				return false
			}
		}
	}
	return true
}

// capsule draws a run as a rounded bar of thickness s/2 running between the
// centres of its first and last modules. A 1x1 run degenerates to a circle of
// diameter s/2, so all three cases share one formula.
func capsule(c ShapeContext, x, y int, r run) render.Path {
	px, py, s := c.Rect(x, y)
	thickness := s / 2
	radius := s / 4
	return render.RoundRect(
		px+radius,
		py+radius,
		float64(r.w-1)*s+thickness,
		float64(r.h-1)*s+thickness,
		radius, radius, radius, radius,
	)
}
