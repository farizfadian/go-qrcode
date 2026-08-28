package qr

import "github.com/farizfadian/go-qrcode/internal/render"

// dotFunc builds the figure for one module. It may claim neighbouring modules
// through ShapeContext.Consume, which is how the stripe shapes merge runs.
type dotFunc func(c ShapeContext, x, y int) render.Path

var dotFuncs = map[DotType]dotFunc{
	DotSquare:       dotSquare,
	DotStripe:       dotStripe,
	DotStripeRow:    dotStripeRow,
	DotStripeColumn: dotStripeColumn,
}

// dotPath returns the path for one module, falling back to a plain square for
// any shape not yet implemented.
func dotPath(t DotType, c ShapeContext, x, y int) render.Path {
	f, ok := dotFuncs[t]
	if !ok {
		f = dotSquare
	}
	return f(c, x, y)
}

// dotSquare fills the module completely. It is the default shape and the one
// with the best scanning margin, so it is also the control every other shape is
// measured against.
func dotSquare(c ShapeContext, x, y int) render.Path {
	px, py, s := c.Rect(x, y)
	c.Consume(x, y)
	return render.RoundRect(px, py, s, s, 0, 0, 0, 0)
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
