// Package render turns resolution-independent vector paths into pixels or SVG
// markup. It knows nothing about QR codes, barcodes or any symbology: it deals
// only in paths, colours and images.
package render

import "math"

// Point is a coordinate in the renderer's pixel space.
type Point struct{ X, Y float64 }

// SegKind identifies which of the three segment types a Segment carries.
type SegKind uint8

// The segment types the path model supports. They map one-to-one onto both
// golang.org/x/image/vector and the SVG path commands L, Q and C.
const (
	SegLine SegKind = iota
	SegQuad
	SegCube
)

// Segment is one step of a SubPath. C1 is the control point for SegQuad and the
// first control point for SegCube; C2 is used only by SegCube. To is always the
// end point.
type Segment struct {
	Kind   SegKind
	C1, C2 Point
	To     Point
}

// SubPath is a single contiguous contour. Its winding direction decides whether
// it adds to or subtracts from the filled area: a subpath wound opposite to the
// contour enclosing it punches a hole.
type SubPath struct {
	Start  Point
	Segs   []Segment
	Closed bool
}

// Path is a fill-only shape made of one or more subpaths. There is deliberately
// no stroke concept: rings are expressed as an outer subpath plus an inner
// subpath wound the other way, which the rasteriser and the SVG renderer
// reproduce identically under the non-zero fill rule.
type Path struct {
	SubPaths []SubPath
}

// Builder accumulates subpaths. The zero Builder is ready to use.
type Builder struct {
	path Path
	open bool
}

// MoveTo starts a new subpath at (x, y), ending any subpath in progress.
func (b *Builder) MoveTo(x, y float64) {
	b.path.SubPaths = append(b.path.SubPaths, SubPath{Start: Point{x, y}})
	b.open = true
}

// LineTo adds a straight segment to (x, y).
func (b *Builder) LineTo(x, y float64) {
	b.add(Segment{Kind: SegLine, To: Point{x, y}})
}

// QuadTo adds a quadratic segment to (x, y) with control point (cx, cy).
func (b *Builder) QuadTo(cx, cy, x, y float64) {
	b.add(Segment{Kind: SegQuad, C1: Point{cx, cy}, To: Point{x, y}})
}

// CubeTo adds a cubic segment to (x, y) with control points (c1x, c1y) and
// (c2x, c2y).
func (b *Builder) CubeTo(c1x, c1y, c2x, c2y, x, y float64) {
	b.add(Segment{Kind: SegCube, C1: Point{c1x, c1y}, C2: Point{c2x, c2y}, To: Point{x, y}})
}

// Close marks the current subpath closed.
func (b *Builder) Close() {
	if b.open {
		b.path.SubPaths[len(b.path.SubPaths)-1].Closed = true
		b.open = false
	}
}

// Path returns the accumulated path.
func (b *Builder) Path() Path { return b.path }

func (b *Builder) add(s Segment) {
	if !b.open {
		return
	}
	i := len(b.path.SubPaths) - 1
	b.path.SubPaths[i].Segs = append(b.path.SubPaths[i].Segs, s)
}

// IsEmpty reports whether the path contains no subpaths.
func (p Path) IsEmpty() bool { return len(p.SubPaths) == 0 }

// Bounds returns the axis-aligned bounding box of the path's anchor points.
// Control points are excluded, so for a curve that bulges past its anchors the
// box is a lower bound rather than a tight fit.
func (p Path) Bounds() (minX, minY, maxX, maxY float64) {
	if p.IsEmpty() {
		return 0, 0, 0, 0
	}
	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)
	grow := func(pt Point) {
		minX, minY = math.Min(minX, pt.X), math.Min(minY, pt.Y)
		maxX, maxY = math.Max(maxX, pt.X), math.Max(maxY, pt.Y)
	}
	for _, sp := range p.SubPaths {
		grow(sp.Start)
		for _, s := range sp.Segs {
			grow(s.To)
		}
	}
	return minX, minY, maxX, maxY
}

