package qr

// DotType selects the figure drawn for each data module.
type DotType int

// The dot shapes ported from qrcode-with-logos. DotSquare is the default and
// the zero value.
const (
	DotSquare DotType = iota
	DotDot
	DotDotSmall
	DotTile
	DotRounded
	DotDiamond
	DotStar
	DotFluid
	DotFluidLine
	DotStripe
	DotStripeRow
	DotStripeColumn
)

var dotNames = [...]string{
	"square", "dot", "dot-small", "tile", "rounded", "diamond", "star",
	"fluid", "fluid-line", "stripe", "stripe-row", "stripe-column",
}

// String returns the shape's name as used by the reference library.
func (d DotType) String() string {
	if int(d) < 0 || int(d) >= len(dotNames) {
		return "square"
	}
	return dotNames[d]
}

// CornerType selects the figure drawn for the three finder patterns.
type CornerType int

// The finder-pattern shapes ported from qrcode-with-logos. CornerSquare is the
// default and the zero value.
const (
	CornerSquare CornerType = iota
	CornerRounded
	CornerCircle
	CornerRoundedCircle
	CornerCircleRounded
	CornerCircleStar
	CornerCircleDiamond
)

var cornerNames = [...]string{
	"square", "rounded", "circle", "rounded-circle", "circle-rounded",
	"circle-star", "circle-diamond",
}

// String returns the shape's name as used by the reference library.
func (c CornerType) String() string {
	if int(c) < 0 || int(c) >= len(cornerNames) {
		return "square"
	}
	return cornerNames[c]
}

// CornerRadius holds the corner radii of a finder pattern in pixels. A zero
// field means "use the default derived from the module size".
type CornerRadius struct {
	Inner float64
	Outer float64
}

// ShapeContext gives a shape function everything it needs to decide its
// geometry: the module grid, the pixel layout, and the consumption state that
// run-merging shapes such as stripe require.
//
// This is deliberately richer than a neighbour mask. The stripe shapes merge
// several modules into one figure and must mark the modules they claim, which a
// read-only snapshot cannot express.
type ShapeContext interface {
	// Dark reports whether (x, y) is a dark module this shape may claim. It is
	// false when the coordinate is out of bounds, light, already consumed, part
	// of a finder pattern, or excluded by the caller.
	Dark(x, y int) bool

	// Consume marks (x, y) as drawn so the main loop skips it.
	Consume(x, y int)

	// Rect returns the pixel rectangle of module (x, y).
	Rect(x, y int) (px, py, size float64)

	// Size returns the module count per side.
	Size() int
}

type shapeContext struct {
	m        *Matrix
	l        layout
	consumed []bool
	excluded func(x, y int) bool
}

// newShapeContext builds the context the dot shapes draw against. excluded may
// be nil; when set it hides further modules, which is how the logo safe zone is
// applied without any shape knowing a logo exists.
func newShapeContext(m *Matrix, l layout, excluded func(x, y int) bool) ShapeContext {
	return &shapeContext{
		m:        m,
		l:        l,
		consumed: make([]bool, m.Size()*m.Size()),
		excluded: excluded,
	}
}

func (c *shapeContext) Size() int { return c.m.Size() }

func (c *shapeContext) Rect(x, y int) (float64, float64, float64) { return c.l.Rect(x, y) }

// Dark folds five separate questions into one, so no shape can ask "is this
// module dark?" without also asking "may I use it?". That is what makes the
// reference library's stripe-runs-through-a-finder-pattern bug impossible here.
func (c *shapeContext) Dark(x, y int) bool {
	n := c.m.Size()
	if x < 0 || y < 0 || x >= n || y >= n {
		return false
	}
	if c.consumed[y*n+x] {
		return false
	}
	if c.m.InFinder(x, y) {
		return false
	}
	if c.excluded != nil && c.excluded(x, y) {
		return false
	}
	return c.m.Dark(x, y)
}

func (c *shapeContext) Consume(x, y int) {
	n := c.m.Size()
	if x >= 0 && y >= 0 && x < n && y < n {
		c.consumed[y*n+x] = true
	}
}

// neighbours4 reports whether the four orthogonal neighbours of (x, y) are
// claimable dark modules. Shapes such as fluid use it to decide which corners
// to round. It is unused until the neighbour-aware shapes land; it lives here so
// the helper sits beside the context it reads.
func neighbours4(c ShapeContext, x, y int) (n, e, s, w bool) {
	return c.Dark(x, y-1), c.Dark(x+1, y), c.Dark(x, y+1), c.Dark(x-1, y)
}
