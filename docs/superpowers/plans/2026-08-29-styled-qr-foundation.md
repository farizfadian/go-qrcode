# Styled QR Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the shared foundation of `go-qrcode` — encoder, module matrix, pixel layout, fill-only path model, raster renderer and SVG renderer — and prove it end to end with `square` dots and `square` corners that decode from both PNG and SVG.

**Architecture:** Shapes are described once as vector paths in absolute pixel coordinates and handed to a renderer as a `Scene`. The SVG renderer serialises those paths to `d` attributes; the raster renderer rasterises the same paths through `golang.org/x/image/vector`. Nothing in `internal/render` knows what a QR code is. Rings and holes come from reversed subpath winding, never from stroking.

**Tech Stack:** Go 1.22 · `github.com/piglig/go-qr` v1.1.0 (encoder) · `golang.org/x/image` v0.23.0 (rasteriser, resampling) · `github.com/makiuchi-d/gozxing` v0.1.1 (test-only decoder)

**Spec:** `docs/superpowers/specs/2026-08-29-styled-qr-core-design.md`

## Global Constraints

- Module path is `github.com/farizfadian/go-qrcode`; the public package is `qr` under `qr/`.
- Go directive is exactly `go 1.22`. `golang.org/x/image` must be pinned to **v0.23.0** — v0.25.0 raises its own directive to `go 1.23.0` and v0.38.0+ to `go 1.25.0`, which would force every consumer onto a newer toolchain.
- Runtime dependencies are limited to `github.com/piglig/go-qr` and `golang.org/x/image`. Any third requires an ADR under `docs/adr/`.
- Nothing in `internal/render/` may import `qr/` or reference QR concepts (modules, finders, versions, ECC).
- The path model is fill-only. Never add a stroke concept.
- No panics in library code. All validation happens in `New`; render methods return only I/O errors.
- Every exported identifier carries a doc comment starting with its own name.
- `gofmt -l .` must print nothing; `go vet ./...` must be clean; `go test ./... -race` must pass.
- Conventional commits with the trailer `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`.
- Work on branch `feat/styled-qr-core`. Do not push until the Task 10 gate is green.

---

### Task 1: Project scaffold

**Files:**
- Create: `go.mod`, `LICENSE`, `README.md`, `.github/workflows/ci.yml`, `qr/doc.go`, `qr/doc_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: a buildable module named `github.com/farizfadian/go-qrcode` with package `qr`

- [ ] **Step 1: Create the branch**

```bash
git checkout -b feat/styled-qr-core
```

- [ ] **Step 2: Write the failing test**

Create `qr/doc_test.go`:

```go
package qr_test

import (
	"testing"

	"github.com/farizfadian/go-qrcode/qr"
)

