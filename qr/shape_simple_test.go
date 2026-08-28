package qr

import (
	"math"
	"testing"
)

// The stateless dot shapes: each one depends only on its own module, so a
// single isolated module is enough to pin its geometry. Dimensions come from
// the design spec, section 7.2, and are stated relative to the module size s.
func TestStatelessDotGeometry(t *testing.T) {
	const s = 100.0

	for _, tc := range []struct {
		typ      DotType
		wantW    float64
		wantH    float64
		why      string
		tolerant bool
	}{
		{DotSquare, s, s, "fills the module", false},
		{DotTile, s - 1, s - 1, "one pixel smaller, leaving a hairline gap", false},
		{DotDot, 0.8 * s, 0.8 * s, "a circle of diameter 0.8s", false},
		{DotDotSmall, 0.6 * s, 0.6 * s, "a circle of diameter 0.6s", false},
		{DotRounded, 0.75 * s, 0.75 * s, "a rounded square at 0.75s", false},
		{DotDiamond, s, s, "a square of 0.5s/sin45 rotated 45 degrees spans exactly s", true},
	} {
		t.Run(tc.typ.String(), func(t *testing.T) {
			c := newFakeContext([]string{
				"...",
				".#.",
				"...",
			}, s)
			p := dotPath(tc.typ, c, 1, 1)
			if p.IsEmpty() {
				t.Fatalf("%v produced an empty path", tc.typ)
			}
			minX, minY, maxX, maxY := p.Bounds()
			w, h := maxX-minX, maxY-minY

			eps := 1e-9
			if tc.tolerant {
				eps = 1e-6 // the diamond's side involves sin(45 degrees)
			}
			if math.Abs(w-tc.wantW) > eps || math.Abs(h-tc.wantH) > eps {
				t.Errorf("%v: %v x %v, want %v x %v (%s)",
					tc.typ, w, h, tc.wantW, tc.wantH, tc.why)
			}
			if c.consumedCount() != 1 {
				t.Errorf("%v consumed %d modules, want 1", tc.typ, c.consumedCount())
			}
		})
	}
}

// Every stateless shape must sit inside or centred on its module. A shape that
// drifts off-centre would misalign the whole grid.
func TestStatelessDotsAreCentredOnTheirModule(t *testing.T) {
	const s = 100.0
	for _, typ := range []DotType{DotSquare, DotDot, DotDotSmall, DotRounded, DotDiamond, DotStar} {
		t.Run(typ.String(), func(t *testing.T) {
			c := newFakeContext([]string{"...", ".#.", "..."}, s)
			px, py, _ := c.Rect(1, 1)
			wantCX, wantCY := px+s/2, py+s/2

			minX, minY, maxX, maxY := dotPath(typ, c, 1, 1).Bounds()
			cx, cy := (minX+maxX)/2, (minY+maxY)/2
			if math.Abs(cx-wantCX) > 1e-6 || math.Abs(cy-wantCY) > 1e-6 {
				t.Errorf("%v centred at (%v,%v), want (%v,%v)", typ, cx, cy, wantCX, wantCY)
			}
		})
	}
}

// tile is deliberately one pixel smaller than its module so neighbouring
// modules read as separate tiles rather than a solid block.
func TestTileLeavesAGapAgainstItsNeighbour(t *testing.T) {
	const s = 100.0
	c := newFakeContext([]string{"...", "##.", "..."}, s)

	_, _, aMaxX, _ := dotPath(DotTile, c, 0, 1).Bounds()
	bMinX, _, _, _ := dotPath(DotTile, c, 1, 1).Bounds()
	if gap := bMinX - aMaxX; gap != 1 {
		t.Errorf("gap between adjacent tiles = %v, want 1 pixel", gap)
	}
}

func TestStatelessShapesAreRegisteredAndDecode(t *testing.T) {
	for _, typ := range []DotType{DotDot, DotDotSmall, DotTile, DotRounded, DotDiamond, DotStar} {
		t.Run(typ.String(), func(t *testing.T) {
			if _, err := ParseDotType(typ.String()); err != nil {
				t.Fatalf("ParseDotType(%q): %v", typ, err)
			}
			requireDecodableBaseline(t, testURL, ECCHigh)
			q, err := New(Options{
				Content: testURL,
				ECC:     ECCHigh,
				Width:   640,
				Dots:    DotOptions{Type: typ},
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			assertDecodes(t, q.Image(), testURL)
		})
	}
}