// Append returns a path holding this path's subpaths followed by those of each
// argument. It is how a ring is assembled: outer.Append(inner.Reverse()).
func (p Path) Append(qs ...Path) Path {
	out := Path{SubPaths: append([]SubPath(nil), p.SubPaths...)}
	for _, q := range qs {
		out.SubPaths = append(out.SubPaths, q.SubPaths...)
	}
	return out
}

// Reverse returns the path with every subpath wound the other way, turning a
// solid contour into a hole when appended to an enclosing contour.
func (p Path) Reverse() Path {
	out := Path{SubPaths: make([]SubPath, 0, len(p.SubPaths))}
	for _, sp := range p.SubPaths {
		pts := make([]Point, 0, len(sp.Segs)+1)
		pts = append(pts, sp.Start)
		for _, s := range sp.Segs {
			pts = append(pts, s.To)
		}
		rev := SubPath{Start: pts[len(pts)-1], Closed: sp.Closed}
		for i := len(sp.Segs) - 1; i >= 0; i-- {
			s := sp.Segs[i]
			rs := Segment{Kind: s.Kind, To: pts[i]}
			switch s.Kind {
			case SegQuad:
				rs.C1 = s.C1
			case SegCube:
				rs.C1, rs.C2 = s.C2, s.C1
			}
			rev.Segs = append(rev.Segs, rs)
		}
		out.SubPaths = append(out.SubPaths, rev)
	}
	return out
}

// Rotate returns the path rotated by rad radians about (cx, cy).
func (p Path) Rotate(rad, cx, cy float64) Path {
	sin, cos := math.Sin(rad), math.Cos(rad)
	rot := func(pt Point) Point {
		dx, dy := pt.X-cx, pt.Y-cy
		return Point{cx + dx*cos - dy*sin, cy + dx*sin + dy*cos}
	}
	out := Path{SubPaths: make([]SubPath, 0, len(p.SubPaths))}
	for _, sp := range p.SubPaths {
		n := SubPath{Start: rot(sp.Start), Closed: sp.Closed}
		for _, s := range sp.Segs {
			n.Segs = append(n.Segs, Segment{
				Kind: s.Kind, C1: rot(s.C1), C2: rot(s.C2), To: rot(s.To),
			})
		}
		out.SubPaths = append(out.SubPaths, n)
	}
	return out
}

// kappa is the cubic Bezier constant that approximates a quarter circle to
// within about 0.02% of the true arc.
const kappa = 0.5522847498307936

// RoundRect returns a closed rectangle with independent corner radii, wound
// clockwise in a y-down coordinate system. Radii are clamped to half the
// shorter side and negative radii are treated as zero.
func RoundRect(x, y, w, h, rTL, rTR, rBR, rBL float64) Path {
	lim := math.Min(w, h) / 2
	clamp := func(r float64) float64 {
		if r < 0 {
			return 0
		}
		return math.Min(r, lim)
	}
	tl, tr, br, bl := clamp(rTL), clamp(rTR), clamp(rBR), clamp(rBL)

	var b Builder
	b.MoveTo(x+tl, y)
	b.LineTo(x+w-tr, y)
	if tr > 0 {
		b.CubeTo(x+w-tr+tr*kappa, y, x+w, y+tr-tr*kappa, x+w, y+tr)
	}
	b.LineTo(x+w, y+h-br)
	if br > 0 {
		b.CubeTo(x+w, y+h-br+br*kappa, x+w-br+br*kappa, y+h, x+w-br, y+h)
	}
	b.LineTo(x+bl, y+h)
	if bl > 0 {
		b.CubeTo(x+bl-bl*kappa, y+h, x, y+h-bl+bl*kappa, x, y+h-bl)
	}
	b.LineTo(x, y+tl)
	if tl > 0 {
		b.CubeTo(x, y+tl-tl*kappa, x+tl-tl*kappa, y, x+tl, y)
	}
	b.Close()
	return b.Path()
}

// Circle returns a closed circle of radius r centred at (cx, cy), wound
// clockwise, built from four cubic arcs.
func Circle(cx, cy, r float64) Path {
	return RoundRect(cx-r, cy-r, 2*r, 2*r, r, r, r, r)
}