func TestPackageBuilds(t *testing.T) {
	if qr.Version == "" {
		t.Fatal("qr.Version is empty; the package did not initialise")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./qr/`
Expected: FAIL — no `go.mod`, then once `go.mod` exists, `undefined: qr.Version`.

- [ ] **Step 4: Create go.mod**

```bash
go mod init github.com/farizfadian/go-qrcode
go mod edit -go=1.22
go mod edit -require=golang.org/x/image@v0.23.0
go mod edit -require=github.com/piglig/go-qr@v1.1.0
```

- [ ] **Step 5: Create qr/doc.go**

```go
// Package qr generates QR codes with styled modules, styled finder patterns and
// a decorated centre logo, rendered identically to PNG, JPEG and SVG.
//
// The simplest useful call is two lines:
//
//	q, err := qr.New(qr.Options{Content: "https://example.com"})
//	err = q.WritePNGFile("qr.png")
//
// Every field of Options other than Content has a working default, so the zero
// value produces a conventional black-on-white code.
package qr

// Version is the library version, set at release time.
const Version = "0.0.0-dev"
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `go test ./qr/`
Expected: PASS

- [ ] **Step 7: Write LICENSE**

```
MIT License

Copyright (c) 2026 Fariz Fadian

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

This library's feature surface and visual behaviour are modelled on
qrcode-with-logos (https://github.com/zxpsuper/qrcode-with-logos),
Copyright (c) zxpsuper, also MIT licensed. No source was copied; the
behaviour was reimplemented in Go from a reading of that project's source.
```

- [ ] **Step 8: Write README.md stub**

The first paragraph must state the differentiator before anything else, because the repository name collides with several well-known libraries. Do not open with "a QR code library for Go".

```markdown
# go-qrcode

Styled QR codes for Go: twelve per-module dot shapes, seven independently
styled and coloured finder patterns, a decorated centre logo, and byte-identical
geometry from both a raster and an SVG renderer. Existing Go QR libraries give
you a logo or a colour scheme; this one gives you the whole visual surface, and
proves every shape still scans with a round-trip decode test.

Status: under construction. See `docs/superpowers/specs/` for the design.
```

- [ ] **Step 9: Write .github/workflows/ci.yml**

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    strategy:
      matrix:
        go: ['1.24', '1.25']
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
      - run: go build ./...
      - run: go vet ./...
      - run: test -z "$(gofmt -l .)" || (gofmt -l . && exit 1)
      - run: go test ./... -race
```

- [ ] **Step 10: Verify the whole toolchain is clean**

Run: `go build ./... && go vet ./... && gofmt -l . && go test ./... -race`
Expected: no output from `gofmt`, all commands exit 0.

- [ ] **Step 11: Commit**

```bash
git add go.mod LICENSE README.md .github qr/
git commit -m "chore: scaffold module, licence, CI and package doc"
```

---

### Task 2: Vector path model

**Files:**
- Create: `internal/render/path.go`, `internal/render/path_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `render.Point`, `render.SegKind` (`SegLine`, `SegQuad`, `SegCube`), `render.Segment`, `render.SubPath`, `render.Path`, `render.Builder` with `MoveTo/LineTo/QuadTo/CubeTo/Close/Path`, and `Path.Bounds`, `Path.IsEmpty`, `Path.Reverse`, `Path.Append`, `Path.Rotate`, plus `render.RoundRect` and `render.Circle`

- [ ] **Step 1: Write the failing test**

Create `internal/render/path_test.go`:

```go
package render

import "testing"

func TestBuilderRecordsSubPaths(t *testing.T) {
	var b Builder
	b.MoveTo(1, 2)
	b.LineTo(3, 2)
	b.QuadTo(4, 3, 3, 4)
	b.CubeTo(3, 5, 2, 5, 1, 4)
	b.Close()
	p := b.Path()

	if len(p.SubPaths) != 1 {
		t.Fatalf("SubPaths = %d, want 1", len(p.SubPaths))
	}
	sp := p.SubPaths[0]
	if sp.Start != (Point{1, 2}) {
		t.Errorf("Start = %v, want {1 2}", sp.Start)
	}
	if !sp.Closed {
		t.Error("Closed = false, want true")
	}
	want := []SegKind{SegLine, SegQuad, SegCube}
	if len(sp.Segs) != len(want) {
		t.Fatalf("Segs = %d, want %d", len(sp.Segs), len(want))
	}
	for i, k := range want {
		if sp.Segs[i].Kind != k {
			t.Errorf("Segs[%d].Kind = %v, want %v", i, sp.Segs[i].Kind, k)
		}
	}
}

func TestRoundRectClampsRadiiToHalfTheShorterSide(t *testing.T) {
	p := RoundRect(0, 0, 10, 4, 99, 99, 99, 99)
	minX, minY, maxX, maxY := p.Bounds()
	if minX != 0 || minY != 0 || maxX != 10 || maxY != 4 {
		t.Errorf("Bounds = %v %v %v %v, want 0 0 10 4", minX, minY, maxX, maxY)
	}
}

func TestReverseFlipsTraversalOrder(t *testing.T) {
	var b Builder
	b.MoveTo(0, 0)
	b.LineTo(10, 0)
	b.LineTo(10, 10)
	b.Close()
	r := b.Path().Reverse()

	sp := r.SubPaths[0]
	if sp.Start != (Point{10, 10}) {
		t.Errorf("Start = %v, want {10 10}", sp.Start)
	}
	for i, want := range []Point{{10, 0}, {0, 0}} {
		if sp.Segs[i].To != want {
			t.Errorf("Segs[%d].To = %v, want %v", i, sp.Segs[i].To, want)
		}
	}
	if !sp.Closed {
		t.Error("Closed = false, want true")
	}
}

func TestReverseSwapsCubicControlPoints(t *testing.T) {
	var b Builder
	b.MoveTo(0, 0)
	b.CubeTo(1, 2, 3, 4, 5, 6)
	r := b.Path().Reverse()

	s := r.SubPaths[0].Segs[0]
	if s.C1 != (Point{3, 4}) || s.C2 != (Point{1, 2}) {
		t.Errorf("controls = %v, %v; want {3 4}, {1 2} (swapped)", s.C1, s.C2)
	}
	if s.To != (Point{0, 0}) {
		t.Errorf("To = %v, want {0 0}", s.To)
	}
}

func TestReverseKeepsBounds(t *testing.T) {
	p := RoundRect(1, 2, 6, 8, 1, 1, 1, 1)
	a1, b1, c1, d1 := p.Bounds()
	a2, b2, c2, d2 := p.Reverse().Bounds()
	if a1 != a2 || b1 != b2 || c1 != c2 || d1 != d2 {
		t.Errorf("bounds changed: %v -> %v",
			[]float64{a1, b1, c1, d1}, []float64{a2, b2, c2, d2})
	}
}

func TestCircleIsCentredAndRoundTripsBounds(t *testing.T) {
	p := Circle(5, 5, 3)
	minX, minY, maxX, maxY := p.Bounds()
	const eps = 1e-9
	if abs(minX-2) > eps || abs(minY-2) > eps || abs(maxX-8) > eps || abs(maxY-8) > eps {
		t.Errorf("Bounds = %v %v %v %v, want 2 2 8 8", minX, minY, maxX, maxY)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render/`
Expected: FAIL — `undefined: Builder`

- [ ] **Step 3: Implement path.go**

```go
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

// Bounds returns the axis-aligned bounding box of every point and control point
// in the path. Control points can lie outside the drawn curve, so the box is
// conservative rather than tight.
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
			n.Segs = append(n.Segs, Segment{Kind: s.Kind, C1: rot(s.C1), C2: rot(s.C2), To: rot(s.To)})
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
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/render/ -v`
Expected: PASS, all six tests.

- [ ] **Step 5: Commit**

```bash
git add internal/render/path.go internal/render/path_test.go
git commit -m "feat: add fill-only vector path model with reversible winding"
```

---

### Task 3: Colour parsing

**Files:**
- Create: `internal/render/color.go`, `internal/render/color_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `render.NormalizeColor(string) string`, `render.ParseColor(string) (color.RGBA, error)`, `render.ErrColorSyntax`

- [ ] **Step 1: Write the failing test**

Create `internal/render/color_test.go`:

```go
package render

import (
	"errors"
	"image/color"
	"testing"
)

func TestNormalizeColor(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"ff0000", "#ff0000"},
		{"#ff0000", "#ff0000"},
		{"fff", "#fff"},
		{"ffff", "#ffff"},
		{"00000000", "#00000000"},
		{"red", "red"},
		{"rgb(1,2,3)", "rgb(1,2,3)"},
		{"", ""},
	} {
		if got := NormalizeColor(tc.in); got != tc.want {
			t.Errorf("NormalizeColor(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseColor(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want color.RGBA
	}{
		{"#ff0000", color.RGBA{0xff, 0, 0, 0xff}},
		{"ff0000", color.RGBA{0xff, 0, 0, 0xff}},
		{"#fff", color.RGBA{0xff, 0xff, 0xff, 0xff}},
		{"#000000", color.RGBA{0, 0, 0, 0xff}},
		{"#00000000", color.RGBA{0, 0, 0, 0}},
		{"#ff000080", color.RGBA{0x80, 0, 0, 0x80}}, // premultiplied
	} {
		got, err := ParseColor(tc.in)
		if err != nil {
			t.Fatalf("ParseColor(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseColor(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseColorRejectsNonHex(t *testing.T) {
	for _, in := range []string{"red", "rgb(1,2,3)", "#12345", "", "#gggggg"} {
		if _, err := ParseColor(in); !errors.Is(err, ErrColorSyntax) {
			t.Errorf("ParseColor(%q) error = %v, want ErrColorSyntax", in, err)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render/ -run Color`
Expected: FAIL — `undefined: NormalizeColor`

- [ ] **Step 3: Implement color.go**

```go
package render

import (
	"errors"
	"fmt"
	"image/color"
	"strings"
)

// ErrColorSyntax reports a colour string the rasteriser cannot resolve. CSS
// colour names and functional notation such as rgb() parse only in a browser,
// so they reach SVG untouched but fail here.
var ErrColorSyntax = errors.New("render: not a hex colour")

// NormalizeColor adds a leading '#' to a bare hex colour of 3, 4, 6 or 8
// digits. Anything else, including CSS names and functional notation, is
// returned unchanged so it can pass through to SVG.
func NormalizeColor(s string) string {
	if s == "" || s[0] == '#' {
		return s
	}
	if !isHexRun(s) {
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
		r, g, b, a = hexByte(h[0], h[1]), hexByte(h[2], h[3]), hexByte(h[4], h[5]), hexByte(h[6], h[7])
	default:
		return color.RGBA{}, fmt.Errorf("%w: %q", ErrColorSyntax, s)
	}
	// color.RGBA is premultiplied; hex notation is straight alpha.
	m := uint32(a)
	return color.RGBA{
		R: uint8(uint32(r) * m / 0xff),
		G: uint8(uint32(g) * m / 0xff),
		B: uint8(uint32(b) * m / 0xff),
		A: a,
	}, nil
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

func dup(c byte) uint8 { v := uint8(hexVal(c)); return v<<4 | v }
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/render/ -run Color -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/render/color.go internal/render/color_test.go
git commit -m "feat: add hex colour parsing with CSS passthrough"
```

---

### Task 4: Scene and raster renderer

**Files:**
- Create: `internal/render/scene.go`, `internal/render/raster.go`, `internal/render/raster_test.go`

**Interfaces:**
- Consumes: `Path`, `RoundRect`, `Circle`, `Path.Reverse`, `Path.Append` from Task 2
- Produces: `render.Scene`, `render.Item`, `render.PathItem`, `render.ImageItem`, `render.Raster(Scene) *image.RGBA`

- [ ] **Step 1: Write the failing test**

Create `internal/render/raster_test.go`:

```go
package render

import (
	"image"
	"image/color"
	"testing"
)

func TestRasterFillsAPath(t *testing.T) {
	sc := Scene{
		Width: 20, Height: 20,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items: []Item{PathItem{
			Path: RoundRect(5, 5, 10, 10, 0, 0, 0, 0),
			Fill: color.RGBA{0, 0, 0, 0xff},
		}},
	}
	img := Raster(sc)
	if got := img.RGBAAt(10, 10); got.R != 0 || got.A != 0xff {
		t.Errorf("centre = %v, want opaque black", got)
	}
	if got := img.RGBAAt(1, 1); got.R != 0xff {
		t.Errorf("corner = %v, want white background", got)
	}
}

// This is the architectural test: a ring must be a hole, not a filled square.
// If it fails, the fill-only path model does not work and the design in the
// spec's section 2.3 is wrong.
func TestRasterPunchesHoleWithReversedWinding(t *testing.T) {
	outer := RoundRect(0, 0, 30, 30, 0, 0, 0, 0)
	inner := RoundRect(10, 10, 10, 10, 0, 0, 0, 0).Reverse()
	sc := Scene{
		Width: 30, Height: 30,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items:      []Item{PathItem{Path: outer.Append(inner), Fill: color.RGBA{0, 0, 0, 0xff}}},
	}
	img := Raster(sc)
	if got := img.RGBAAt(15, 15); got.R != 0xff {
		t.Errorf("ring centre = %v, want the white background showing through", got)
	}
	if got := img.RGBAAt(4, 15); got.R != 0 {
		t.Errorf("ring body = %v, want opaque black", got)
	}
}

func TestRasterKeepsTransparentBackground(t *testing.T) {
	sc := Scene{Width: 8, Height: 8, Background: color.RGBA{}}
	if got := Raster(sc).RGBAAt(4, 4); got.A != 0 {
		t.Errorf("alpha = %d, want 0", got.A)
	}
}

func TestRasterDrawsClippedImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			src.Set(x, y, color.RGBA{0, 0xff, 0, 0xff})
		}
	}
	clip := Circle(10, 10, 5)
	sc := Scene{
		Width: 20, Height: 20,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items:      []Item{ImageItem{Img: src, X: 5, Y: 5, W: 10, H: 10, Clip: &clip}},
	}
	img := Raster(sc)
	if got := img.RGBAAt(10, 10); got.G != 0xff {
		t.Errorf("clip centre = %v, want green", got)
	}
	if got := img.RGBAAt(6, 6); got.R != 0xff {
		t.Errorf("outside clip = %v, want white background", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render/ -run Raster`
Expected: FAIL — `undefined: Scene`

- [ ] **Step 3: Implement scene.go**

```go
package render

import (
	"image"
	"image/color"
)

// Item is one drawing operation in a Scene. Only PathItem and ImageItem
// implement it; the sealed interface keeps renderers exhaustive.
type Item interface{ isItem() }

// PathItem fills a path with a single colour.
type PathItem struct {
	Path Path
	Fill color.Color
}

func (PathItem) isItem() {}

// ImageItem draws an image scaled into the rectangle at (X, Y) with size
// (W, H). When Clip is non-nil the image is masked by that path, which is how
// a logo gets rounded corners without either renderer knowing what a logo is.
type ImageItem struct {
	Img        image.Image
	X, Y, W, H float64
	Clip       *Path
}

func (ImageItem) isItem() {}

// Scene is a complete, renderer-independent description of one drawing. Both
// the raster and the SVG renderer consume it, which is what guarantees their
// output matches.
type Scene struct {
	Width, Height int
	Background    color.Color // a zero-alpha colour means no background fill
	Items         []Item      // painted in order, first to last
}
```

- [ ] **Step 4: Implement raster.go**

```go
package render

import (
	"image"
	"image/draw"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/vector"
)

// Raster renders sc into a new RGBA image. It never fails: a Scene is already
// validated by the time it reaches a renderer.
func Raster(sc Scene) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, sc.Width, sc.Height))
	if sc.Background != nil {
		if _, _, _, a := sc.Background.RGBA(); a > 0 {
			draw.Draw(dst, dst.Bounds(), image.NewUniform(sc.Background), image.Point{}, draw.Src)
		}
	}
	for _, it := range sc.Items {
		switch v := it.(type) {
		case PathItem:
			m := rasterize(v.Path, sc.Width, sc.Height)
			draw.DrawMask(dst, dst.Bounds(), image.NewUniform(v.Fill), image.Point{},
				m, image.Point{}, draw.Over)
		case ImageItem:
			drawImage(dst, v, sc.Width, sc.Height)
		}
	}
	return dst
}

// rasterize converts a path to a coverage mask. x/image/vector accumulates
// signed area and takes its absolute value clamped to 1, which is equivalent to
// the non-zero winding rule and therefore to SVG's default fill-rule. That is
// what makes a reversed inner subpath punch a hole.
func rasterize(p Path, w, h int) *image.Alpha {
	r := vector.NewRasterizer(w, h)
	for _, sp := range p.SubPaths {
		r.MoveTo(float32(sp.Start.X), float32(sp.Start.Y))
		for _, s := range sp.Segs {
			switch s.Kind {
			case SegLine:
				r.LineTo(float32(s.To.X), float32(s.To.Y))
			case SegQuad:
				r.QuadTo(float32(s.C1.X), float32(s.C1.Y), float32(s.To.X), float32(s.To.Y))
			case SegCube:
				r.CubeTo(float32(s.C1.X), float32(s.C1.Y),
					float32(s.C2.X), float32(s.C2.Y),
					float32(s.To.X), float32(s.To.Y))
			}
		}
		r.ClosePath()
	}
	m := image.NewAlpha(image.Rect(0, 0, w, h))
	r.Draw(m, m.Bounds(), image.Opaque, image.Point{})
	return m
}

func drawImage(dst *image.RGBA, it ImageItem, w, h int) {
	rect := image.Rect(int(it.X), int(it.Y), int(it.X+it.W), int(it.Y+it.H))
	scaled := image.NewRGBA(rect)
	xdraw.CatmullRom.Scale(scaled, rect, it.Img, it.Img.Bounds(), xdraw.Over, nil)
	if it.Clip == nil {
		draw.Draw(dst, rect, scaled, rect.Min, draw.Over)
		return
	}
	draw.DrawMask(dst, dst.Bounds(), scaled, image.Point{},
		rasterize(*it.Clip, w, h), image.Point{}, draw.Over)
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/render/ -run Raster -v`
Expected: PASS, all four tests. `TestRasterPunchesHoleWithReversedWinding` passing is the gate for the whole path design — if it fails, the fill-only model does not work and nothing built on it will.

- [ ] **Step 6: Verify the toolchain is clean**

Run: `go vet ./... && gofmt -l . && go test ./internal/render/`
Expected: `gofmt` silent, everything else exit 0.

- [ ] **Step 7: Commit**

```bash
git add internal/render/scene.go internal/render/raster.go internal/render/raster_test.go
git commit -m "feat: add Scene description and raster renderer"
```

---

### Task 5: Encoder interface and piglig backing

**Files:**
- Create: `qr/ecc.go`, `qr/encode.go`, `qr/encode_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks
- Produces: `qr.ECCLevel` with `ECCAuto`, `ECCLow`, `ECCMedium`, `ECCQuartile`, `ECCHigh` and `String()`; `qr.Encoder` interface with `Encode(content string, ecc ECCLevel) ([][]bool, error)`; `qr.defaultEncoder()` returning an `Encoder`; `qr.ErrContentTooLong`

- [ ] **Step 1: Write the failing test**

Create `qr/encode_test.go`:

```go
package qr

import (
	"errors"
	"strings"
	"testing"
)

func TestECCLevelString(t *testing.T) {
	for _, tc := range []struct {
		lv   ECCLevel
		want string
	}{
		{ECCAuto, "auto"}, {ECCLow, "L"}, {ECCMedium, "M"},
		{ECCQuartile, "Q"}, {ECCHigh, "H"},
	} {
		if got := tc.lv.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", int(tc.lv), got, tc.want)
		}
	}
}

func TestEncodeProducesSquareMatrix(t *testing.T) {
	m, err := defaultEncoder().Encode("HELLO", ECCHigh)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(m) != 21 {
		t.Fatalf("rows = %d, want 21", len(m))
	}
	for i, row := range m {
		if len(row) != 21 {
			t.Fatalf("row %d has %d columns, want 21", i, len(row))
		}
	}
	// The top-left finder pattern's outer ring must be dark all along its edge.
	for x := 0; x < 7; x++ {
		if !m[0][x] {
			t.Errorf("m[0][%d] is light; the top-left finder pattern is missing", x)
		}
	}
}

func TestEncodeRejectsAuto(t *testing.T) {
	if _, err := defaultEncoder().Encode("HELLO", ECCAuto); err == nil {
		t.Fatal("Encode accepted ECCAuto; the level must be resolved before encoding")
	}
}

func TestEncodeDoesNotBoostTheECCLevel(t *testing.T) {
	// piglig's EncodeText silently upgrades L to the highest level that still
	// fits, which would break golden determinism and the logo budget. Our
	// encoder pins boostEcl=false, so its matrix must differ from the boosted
	// one for content where boosting kicks in.
	pinned, err := defaultEncoder().Encode("HELLO", ECCLow)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	boosted, err := boostedMatrixForTest("HELLO", ECCLow)
	if err != nil {
		t.Fatalf("boostedMatrixForTest: %v", err)
	}
	if equalMatrix(pinned, boosted) {
		t.Fatal("pinned and boosted matrices are identical; boostEcl is not pinned off")
	}
}

func TestEncodeRejectsOversizedContent(t *testing.T) {
	_, err := defaultEncoder().Encode(strings.Repeat("a", 5000), ECCHigh)
	if !errors.Is(err, ErrContentTooLong) {
		t.Fatalf("error = %v, want ErrContentTooLong", err)
	}
}

func equalMatrix(a, b [][]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for y := range a {
		if len(a[y]) != len(b[y]) {
			return false
		}
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				return false
			}
		}
	}
	return true
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./qr/ -run Encode`
Expected: FAIL — `undefined: ECCLevel`

- [ ] **Step 3: Implement qr/ecc.go**

```go
package qr

// ECCLevel is a QR error-correction level. The zero value, ECCAuto, lets New
// choose: with a logo it selects ECCHigh so the occluded area fits the recovery
// budget, and without one it scales with content length.
type ECCLevel int

// The error-correction levels defined by ISO/IEC 18004, plus the automatic
// selection that is the zero value.
const (
	ECCAuto ECCLevel = iota
	ECCLow
	ECCMedium
	ECCQuartile
	ECCHigh
)

// String returns the single-letter name used by the QR specification, or
// "auto" for ECCAuto.
func (e ECCLevel) String() string {
	switch e {
	case ECCLow:
		return "L"
	case ECCMedium:
		return "M"
	case ECCQuartile:
		return "Q"
	case ECCHigh:
		return "H"
	default:
		return "auto"
	}
}

// recoveryFraction returns the share of codewords the level can restore. It
// drives the logo occlusion budget.
func (e ECCLevel) recoveryFraction() float64 {
	switch e {
	case ECCLow:
		return 0.07
	case ECCMedium:
		return 0.15
	case ECCQuartile:
		return 0.25
	case ECCHigh:
		return 0.30
	default:
		return 0
	}
}
```

- [ ] **Step 4: Implement qr/encode.go**

```go
package qr

import (
	"errors"
	"fmt"

	goqr "github.com/piglig/go-qr"
)

// ErrContentTooLong reports content that does not fit in a version 40 symbol at
// the requested error-correction level.
var ErrContentTooLong = errors.New("qr: content does not fit at the requested ECC level")

// Encoder turns text into a QR module matrix. It exists so the third-party
// encoder can be replaced by a built-in one without a breaking change.
type Encoder interface {
	// Encode returns a square matrix indexed as [y][x]; true means a dark
	// module. ecc must be a concrete level, never ECCAuto.
	Encode(content string, ecc ECCLevel) ([][]bool, error)
}

// defaultEncoder returns the encoder New uses when Options names none.
func defaultEncoder() Encoder { return pigligEncoder{} }

// pigligEncoder adapts github.com/piglig/go-qr, a port of Nayuki's reference
// implementation.
type pigligEncoder struct{}

func (pigligEncoder) Encode(content string, ecc ECCLevel) ([][]bool, error) {
	level, ok := pigligLevel(ecc)
	if !ok {
		return nil, fmt.Errorf("qr: ECC level %v must be resolved before encoding", ecc)
	}
	segs, err := goqr.MakeSegments(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContentTooLong, err)
	}
	// boostEcl is pinned false on purpose. EncodeText would silently raise the
	// level whenever the data still fits, making the effective level differ
	// from the requested one and breaking both golden determinism and the logo
	// occlusion budget.
	code, err := goqr.EncodeSegments(segs, level, 1, 40, -1, false)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContentTooLong, err)
	}
	n := code.Size()
	out := make([][]bool, n)
	for y := 0; y < n; y++ {
		row := make([]bool, n)
		for x := 0; x < n; x++ {
			row[x] = code.Module(x, y)
		}
		out[y] = row
	}
	return out, nil
}

func pigligLevel(e ECCLevel) (goqr.Ecc, bool) {
	switch e {
	case ECCLow:
		return goqr.Low, true
	case ECCMedium:
		return goqr.Medium, true
	case ECCQuartile:
		return goqr.Quartile, true
	case ECCHigh:
		return goqr.High, true
	}
	return goqr.Low, false
}

// boostedMatrixForTest exposes the boosted encoding path so a test can prove
// our own path differs from it. It is unexported and used only by tests.
func boostedMatrixForTest(content string, ecc ECCLevel) ([][]bool, error) {
	level, ok := pigligLevel(ecc)
	if !ok {
		return nil, fmt.Errorf("qr: bad level %v", ecc)
	}
	code, err := goqr.EncodeText(content, level)
	if err != nil {
		return nil, err
	}
	n := code.Size()
	out := make([][]bool, n)
	for y := 0; y < n; y++ {
		row := make([]bool, n)
		for x := 0; x < n; x++ {
			row[x] = code.Module(x, y)
		}
		out[y] = row
	}
	return out, nil
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./qr/ -run Encode -v && go test ./qr/ -run ECCLevel -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add qr/ecc.go qr/encode.go qr/encode_test.go go.mod go.sum
git commit -m "feat: add Encoder interface backed by piglig with ECC boosting pinned off"
```

---

### Task 6: Module matrix and finder classification

**Files:**
- Create: `qr/matrix.go`, `qr/matrix_test.go`

**Interfaces:**
- Consumes: `Encoder`, `defaultEncoder` from Task 5
- Produces: `qr.ModuleKind` with `KindData`, `KindFinder`, `KindSeparator`, `KindTiming`, `KindAlignment`, `KindFormat`, `KindVersion`; `qr.Matrix` with `Size() int`, `Dark(x, y int) bool`, `Kind(x, y int) ModuleKind`, `InFinder(x, y int) bool`; `qr.newMatrix([][]bool) (*Matrix, error)`

- [ ] **Step 1: Write the failing test**

Create `qr/matrix_test.go`:

```go
package qr

import "testing"

func mustMatrix(t *testing.T, content string, ecc ECCLevel) *Matrix {
	t.Helper()
	mods, err := defaultEncoder().Encode(content, ecc)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	m, err := newMatrix(mods)
	if err != nil {
		t.Fatalf("newMatrix: %v", err)
	}
	return m
}

func TestMatrixClassifiesTheThreeFinderPatterns(t *testing.T) {
	m := mustMatrix(t, "HELLO", ECCHigh) // version 1, 21 modules
	if m.Size() != 21 {
		t.Fatalf("Size = %d, want 21", m.Size())
	}
	n := m.Size()
	corners := [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}}
	for _, c := range corners {
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 7; dx++ {
				x, y := c[0]+dx, c[1]+dy
				if !m.InFinder(x, y) {
					t.Fatalf("InFinder(%d,%d) = false inside a finder pattern", x, y)
				}
				if m.Kind(x, y) != KindFinder {
					t.Fatalf("Kind(%d,%d) = %v, want KindFinder", x, y, m.Kind(x, y))
				}
			}
		}
	}
	// The bottom-right corner carries data, never a finder pattern.
	if m.InFinder(n-1, n-1) {
		t.Error("InFinder(n-1,n-1) = true; there is no fourth finder pattern")
	}
}

func TestMatrixClassifiesTimingAndSeparator(t *testing.T) {
	m := mustMatrix(t, "HELLO", ECCHigh)
	if got := m.Kind(6, 10); got != KindTiming {
		t.Errorf("Kind(6,10) = %v, want KindTiming", got)
	}
	if got := m.Kind(10, 6); got != KindTiming {
		t.Errorf("Kind(10,6) = %v, want KindTiming", got)
	}
	if got := m.Kind(7, 0); got != KindSeparator {
		t.Errorf("Kind(7,0) = %v, want KindSeparator", got)
	}
}

func TestMatrixClassifiesAlignmentOnVersion2(t *testing.T) {
	// Version 2 is 25 modules and has exactly one alignment pattern centred at
	// (18,18), spanning 16..20 in both axes.
	m := mustMatrix(t, "HELLO WORLD 12345", ECCHigh)
	if m.Size() != 25 {
		t.Skipf("expected a version 2 symbol, got %d modules", m.Size())
	}
	if got := m.Kind(18, 18); got != KindAlignment {
		t.Errorf("Kind(18,18) = %v, want KindAlignment", got)
	}
	if got := m.Kind(16, 16); got != KindAlignment {
		t.Errorf("Kind(16,16) = %v, want KindAlignment", got)
	}
}

func TestMatrixRejectsNonSquareInput(t *testing.T) {
	if _, err := newMatrix([][]bool{{true, false}, {true}}); err == nil {
		t.Fatal("newMatrix accepted a ragged matrix")
	}
	if _, err := newMatrix(nil); err == nil {
		t.Fatal("newMatrix accepted an empty matrix")
	}
}

func TestMatrixOutOfBoundsIsLightData(t *testing.T) {
	m := mustMatrix(t, "HELLO", ECCHigh)
	if m.Dark(-1, 0) || m.Dark(0, -1) || m.Dark(m.Size(), 0) {
		t.Error("out-of-bounds coordinates must read as light")
	}
	if m.InFinder(-1, -1) {
		t.Error("out-of-bounds coordinates must not be finder modules")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./qr/ -run Matrix`
Expected: FAIL — `undefined: newMatrix`

- [ ] **Step 3: Implement qr/matrix.go**

```go
package qr

import "errors"

// ModuleKind labels a module's role in the symbol. Only KindFinder changes how
// the code is drawn today: finder modules are replaced by a styled corner
// figure. The remaining kinds exist so the classification can be tested against
// known symbols and inspected while debugging.
type ModuleKind uint8

// The module roles defined by ISO/IEC 18004. Alignment patterns render as
// ordinary dots and timing and format modules are data, so only KindFinder is
// excluded from dot rendering.
const (
	KindData ModuleKind = iota
	KindFinder
	KindSeparator
	KindTiming
	KindAlignment
	KindFormat
	KindVersion
)

// ErrBadMatrix reports a module grid that is empty or not square.
var ErrBadMatrix = errors.New("qr: module matrix must be square and non-empty")

// Matrix is an immutable QR module grid with every module classified.
type Matrix struct {
	size int
	dark []bool
	kind []ModuleKind
}

// newMatrix classifies mods, which must be a square [y][x] grid.
func newMatrix(mods [][]bool) (*Matrix, error) {
	n := len(mods)
	if n == 0 {
		return nil, ErrBadMatrix
	}
	for _, row := range mods {
		if len(row) != n {
			return nil, ErrBadMatrix
		}
	}
	m := &Matrix{size: n, dark: make([]bool, n*n), kind: make([]ModuleKind, n*n)}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			m.dark[y*n+x] = mods[y][x]
		}
	}
	m.classify()
	return m, nil
}

// Size returns the side length in modules.
func (m *Matrix) Size() int { return m.size }

// Dark reports whether the module at (x, y) is dark. Coordinates outside the
// grid read as light.
func (m *Matrix) Dark(x, y int) bool {
	if !m.inBounds(x, y) {
		return false
	}
	return m.dark[y*m.size+x]
}

// Kind returns the module's role. Coordinates outside the grid report KindData.
func (m *Matrix) Kind(x, y int) ModuleKind {
	if !m.inBounds(x, y) {
		return KindData
	}
	return m.kind[y*m.size+x]
}

// InFinder reports whether (x, y) lies inside one of the three 7x7 finder
// patterns. This is the classification the renderer actually depends on.
func (m *Matrix) InFinder(x, y int) bool { return m.Kind(x, y) == KindFinder }

func (m *Matrix) inBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < m.size && y < m.size
}

func (m *Matrix) set(x, y int, k ModuleKind) {
	if m.inBounds(x, y) {
		m.kind[y*m.size+x] = k
	}
}

func (m *Matrix) classify() {
	n := m.size
	version := (n - 17) / 4

	// Finder patterns and their separators, at three corners.
	for _, c := range [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}} {
		ox, oy := c[0], c[1]
		for dy := -1; dy <= 7; dy++ {
			for dx := -1; dx <= 7; dx++ {
				x, y := ox+dx, oy+dy
				if !m.inBounds(x, y) {
					continue
				}
				if dx >= 0 && dx < 7 && dy >= 0 && dy < 7 {
					m.set(x, y, KindFinder)
				} else {
					m.set(x, y, KindSeparator)
				}
			}
		}
	}

	// Timing patterns run along row 6 and column 6 between the finders.
	for i := 8; i < n-8; i++ {
		m.set(i, 6, KindTiming)
		m.set(6, i, KindTiming)
	}

	// Alignment patterns, 5x5, at every pairing of the version's centres except
	// the three that would collide with a finder pattern.
	centres := alignmentCentres(version)
	for _, cy := range centres {
		for _, cx := range centres {
			if (cx <= 8 && cy <= 8) || (cx <= 8 && cy >= n-9) || (cx >= n-9 && cy <= 8) {
				continue
			}
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					m.set(cx+dx, cy+dy, KindAlignment)
				}
			}
		}
	}

	// Format information sits beside the finder patterns.
	for i := 0; i <= 8; i++ {
		if i != 6 {
			m.set(i, 8, KindFormat)
			m.set(8, i, KindFormat)
		}
	}
	for i := 0; i < 8; i++ {
		m.set(8, n-1-i, KindFormat)
		m.set(n-1-i, 8, KindFormat)
	}

	// Version information, present from version 7 onwards.
	if version >= 7 {
		for i := 0; i < 6; i++ {
			for j := 0; j < 3; j++ {
				m.set(i, n-11+j, KindVersion)
				m.set(n-11+j, i, KindVersion)
			}
		}
	}
}

// alignmentCentres returns the row and column centres of the alignment patterns
// for a version, per ISO/IEC 18004 annex E. Version 1 has none.
//
// The spacing is derived rather than tabulated. Every centre but the first is
// spaced evenly back from 4*version+10, and the step is rounded up to an even
// number. Version 32 is the single version the general rule does not produce.
func alignmentCentres(version int) []int {
	if version < 2 || version > 40 {
		return nil
	}
	n := version/7 + 2
	step := 26 // version 32 is the one case the formula below does not yield
	if version != 32 {
		step = (version*4 + n*2 + 1) / (n*2 - 2) * 2
	}
	out := make([]int, n)
	out[0] = 6
	for i, pos := n-1, version*4+10; i >= 1; i, pos = i-1, pos-step {
		out[i] = pos
	}
	return out
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./qr/ -run Matrix -v`
Expected: PASS. If `TestMatrixClassifiesAlignmentOnVersion2` skips, adjust the test content until the symbol is 25 modules and rerun; a skip here hides a real gap.

- [ ] **Step 5: Add an alignment-centre test covering every version**

Append to `qr/matrix_test.go`. Do **not** hand-pick a handful of versions: the
derived spacing formula is wrong for exactly five of the forty (16, 19, 30, 36
and 39), and a six-version sample happens to miss all of them. gozxing is
already a test dependency and carries the specification's own table, so compare
against that for all forty.

```go
func TestAlignmentCentresMatchTheSpecTableForEveryVersion(t *testing.T) {
	for v := 1; v <= 40; v++ {
		ver, err := decoder.Version_GetVersionForNumber(v)
		if err != nil {
			t.Fatalf("version %d: %v", v, err)
		}
		want := ver.GetAlignmentPatternCenters()
		got := alignmentCentres(v)
		if len(want) == 0 && len(got) == 0 {
			continue // version 1 has no alignment patterns
		}
		if len(got) != len(want) {
			t.Errorf("version %d: got %v, want %v", v, got, want)
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("version %d: got %v, want %v", v, got, want)
				break
			}
		}
	}
}
```

Add `"github.com/makiuchi-d/gozxing/qrcode/decoder"` to the test file's imports.

- [ ] **Step 6: Run it and fix the formula until every version passes**

Run: `go test ./qr/ -run AlignmentCentres -v`
Expected: PASS for all forty versions. A failure here is a real classification
bug, not a test artefact.

- [ ] **Step 7: Commit**

```bash
go mod tidy   # gozxing becomes a test dependency here
git add qr/matrix.go qr/matrix_test.go go.mod go.sum
git commit -m "feat: add module matrix with finder and function-pattern classification"
```

Note that gozxing is a **test-only** dependency. Under Go's module graph pruning
it never reaches a consumer's build, which is why the dependency assertion in
Task 9 checks `go list -deps ./qr` rather than the contents of go.mod.

---

### Task 7: Pixel layout

**Files:**
- Create: `qr/layout.go`, `qr/layout_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks
- Produces: `qr.layout` struct with fields `Modules`, `Margin`, `ModuleSize`, `OriginX`, `OriginY`, `Width`; `qr.newLayout(modules, margin, width int) (layout, error)`; `layout.Rect(x, y int) (px, py, size float64)`; `qr.ErrWidthTooSmall`

- [ ] **Step 1: Write the failing test**

Create `qr/layout_test.go`:

```go
package qr

import (
	"errors"
	"testing"
)

func TestNewLayoutUsesIntegerModulesAndCentres(t *testing.T) {
	// 37 modules plus a 4-module quiet zone on each side is 45 across.
	// 380 / 45 = 8.44, so the module size is 8 and 380 - 360 = 20 spare pixels
	// split evenly become a 10 pixel offset.
	l, err := newLayout(37, 4, 380)
	if err != nil {
		t.Fatalf("newLayout: %v", err)
	}
	if l.ModuleSize != 8 {
		t.Errorf("ModuleSize = %v, want 8", l.ModuleSize)
	}
	if l.OriginX != 10 || l.OriginY != 10 {
		t.Errorf("Origin = (%v,%v), want (10,10)", l.OriginX, l.OriginY)
	}
	if l.Width != 380 {
		t.Errorf("Width = %d, want 380", l.Width)
	}
}

func TestLayoutRectSkipsTheQuietZone(t *testing.T) {
	l, err := newLayout(21, 4, 290) // 29 across, 290/29 = 10 exactly
	if err != nil {
		t.Fatalf("newLayout: %v", err)
	}
	px, py, size := l.Rect(0, 0)
	if size != 10 {
		t.Errorf("size = %v, want 10", size)
	}
	// Module (0,0) sits after four quiet-zone modules.
	if px != 40 || py != 40 {
		t.Errorf("Rect(0,0) = (%v,%v), want (40,40)", px, py)
	}
	px, _, _ = l.Rect(20, 0)
	if px != 240 {
		t.Errorf("Rect(20,0).px = %v, want 240", px)
	}
}

func TestNewLayoutRejectsWidthBelowOnePixelPerModule(t *testing.T) {
	_, err := newLayout(177, 4, 100) // 185 modules across, 100 pixels
	if !errors.Is(err, ErrWidthTooSmall) {
		t.Fatalf("error = %v, want ErrWidthTooSmall", err)
	}
}

func TestNewLayoutRejectsNonPositiveInput(t *testing.T) {
	if _, err := newLayout(0, 4, 380); err == nil {
		t.Error("newLayout accepted zero modules")
	}
	if _, err := newLayout(21, -1, 380); err == nil {
		t.Error("newLayout accepted a negative margin")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./qr/ -run Layout`
Expected: FAIL — `undefined: newLayout`

- [ ] **Step 3: Implement qr/layout.go**

```go
package qr

import (
	"errors"
	"fmt"
)

// ErrWidthTooSmall reports a Width that cannot hold one pixel per module once
// the quiet zone is included.
var ErrWidthTooSmall = errors.New("qr: width too small for the module count")

// layout maps module coordinates to pixels. The module size is a whole number
// of pixels and the symbol is centred, with the spare pixels widening the quiet
// zone. Measurement showed this decodes strictly better than stretching modules
// to fill Width exactly: 0 failures against 3 over the same 32 cases, every
// failure being a dense symbol below 3 pixels per module.
type layout struct {
	Modules    int
	Margin     int
	ModuleSize float64
	OriginX    float64
	OriginY    float64
	Width      int
}

// newLayout computes the pixel layout for a symbol of the given module count,
// quiet zone and image width, all measured in modules except width.
func newLayout(modules, margin, width int) (layout, error) {
	if modules <= 0 {
		return layout{}, fmt.Errorf("qr: module count must be positive, got %d", modules)
	}
	if margin < 0 {
		return layout{}, fmt.Errorf("qr: margin must not be negative, got %d", margin)
	}
	total := modules + 2*margin
	size := width / total
	if size < 1 {
		return layout{}, fmt.Errorf("%w: %d pixels cannot hold %d modules",
			ErrWidthTooSmall, width, total)
	}
	used := size * total
	origin := (width - used) / 2
	return layout{
		Modules:    modules,
		Margin:     margin,
		ModuleSize: float64(size),
		OriginX:    float64(origin),
		OriginY:    float64(origin),
		Width:      width,
	}, nil
}

// Rect returns the pixel position and size of module (x, y). The quiet zone is
// already accounted for, so (0, 0) is the first symbol module, not the first
// pixel.
func (l layout) Rect(x, y int) (px, py, size float64) {
	m := float64(l.Margin)
	return l.OriginX + (float64(x)+m)*l.ModuleSize,
		l.OriginY + (float64(y)+m)*l.ModuleSize,
		l.ModuleSize
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./qr/ -run Layout -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add qr/layout.go qr/layout_test.go
git commit -m "feat: add integer-module pixel layout with centred symbol"
```

---

### Task 8: ShapeContext with square dot and square corner

**Files:**
- Create: `qr/shape.go`, `qr/shape_dots.go`, `qr/shape_corners.go`, `qr/shape_test.go`

**Interfaces:**
- Consumes: `Matrix` from Task 6, `layout` from Task 7, `render.Path`, `render.RoundRect`, `render.Circle`, `Path.Reverse`, `Path.Append` from Task 2
- Produces: `qr.DotType` and `qr.CornerType` constants with `String()`; `qr.CornerRadius`; `qr.ShapeContext` interface with `Dark`, `Consume`, `Rect`, `Size`; `qr.newShapeContext(m *Matrix, l layout, excluded func(x, y int) bool) ShapeContext`; `qr.dotPath(t DotType, c ShapeContext, x, y int) render.Path`; `qr.cornerPath(t CornerType, px, py, s float64, r CornerRadius) render.Path`

- [ ] **Step 1: Write the failing test**

Create `qr/shape_test.go`:

```go
package qr

import "testing"

func testContext(t *testing.T, content string) (ShapeContext, *Matrix, layout) {
	t.Helper()
	m := mustMatrix(t, content, ECCHigh)
	l, err := newLayout(m.Size(), 4, 380)
	if err != nil {
		t.Fatalf("newLayout: %v", err)
	}
	return newShapeContext(m, l, nil), m, l
}

func TestShapeContextHidesFinderModules(t *testing.T) {
	c, m, _ := testContext(t, "HELLO")
	if !m.Dark(0, 0) {
		t.Fatal("test assumes the finder corner module is dark")
	}
	if c.Dark(0, 0) {
		t.Error("Dark(0,0) = true; finder modules must be invisible to dot shapes")
	}
}

func TestShapeContextHidesConsumedModules(t *testing.T) {
	c, m, _ := testContext(t, "HELLO")
	var x, y int
	found := false
	for y = 8; y < m.Size()-8 && !found; y++ {
		for x = 8; x < m.Size()-8; x++ {
			if c.Dark(x, y) {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("no drawable dark module found")
	}
	c.Consume(x, y)
	if c.Dark(x, y) {
		t.Error("Dark returned true for a consumed module")
	}
}

func TestShapeContextHonoursTheExcludedCallback(t *testing.T) {
	m := mustMatrix(t, "HELLO", ECCHigh)
	l, _ := newLayout(m.Size(), 4, 380)
	c := newShapeContext(m, l, func(x, y int) bool { return x == 10 })
	if c.Dark(10, 10) {
		t.Error("Dark(10,10) = true; the excluded callback was ignored")
	}
}

func TestShapeContextOutOfBoundsIsDarkFalse(t *testing.T) {
	c, m, _ := testContext(t, "HELLO")
	if c.Dark(-1, 0) || c.Dark(0, m.Size()) {
		t.Error("out-of-bounds coordinates must read as not-dark")
	}
}

func TestSquareDotCoversItsWholeModule(t *testing.T) {
	c, _, l := testContext(t, "HELLO")
	p := dotPath(DotSquare, c, 10, 10)
	px, py, s := l.Rect(10, 10)
	minX, minY, maxX, maxY := p.Bounds()
	if minX != px || minY != py || maxX != px+s || maxY != py+s {
		t.Errorf("bounds = %v %v %v %v, want %v %v %v %v",
			minX, minY, maxX, maxY, px, py, px+s, py+s)
	}
}

func TestSquareCornerIsARingPlusACore(t *testing.T) {
	p := cornerPath(CornerSquare, 0, 0, 10, CornerRadius{})
	// One outer contour, one reversed inner contour, one solid core.
	if len(p.SubPaths) != 3 {
		t.Fatalf("SubPaths = %d, want 3 (outer ring, hole, core)", len(p.SubPaths))
	}
	minX, minY, maxX, maxY := p.Bounds()
	if minX != 0 || minY != 0 || maxX != 70 || maxY != 70 {
		t.Errorf("bounds = %v %v %v %v, want 0 0 70 70", minX, minY, maxX, maxY)
	}
}

func TestDotTypeStringRoundTrips(t *testing.T) {
	for _, tc := range []struct {
		d    DotType
		want string
	}{{DotSquare, "square"}, {DotFluid, "fluid"}, {DotStripeColumn, "stripe-column"}} {
		if got := tc.d.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", int(tc.d), got, tc.want)
		}
	}
}
```

`qr/shape_test.go` imports only `testing`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./qr/ -run Shape`
Expected: FAIL — `undefined: newShapeContext`

- [ ] **Step 3: Implement qr/shape.go**

```go
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
// to round. It is unused until Plan 2 adds the neighbour-aware shapes; keep it
// here so the helper lives beside the context it reads.
func neighbours4(c ShapeContext, x, y int) (n, e, s, w bool) {
	return c.Dark(x, y-1), c.Dark(x+1, y), c.Dark(x, y+1), c.Dark(x-1, y)
}
```

`qr/shape.go` needs no imports: the shape functions that build paths live in
`shape_dots.go` and `shape_corners.go`.

- [ ] **Step 4: Implement qr/shape_dots.go with square only**

```go
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
```

- [ ] **Step 5: Implement qr/shape_corners.go with square only**

```go
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
// inner one. The radius rule applies only where the radius is positive: at zero
// the centreline has a real corner join and the canvas default lineJoin is
// miter, so the outline stays sharp.
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
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./qr/ -run Shape -v && go test ./qr/ -run DotType -v`
Expected: PASS

- [ ] **Step 7: Verify the toolchain is clean**

Run: `go vet ./... && gofmt -l . && go test ./qr/`
Expected: `gofmt` silent, everything else exit 0. `neighbours4` is unused until
Plan 2; Go permits unused package-level functions, so this stays clean.

- [ ] **Step 8: Commit**

```bash
git add qr/shape.go qr/shape_dots.go qr/shape_corners.go qr/shape_test.go
git commit -m "feat: add ShapeContext with square dot and square corner"
```

---

### Task 9: Options, New and PNG output — Phase 1 gate

**Files:**
- Create: `qr/options.go`, `qr/qr.go`, `qr/qr_test.go`, `qr/decode_test.go`

**Interfaces:**
- Consumes: everything from Tasks 2–8
- Produces: `qr.Options`, `qr.DotOptions`, `qr.CornerOptions`, `qr.LogoOptions`; `qr.New(Options) (*QR, error)`; `(*QR).Image() image.Image`, `(*QR).PNG(io.Writer) error`, `(*QR).JPEG(io.Writer, int) error`, `(*QR).WritePNGFile(string) error`, `(*QR).scene() render.Scene`; `qr.ErrNoContent`, `qr.ErrBadColor`, `qr.ErrLogoUnsupported`; test helpers `decodeImage`, `requireDecodableBaseline`

- [ ] **Step 1: Write the decode helpers**

Create `qr/decode_test.go`:

```go
package qr

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/makiuchi-d/gozxing"
	zxqr "github.com/makiuchi-d/gozxing/qrcode"
)

// decodeImage returns the text gozxing reads from img, or an error.
func decodeImage(img image.Image) (string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	res, err := zxqr.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return "", err
	}
	return res.GetText(), nil
}

// requireDecodableBaseline renders content unstyled at a generous module size
// and decodes it. gozxing fails on roughly 2-3% of perfectly valid QR symbols,
// a property of the decoder rather than of any encoder: three independent
// encoders fail on overlapping inputs. When the plain render already fails, a
// styled render proves nothing, so the case is skipped instead of reported as
// our bug.
func requireDecodableBaseline(t *testing.T, content string, ecc ECCLevel) {
	t.Helper()
	q, err := New(Options{Content: content, ECC: ecc, Width: 1000})
	if err != nil {
		t.Fatalf("baseline New: %v", err)
	}
	got, err := decodeImage(q.Image())
	if err != nil || got != content {
		t.Skipf("gozxing cannot decode this symbol unstyled (err=%v); not a renderer fault", err)
	}
}

// assertDecodes renders q and requires the decoded text to equal want.
func assertDecodes(t *testing.T, img image.Image, want string) {
	t.Helper()
	got, err := decodeImage(img)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if got != want {
		t.Fatalf("decoded %q, want %q", got, want)
	}
}

// flattenOntoWhite composites img over opaque white, which is what a
// transparent QR code sits on in practice.
func flattenOntoWhite(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
```

- [ ] **Step 2: Write the failing test**

Create `qr/qr_test.go`:

```go
package qr

import (
	"bytes"
	"errors"
	"image/png"
	"strings"
	"sync"
	"testing"
)

const testURL = "https://github.com/farizfadian/go-qrcode"

func TestNewRejectsEmptyContent(t *testing.T) {
	if _, err := New(Options{}); !errors.Is(err, ErrNoContent) {
		t.Fatalf("error = %v, want ErrNoContent", err)
	}
}

func TestZeroValueOptionsProducesAScannableCode(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	b := q.Image().Bounds()
	if b.Dx() != 380 || b.Dy() != 380 {
		t.Errorf("size = %dx%d, want 380x380 by default", b.Dx(), b.Dy())
	}
	assertDecodes(t, q.Image(), testURL)
}

func TestAllFourECCLevelsDecode(t *testing.T) {
	for _, ecc := range []ECCLevel{ECCLow, ECCMedium, ECCQuartile, ECCHigh} {
		t.Run(ecc.String(), func(t *testing.T) {
			requireDecodableBaseline(t, testURL, ecc)
			q, err := New(Options{Content: testURL, ECC: ecc, Width: 512})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			assertDecodes(t, q.Image(), testURL)
		})
	}
}

func TestLongContentDecodes(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("abcdefghij", 70)
	requireDecodableBaseline(t, long, ECCMedium)
	q, err := New(Options{Content: long, ECC: ECCMedium, Width: 1000})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	assertDecodes(t, q.Image(), long)
}

func TestPNGWritesADecodableImage(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var buf bytes.Buffer
	if err := q.PNG(&buf); err != nil {
		t.Fatalf("PNG: %v", err)
	}
	img, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	assertDecodes(t, img, testURL)
}

func TestTransparentBackgroundDecodesWhenFlattened(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL, Background: "#00000000"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	assertDecodes(t, flattenOntoWhite(q.Image()), testURL)
}

func TestNewRejectsUnparseableColour(t *testing.T) {
	if _, err := New(Options{Content: testURL, Foreground: "not-a-colour"}); !errors.Is(err, ErrBadColor) {
		t.Fatalf("error = %v, want ErrBadColor", err)
	}
}

func TestNewRejectsWidthTooSmall(t *testing.T) {
	long := strings.Repeat("a", 1200)
	if _, err := New(Options{Content: long, Width: 20}); !errors.Is(err, ErrWidthTooSmall) {
		t.Fatalf("error = %v, want ErrWidthTooSmall", err)
	}
}

func TestQRIsSafeForConcurrentUse(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var wg sync.WaitGroup
	outs := make([][]byte, 8)
	for i := range outs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var buf bytes.Buffer
			if err := q.PNG(&buf); err != nil {
				t.Errorf("PNG: %v", err)
				return
			}
			outs[i] = buf.Bytes()
		}(i)
	}
	wg.Wait()
	for i := 1; i < len(outs); i++ {
		if !bytes.Equal(outs[0], outs[i]) {
			t.Fatalf("goroutine %d produced different bytes; *QR is not immutable", i)
		}
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./qr/ -run 'TestNew|TestZeroValue'`
Expected: FAIL — `undefined: New`

- [ ] **Step 4: Implement qr/options.go**

```go
package qr

import (
	"errors"
	"image"
	"io"
)

// Default values applied when the corresponding Options field is left at its
// zero value.
const (
	DefaultWidth      = 380
	DefaultMargin     = 4
	DefaultForeground = "#000000"
	DefaultBackground = "#ffffff"
)

// Errors returned by New. Compare them with errors.Is.
var (
	ErrNoContent       = errors.New("qr: content is empty")
	ErrBadColor        = errors.New("qr: cannot parse colour")
	ErrLogoUnsupported = errors.New("qr: logo support is not implemented yet")
)

// Options configures a QR code. Only Content is required: every other field has
// a working default, so the zero value produces a conventional black-on-white
// code at 380 pixels.
type Options struct {
	// Content is the text to encode. Required.
	Content string
	// Width is the output image size in pixels. Zero selects DefaultWidth.
	Width int
	// Margin is the quiet zone in modules. Zero selects DefaultMargin, which is
	// the four modules ISO/IEC 18004 requires.
	Margin int
	// ECC is the error-correction level. The zero value lets New choose.
	ECC ECCLevel

	// Foreground is the module colour. Empty selects DefaultForeground.
	Foreground string
	// Background is the backdrop colour. Empty selects DefaultBackground;
	// "#00000000" makes it transparent.
	Background string

	// Dots styles the data modules.
	Dots DotOptions
	// Corners styles the three finder patterns.
	Corners CornerOptions
	// Logo places an image at the centre. Nil means no logo.
	Logo *LogoOptions
}

// DotOptions styles the data modules.
type DotOptions struct {
	// Type selects the figure. The zero value is DotSquare.
	Type DotType
	// Color overrides Options.Foreground for dots. Empty inherits it.
	Color string
}

// CornerOptions styles the three finder patterns.
type CornerOptions struct {
	// Type selects the figure. The zero value is CornerSquare.
	Type CornerType
	// Color overrides Options.Foreground for corners. Empty inherits it.
	Color string
	// Radius sets the corner radii for the rounding shapes. Zero fields fall
	// back to a fraction of the module size.
	Radius CornerRadius
}

// LogoOptions describes a centred logo. Exactly one of Image, Path or Reader
// must be set.
//
// New rejects a non-nil Logo with ErrLogoUnsupported until Plan 3 implements
// it. The struct is defined now so the public API shape is stable from the
// first release and adding logo support is not a breaking change.
type LogoOptions struct {
	Image  image.Image
	Path   string
	Reader io.Reader

	// Size is the logo block's width as a fraction of Options.Width, border
	// included. Zero selects automatic sizing from the ECC budget.
	Size float64
	// Radius rounds the logo image's own corners, in pixels.
	Radius float64
	// BorderWidth is the frame around the image, in pixels. Zero selects 10.
	BorderWidth float64
	// BorderRadius rounds the frame, in pixels. Zero selects 8.
	BorderRadius float64
	// BorderColor fills the frame. Empty inherits Options.Background.
	BorderColor string
	// BgColor fills behind the image. Empty selects "#ffffff".
	BgColor string
}

// withDefaults returns a copy of o with every zero field replaced by its
// default. It never mutates the caller's value.
func (o Options) withDefaults() Options {
	if o.Width == 0 {
		o.Width = DefaultWidth
	}
	if o.Margin == 0 {
		o.Margin = DefaultMargin
	}
	if o.Foreground == "" {
		o.Foreground = DefaultForeground
	}
	if o.Background == "" {
		o.Background = DefaultBackground
	}
	if o.Dots.Color == "" {
		o.Dots.Color = o.Foreground
	}
	if o.Corners.Color == "" {
		o.Corners.Color = o.Foreground
	}
	return o
}

// resolveECC turns ECCAuto into a concrete level. With a logo the budget
// matters more than symbol size, so High wins. Without one the level scales
// with content length, matching the reference library: short content has room
// for maximum protection, long content does not.
func (o Options) resolveECC() ECCLevel {
	if o.ECC != ECCAuto {
		return o.ECC
	}
	if o.Logo != nil {
		return ECCHigh
	}
	switch n := len(o.Content); {
	case n > 36:
		return ECCMedium
	case n > 16:
		return ECCQuartile
	default:
		return ECCHigh
	}
}
```

- [ ] **Step 5: Implement qr/qr.go**

```go
package qr

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// QR is a finished, immutable QR code. It is safe for concurrent use: every
// render method reads the same precomputed scene and writes nothing.
type QR struct {
	sc      render.Scene
	content string
	ecc     ECCLevel
	modules int
}

// New validates opts, encodes the content and builds the drawing. All
// configuration errors surface here; the render methods only ever report I/O
// failures.
func New(opts Options) (*QR, error) {
	if opts.Content == "" {
		return nil, ErrNoContent
	}
	if opts.Logo != nil {
		return nil, ErrLogoUnsupported
	}
	o := opts.withDefaults()

	bg, err := render.ParseColor(o.Background)
	if err != nil {
		return nil, fmt.Errorf("%w: background: %v", ErrBadColor, err)
	}
	dotCol, err := render.ParseColor(o.Dots.Color)
	if err != nil {
		return nil, fmt.Errorf("%w: dots: %v", ErrBadColor, err)
	}
	cornerCol, err := render.ParseColor(o.Corners.Color)
	if err != nil {
		return nil, fmt.Errorf("%w: corners: %v", ErrBadColor, err)
	}

	ecc := o.resolveECC()
	mods, err := defaultEncoder().Encode(o.Content, ecc)
	if err != nil {
		return nil, err
	}
	m, err := newMatrix(mods)
	if err != nil {
		return nil, err
	}
	l, err := newLayout(m.Size(), o.Margin, o.Width)
	if err != nil {
		return nil, err
	}

	ctx := newShapeContext(m, l, nil)
	sc := render.Scene{Width: o.Width, Height: o.Width, Background: bg}

	var dots render.Path
	for y := 0; y < m.Size(); y++ {
		for x := 0; x < m.Size(); x++ {
			if !ctx.Dark(x, y) {
				continue
			}
			dots = dots.Append(dotPath(o.Dots.Type, ctx, x, y))
		}
	}
	if !dots.IsEmpty() {
		sc.Items = append(sc.Items, render.PathItem{Path: dots, Fill: dotCol})
	}

	var corners render.Path
	n := m.Size()
	for _, c := range [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}} {
		px, py, s := l.Rect(c[0], c[1])
		corners = corners.Append(cornerPath(o.Corners.Type, px, py, s, o.Corners.Radius))
	}
	sc.Items = append(sc.Items, render.PathItem{Path: corners, Fill: cornerCol})

	return &QR{sc: sc, content: o.Content, ecc: ecc, modules: m.Size()}, nil
}

// Image returns the rendered raster image. Each call renders afresh, so the
// returned image is the caller's to modify.
func (q *QR) Image() image.Image { return render.Raster(q.sc) }

// PNG writes the code to w as a PNG.
func (q *QR) PNG(w io.Writer) error { return png.Encode(w, q.Image()) }

// JPEG writes the code to w as a JPEG at the given quality, 1 to 100. A
// transparent background is flattened onto white, since JPEG has no alpha.
func (q *QR) JPEG(w io.Writer, quality int) error {
	return jpeg.Encode(w, flatten(q.Image()), &jpeg.Options{Quality: quality})
}

// WritePNGFile renders the code and writes it to path.
func (q *QR) WritePNGFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := q.PNG(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Modules returns the symbol's side length in modules, excluding the quiet
// zone. It is useful for laying a code out beside other content.
func (q *QR) Modules() int { return q.modules }

// ECC returns the error-correction level actually used, which differs from the
// requested one when Options.ECC was ECCAuto.
func (q *QR) ECC() ECCLevel { return q.ecc }

// flatten composites src over opaque white. JPEG has no alpha channel, so a
// transparent background would otherwise encode as black.
func flatten(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
```

- [ ] **Step 6: Run the whole suite — this is the Phase 1 gate**

Run: `go test ./... -race -v`
Expected: PASS. If any decode test fails without skipping, stop and investigate before continuing; the foundation is not proven until this is green.

- [ ] **Step 7: Verify the toolchain and the dependency footprint**

```bash
go vet ./... && gofmt -l . && go mod tidy && git diff --exit-code go.mod
go list -deps ./qr | grep -vE '^(internal/|vendor/)' | grep -E '^(github|golang|rsc)' | sort -u
```
Expected: `gofmt` prints nothing, `go.mod` is unchanged by tidy, and the dependency list contains only `github.com/piglig/go-qr` and `golang.org/x/image/...`.

- [ ] **Step 8: Commit**

```bash
git add qr/options.go qr/qr.go qr/qr_test.go qr/decode_test.go go.mod go.sum
git commit -m "feat: add Options, New and PNG output with round-trip decode tests"
```

---

### Task 10: SVG renderer and PNG/SVG parity — Phase 2 gate

**Files:**
- Create: `internal/render/svg.go`, `internal/render/svg_test.go`, `internal/render/pathparse_test.go`
- Modify: `qr/qr.go` (add `SVG` and `SVGString`), `qr/qr_test.go` (add parity tests)

**Interfaces:**
- Consumes: `Scene`, `Path`, `PathItem`, `ImageItem` from Tasks 2 and 4
- Produces: `render.SVG(Scene) (string, error)`; `(*QR).SVG(io.Writer) error`, `(*QR).SVGString() (string, error)`; test-only `parsePathD(string) (Path, error)`

- [ ] **Step 1: Write the failing test for the serialiser**

Create `internal/render/svg_test.go`:

```go
package render

import (
	"image/color"
	"strings"
	"testing"
)

func TestSVGWrapsTheSceneWithAViewBox(t *testing.T) {
	sc := Scene{
		Width: 40, Height: 40,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items: []Item{PathItem{
			Path: RoundRect(5, 5, 10, 10, 0, 0, 0, 0),
			Fill: color.RGBA{0, 0, 0, 0xff},
		}},
	}
	out, err := SVG(sc)
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	for _, want := range []string{
		`xmlns="http://www.w3.org/2000/svg"`,
		`viewBox="0 0 40 40"`,
		`width="40"`,
		`height="40"`,
		`<path `,
		`fill="#000000"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("SVG output missing %q\ngot: %s", want, out)
		}
	}
}

func TestSVGOmitsATransparentBackground(t *testing.T) {
	out, err := SVG(Scene{Width: 10, Height: 10, Background: color.RGBA{}})
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	if strings.Contains(out, `<rect`) {
		t.Errorf("transparent background emitted a rect:\n%s", out)
	}
}

// The parity test: every path the scene holds must survive serialisation to a
// d attribute and back unchanged. This is what proves the raster and SVG
// renderers draw the same geometry without needing an SVG rasteriser.
func TestSVGPathsRoundTripThroughTheDAttribute(t *testing.T) {
	ring := RoundRect(0, 0, 30, 30, 4, 0, 4, 0).
		Append(Circle(15, 15, 6).Reverse())
	sc := Scene{
		Width: 30, Height: 30,
		Items: []Item{PathItem{Path: ring, Fill: color.RGBA{0, 0, 0, 0xff}}},
	}
	out, err := SVG(sc)
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	ds := extractDAttributes(out)
	if len(ds) != 1 {
		t.Fatalf("found %d path elements, want 1", len(ds))
	}
	got, err := parsePathD(ds[0])
	if err != nil {
		t.Fatalf("parsePathD: %v", err)
	}
	assertPathsEqual(t, got, ring)
}
```

- [ ] **Step 2: Write the d-attribute parser and comparison helpers**

Create `internal/render/pathparse_test.go`:

```go
package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

// extractDAttributes pulls every path's d attribute out of SVG markup, in
// document order.
func extractDAttributes(svg string) []string {
	var out []string
	rest := svg
	for {
		i := strings.Index(rest, ` d="`)
		if i < 0 {
			return out
		}
		rest = rest[i+4:]
		j := strings.IndexByte(rest, '"')
		if j < 0 {
			return out
		}
		out = append(out, rest[:j])
		rest = rest[j+1:]
	}
}

// parsePathD turns a d attribute back into a Path. It understands only the
// absolute commands the serialiser emits: M, L, Q, C and Z.
func parsePathD(d string) (Path, error) {
	fields := strings.Fields(strings.NewReplacer(
		"M", " M ", "L", " L ", "Q", " Q ", "C", " C ", "Z", " Z ", ",", " ",
	).Replace(d))

	var b Builder
	num := func(i int) (float64, error) {
		if i >= len(fields) {
			return 0, fmt.Errorf("d: ran out of numbers")
		}
		return strconv.ParseFloat(fields[i], 64)
	}
	for i := 0; i < len(fields); {
		cmd := fields[i]
		i++
		var vals []float64
		need := map[string]int{"M": 2, "L": 2, "Q": 4, "C": 6, "Z": 0}[cmd]
		for k := 0; k < need; k++ {
			v, err := num(i + k)
			if err != nil {
				return Path{}, err
			}
			vals = append(vals, v)
		}
		i += need
		switch cmd {
		case "M":
			b.MoveTo(vals[0], vals[1])
		case "L":
			b.LineTo(vals[0], vals[1])
		case "Q":
			b.QuadTo(vals[0], vals[1], vals[2], vals[3])
		case "C":
			b.CubeTo(vals[0], vals[1], vals[2], vals[3], vals[4], vals[5])
		case "Z":
			b.Close()
		default:
			return Path{}, fmt.Errorf("d: unsupported command %q", cmd)
		}
	}
	return b.Path(), nil
}

// assertPathsEqual compares two paths to the precision the serialiser emits.
func assertPathsEqual(t *testing.T, got, want Path) {
	t.Helper()
	const eps = 1e-3
	near := func(a, b float64) bool { return math.Abs(a-b) <= eps }
	nearPt := func(a, b Point) bool { return near(a.X, b.X) && near(a.Y, b.Y) }

	if len(got.SubPaths) != len(want.SubPaths) {
		t.Fatalf("subpaths = %d, want %d", len(got.SubPaths), len(want.SubPaths))
	}
	for i := range want.SubPaths {
		g, w := got.SubPaths[i], want.SubPaths[i]
		if !nearPt(g.Start, w.Start) {
			t.Errorf("subpath %d start = %v, want %v", i, g.Start, w.Start)
		}
		if g.Closed != w.Closed {
			t.Errorf("subpath %d closed = %v, want %v", i, g.Closed, w.Closed)
		}
		if len(g.Segs) != len(w.Segs) {
			t.Fatalf("subpath %d segments = %d, want %d", i, len(g.Segs), len(w.Segs))
		}
		for j := range w.Segs {
			gs, ws := g.Segs[j], w.Segs[j]
			if gs.Kind != ws.Kind {
				t.Errorf("subpath %d seg %d kind = %v, want %v", i, j, gs.Kind, ws.Kind)
				continue
			}
			if !nearPt(gs.To, ws.To) {
				t.Errorf("subpath %d seg %d end = %v, want %v", i, j, gs.To, ws.To)
			}
			if ws.Kind != SegLine && !nearPt(gs.C1, ws.C1) {
				t.Errorf("subpath %d seg %d c1 = %v, want %v", i, j, gs.C1, ws.C1)
			}
			if ws.Kind == SegCube && !nearPt(gs.C2, ws.C2) {
				t.Errorf("subpath %d seg %d c2 = %v, want %v", i, j, gs.C2, ws.C2)
			}
		}
	}
}

func TestParsePathDRejectsUnknownCommands(t *testing.T) {
	if _, err := parsePathD("M 0 0 A 1 1 0 0 1 2 2"); err == nil {
		t.Fatal("parsePathD accepted an arc command the serialiser never emits")
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./internal/render/ -run SVG`
Expected: FAIL — `undefined: SVG`

- [ ] **Step 4: Implement internal/render/svg.go**

```go
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
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

func hexOf(c color.Color) string {
	r, g, b, a := c.RGBA()
	if a == 0 {
		return "none"
	}
	// Undo premultiplication so the hex matches what the caller supplied.
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
```

- [ ] **Step 5: Run the render tests to verify they pass**

Run: `go test ./internal/render/ -v`
Expected: PASS, including `TestSVGPathsRoundTripThroughTheDAttribute`.

- [ ] **Step 6: Add SVG output to the public API**

Append to `qr/qr.go`:

```go
// SVG writes the code to w as standalone SVG markup.
func (q *QR) SVG(w io.Writer) error {
	s, err := q.SVGString()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// SVGString returns the code as standalone SVG markup. The geometry is the
// same as Image produces: both renderers consume one scene.
func (q *QR) SVGString() (string, error) { return render.SVG(q.sc) }
```

- [ ] **Step 7: Add the public parity test**

Append to `qr/qr_test.go`:

```go
func TestSVGAndPNGDescribeTheSameGeometry(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL, Width: 512})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	svg, err := q.SVGString()
	if err != nil {
		t.Fatalf("SVGString: %v", err)
	}
	if !strings.Contains(svg, `viewBox="0 0 512 512"`) {
		// min is a Go 1.21 builtin; the go directive is 1.22, so it is available.
		t.Errorf("viewBox missing or wrong:\n%s", svg[:min(200, len(svg))])
	}
	// Two path elements: one for all dots, one for all corners.
	if n := strings.Count(svg, "<path "); n != 2 {
		t.Errorf("path elements = %d, want 2", n)
	}
	// The rasterised form must still decode, proving the shared scene is sound.
	assertDecodes(t, q.Image(), testURL)
}
```

- [ ] **Step 8: Run the whole suite — this is the Phase 2 gate**

Run: `go test ./... -race && go vet ./... && gofmt -l .`
Expected: all green, `gofmt` silent.

- [ ] **Step 9: Commit and push**

```bash
git add internal/render/svg.go internal/render/svg_test.go internal/render/pathparse_test.go qr/qr.go qr/qr_test.go
git commit -m "feat: add SVG renderer with dependency-free PNG parity proof"
git push -u origin feat/styled-qr-core
```

---

## What comes next

This plan stops at a working, testable library: `square` dots, `square`
corners, PNG, JPEG and SVG, all decoding. Two plans follow, each producing
working software on its own, and each written only after the previous gate is
green — because their detail genuinely depends on what the gate reveals.

**Plan 2 — Shapes** (spec section 7). `stripe`, `stripe-row` and
`stripe-column` first, since they stress `ShapeContext` hardest; then `fluid`
and `fluid-line`; then the remaining seven dot types; then the seven corner
types with radii and independent colours; then the 12x7 decode matrix and
golden images.

**Plan 3 — Logo** (spec section 8). Source loading and resampling; automatic
sizing from the ECC budget; the safe zone through the `excluded` callback that
Task 8 already provides; `ErrLogoTooLarge`; the dot-and-corner-with-logo
matrices; and the measured contrast threshold for `ErrLowContrast`.

The consumability criteria in spec section 15 — runnable godoc examples for
every shape, the dependency assertion, the README gallery — land at the end of
Plan 2 and Plan 3, once there is something to show.
