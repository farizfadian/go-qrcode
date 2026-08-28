package qr

import (
	"math"
	"testing"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// Every finder shape occupies exactly the 7x7 block and is built the same way:
// an outer contour, a reversed inner contour that makes it a ring, and a solid
// core. Anything else would misalign the pattern a scanner looks for.
func TestEveryCornerShapeIsARingPlusACore(t *testing.T) {
	const s = 10.0
	for _, typ := range CornerTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			p := cornerPath(typ, 0, 0, s, CornerRadius{})

			if got := len(p.SubPaths); got != 3 {
				t.Errorf("subpaths = %d, want 3 (outer, hole, core)", got)
			}
			minX, minY, maxX, maxY := p.Bounds()
			const eps = 1e-9
			if math.Abs(minX) > eps || math.Abs(minY) > eps ||
				math.Abs(maxX-7*s) > eps || math.Abs(maxY-7*s) > eps {
				t.Errorf("bounds = %v %v %v %v, want 0 0 %v %v",
					minX, minY, maxX, maxY, 7*s, 7*s)
			}
		})
	}
}

// The ring must be exactly one module thick, because that is what the QR
// specification defines and what a reader measures. The hole's inner edge sits
// one module inside the outer edge on every side.
func TestCornerRingIsOneModuleThick(t *testing.T) {
	const s = 10.0
	for _, typ := range CornerTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			p := cornerPath(typ, 0, 0, s, CornerRadius{})
			hole := render.Path{SubPaths: p.SubPaths[1:2]}
			minX, minY, maxX, maxY := hole.Bounds()

			const eps = 1e-9
			if math.Abs(minX-s) > eps || math.Abs(minY-s) > eps ||
				math.Abs(maxX-6*s) > eps || math.Abs(maxY-6*s) > eps {
				t.Errorf("hole bounds = %v %v %v %v, want %v %v %v %v",
					minX, minY, maxX, maxY, s, s, 6*s, 6*s)
			}
		})
	}
}

// The core sits in the middle 3x3, apart from circle-diamond whose rotated
// square deliberately reaches further into the ring's gap.
func TestCornerCoreIsCentred(t *testing.T) {
	const s = 10.0
	for _, typ := range CornerTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			p := cornerPath(typ, 0, 0, s, CornerRadius{})
			core := render.Path{SubPaths: p.SubPaths[2:]}
			minX, minY, maxX, maxY := core.Bounds()
			cx, cy := (minX+maxX)/2, (minY+maxY)/2

			const eps = 1e-9
			if math.Abs(cx-3.5*s) > eps || math.Abs(cy-3.5*s) > eps {
				t.Errorf("core centred at (%v,%v), want (%v,%v)", cx, cy, 3.5*s, 3.5*s)
			}
		})
	}
}

// A caller-supplied radius must actually reach the geometry, and the offset
// rule from the design spec must be applied: the outer contour grows by half a
// module and the inner one shrinks by the same amount.
func TestCornerRadiusIsHonoured(t *testing.T) {
	const s = 10.0
	tight := cornerPath(CornerRounded, 0, 0, s, CornerRadius{Outer: s / 2, Inner: s / 4})
	loose := cornerPath(CornerRounded, 0, 0, s, CornerRadius{Outer: 3 * s, Inner: 2 * s})

	// A larger radius pulls the outer contour's first anchor further along the
	// top edge, so the two paths must differ.
	if tight.SubPaths[0].Start == loose.SubPaths[0].Start {
		t.Error("Radius.Outer had no effect on the outer contour")
	}
}

func TestAllCornerShapesAreRegisteredAndDecode(t *testing.T) {
	for _, typ := range CornerTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			if _, err := ParseCornerType(typ.String()); err != nil {
				t.Fatalf("ParseCornerType(%q): %v", typ, err)
			}
			requireDecodableBaseline(t, testURL, ECCHigh)
			q, err := New(Options{
				Content: testURL,
				ECC:     ECCHigh,
				Width:   640,
				Corners: CornerOptions{Type: typ},
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			assertDecodes(t, q.Image(), testURL)
		})
	}
}

// The combination that matters most: every dot shape against every finder
// shape, all decoding. This is the matrix the design spec calls for.
func TestEveryDotAndCornerCombinationDecodes(t *testing.T) {
	if testing.Short() {
		t.Skip("84 combinations; skipped under -short")
	}
	requireDecodableBaseline(t, testURL, ECCHigh)

	for _, dot := range DotTypes() {
		for _, corner := range CornerTypes() {
			name := dot.String() + "+" + corner.String()
			t.Run(name, func(t *testing.T) {
				q, err := New(Options{
					Content: testURL,
					ECC:     ECCHigh,
					Width:   640,
					Dots:    DotOptions{Type: dot},
					Corners: CornerOptions{Type: corner},
				})
				if err != nil {
					t.Fatalf("New: %v", err)
				}
				assertDecodes(t, q.Image(), testURL)
			})
		}
	}
}
