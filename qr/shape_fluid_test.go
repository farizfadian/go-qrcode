package qr

import (
	"testing"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// countCubics counts the curved segments in a path. A rounded rectangle emits
// exactly one cubic per rounded corner and none for a square one, so this
// counts how many corners a fluid module rounded.
func countCubics(p render.Path) int {
	n := 0
	for _, sp := range p.SubPaths {
		for _, seg := range sp.Segs {
			if seg.Kind == render.SegCube {
				n++
			}
		}
	}
	return n
}

func TestFluidRoundsEveryCornerOfAnIsolatedModule(t *testing.T) {
	c := newFakeContext([]string{
		"...",
		".#.",
		"...",
	}, 100)
	if got := countCubics(dotPath(DotFluid, c, 1, 1)); got != 4 {
		t.Errorf("rounded corners = %d, want 4 (an isolated module is a circle)", got)
	}
}

func TestFluidSquaresTheCornersFacingANeighbour(t *testing.T) {
	// Two modules side by side. The left one keeps its left corners round and
	// squares the two facing its neighbour, so the pair reads as one capsule.
	c := newFakeContext([]string{
		"...",
		"##.",
		"...",
	}, 100)
	if got := countCubics(dotPath(DotFluid, c, 0, 1)); got != 2 {
		t.Errorf("left module rounded %d corners, want 2", got)
	}
}

// This is the case that exposes the difference between "is dark" and "is
// claimable". By the time the main loop reaches the middle module of a
// horizontal run, the module to its west has already been drawn and consumed.
// It is still visually adjacent, so all four corners must stay square.
func TestFluidSeesNeighboursThatHaveAlreadyBeenDrawn(t *testing.T) {
	c := newFakeContext([]string{
		"....",
		"###.",
		"....",
		"....",
	}, 100)

	// Simulate the main loop having already drawn the module to the west.
	c.Consume(0, 1)

	if got := countCubics(dotPath(DotFluid, c, 1, 1)); got != 0 {
		t.Errorf("middle of a run rounded %d corners, want 0; "+
			"a neighbour that has already been drawn is still adjacent", got)
	}
}

func TestFluidLineAddsFilletsToLowerDiagonals(t *testing.T) {
	// A module with no neighbour below but a dark diagonal below-right. Plain
	// fluid draws one contour; fluid-line adds a fillet bridging the diagonal.
	rows := []string{
		"....",
		".#..",
		"..#.",
		"....",
	}

	plain := newFakeContext(rows, 100)
	withLine := newFakeContext(rows, 100)

	nPlain := len(dotPath(DotFluid, plain, 1, 1).SubPaths)
	nLine := len(dotPath(DotFluidLine, withLine, 1, 1).SubPaths)

	if nPlain != 1 {
		t.Fatalf("fluid produced %d subpaths, want 1", nPlain)
	}
	if nLine != 2 {
		t.Errorf("fluid-line produced %d subpaths, want 2 (body plus one fillet)", nLine)
	}
}

func TestFluidLineAddsNoFilletWhenSomethingSitsBelow(t *testing.T) {
	// The fillet only bridges a diagonal gap. With a module directly below
	// there is no gap to bridge.
	c := newFakeContext([]string{
		"....",
		".#..",
		".##.",
		"....",
	}, 100)
	if n := len(dotPath(DotFluidLine, c, 1, 1).SubPaths); n != 1 {
		t.Errorf("subpaths = %d, want 1; no fillet is needed above a neighbour", n)
	}
}

func TestFluidShapesAreRegisteredAndDecode(t *testing.T) {
	for _, typ := range []DotType{DotFluid, DotFluidLine} {
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
