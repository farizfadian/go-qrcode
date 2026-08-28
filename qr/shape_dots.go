package qr

import "github.com/farizfadian/go-qrcode/internal/render"

// dotFunc builds the figure for one module. It may claim neighbouring modules
// through ShapeContext.Consume, which is how the stripe shapes merge runs.
type dotFunc func(c ShapeContext, x, y int) render.Path

var dotFuncs = map[DotType]dotFunc{
	DotSquare: dotSquare,
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
