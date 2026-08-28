package qr

import (
	"strings"
	"testing"
)

// fakeContext is a ShapeContext built from an ASCII grid, so a run-merging test
// reads as the picture it is testing. '#' is a claimable dark module, '.' is
// not — which covers light modules, finder patterns and anything the logo
// hides, since ShapeContext deliberately folds all of those into one question.
type fakeContext struct {
	n        int
	dark     []bool
	consumed []bool
	module   float64
}

func newFakeContext(rows []string, moduleSize float64) *fakeContext {
	n := len(rows)
	for _, r := range rows {
		if len(r) != n {
			panic("fakeContext needs a square grid")
		}
	}
	f := &fakeContext{n: n, dark: make([]bool, n*n), consumed: make([]bool, n*n), module: moduleSize}
	for y, row := range rows {
		for x, ch := range row {
			f.dark[y*n+x] = ch == '#'
		}
	}
	return f
}

func (f *fakeContext) Size() int { return f.n }

func (f *fakeContext) Rect(x, y int) (float64, float64, float64) {
	return float64(x) * f.module, float64(y) * f.module, f.module
}

func (f *fakeContext) Dark(x, y int) bool {
	if x < 0 || y < 0 || x >= f.n || y >= f.n {
		return false
	}
	i := y*f.n + x
	return f.dark[i] && !f.consumed[i]
}

// Adjacent ignores consumption: a module that has already been drawn is still
// visually next to this one.
func (f *fakeContext) Adjacent(x, y int) bool {
	if x < 0 || y < 0 || x >= f.n || y >= f.n {
		return false
	}
	return f.dark[y*f.n+x]
}

func (f *fakeContext) Consume(x, y int) {
	if x >= 0 && y >= 0 && x < f.n && y < f.n {
		f.consumed[y*f.n+x] = true
	}
}

func (f *fakeContext) consumedCount() int {
	n := 0
	for _, c := range f.consumed {
		if c {
			n++
		}
	}
	return n
}

// bounds is a small helper so the expectations below read as geometry.
func bounds(t *testing.T, c ShapeContext, typ DotType, x, y int) (w, h float64) {
	t.Helper()
	p := dotPath(typ, c, x, y)
	if p.IsEmpty() {
		t.Fatalf("%v at (%d,%d) produced an empty path", typ, x, y)
	}
	minX, minY, maxX, maxY := p.Bounds()
	return maxX - minX, maxY - minY
}

func TestStripeMergesAHorizontalRunOfThree(t *testing.T) {
	// Three in a row, nothing else. The run should become one bar, not three
	// separate figures, and should claim all three modules.
	c := newFakeContext([]string{
		".....",
		".###.",
		".....",
		".....",
		".....",
	}, 10)

	// A 3-wide bar spans two module steps plus the capsule thickness:
	// 2*10 + 10/2 = 25 wide, 10/2 = 5 tall.
	w, h := bounds(t, c, DotStripe, 1, 1)
	if w != 25 || h != 5 {
		t.Errorf("bar = %v x %v, want 25 x 5", w, h)
	}
	if got := c.consumedCount(); got != 3 {
		t.Errorf("consumed %d modules, want 3", got)
	}
	// Every module of the run is now invisible to the main loop.
	for x := 1; x <= 3; x++ {
		if c.Dark(x, 1) {
			t.Errorf("module (%d,1) is still drawable after being merged", x)
		}
	}
}

func TestStripeMergesAVerticalRunOfThree(t *testing.T) {
	c := newFakeContext([]string{
		".....",
		".#...",
		".#...",
		".#...",
		".....",
	}, 10)

	w, h := bounds(t, c, DotStripe, 1, 1)
	if w != 5 || h != 25 {
		t.Errorf("bar = %v x %v, want 5 x 25", w, h)
	}
	if got := c.consumedCount(); got != 3 {
		t.Errorf("consumed %d modules, want 3", got)
	}
}

func TestStripeFallsBackToADotWhenNothingAdjoins(t *testing.T) {
	c := newFakeContext([]string{
		".....",
		"..#..",
		".....",
		".....",
		".....",
	}, 10)

	w, h := bounds(t, c, DotStripe, 2, 1)
	if w != 5 || h != 5 {
		t.Errorf("lone module = %v x %v, want 5 x 5 (a dot of diameter s/2)", w, h)
	}
	if got := c.consumedCount(); got != 1 {
		t.Errorf("consumed %d modules, want 1", got)
	}
}

func TestStripeRowNeverMergesVertically(t *testing.T) {
	// A vertical run only. stripe-row must not join it; it should fall back to
	// a single dot and claim one module.
	c := newFakeContext([]string{
		".....",
		".#...",
		".#...",
		".#...",
		".....",
	}, 10)

	w, h := bounds(t, c, DotStripeRow, 1, 1)
	if w != 5 || h != 5 {
		t.Errorf("stripe-row on a vertical run = %v x %v, want 5 x 5", w, h)
	}
	if got := c.consumedCount(); got != 1 {
		t.Errorf("consumed %d modules, want 1", got)
	}
}

func TestStripeColumnNeverMergesHorizontally(t *testing.T) {
	c := newFakeContext([]string{
		".....",
		".###.",
		".....",
		".....",
		".....",
	}, 10)

	w, h := bounds(t, c, DotStripeColumn, 1, 1)
	if w != 5 || h != 5 {
		t.Errorf("stripe-column on a horizontal run = %v x %v, want 5 x 5", w, h)
	}
	if got := c.consumedCount(); got != 1 {
		t.Errorf("consumed %d modules, want 1", got)
	}
}

func TestStripePrefersTheLongestRun(t *testing.T) {
	// Four in a row. The greedy order tries three first, so the first call
	// claims three and leaves the fourth for the next iteration.
	c := newFakeContext([]string{
		".....",
		"####.",
		".....",
		".....",
		".....",
	}, 10)

	w, _ := bounds(t, c, DotStripe, 0, 1)
	if w != 25 {
		t.Errorf("first bar width = %v, want 25 (a run of three)", w)
	}
	if !c.Dark(3, 1) {
		t.Fatal("the fourth module was claimed; it should be left for the next pass")
	}
	w2, _ := bounds(t, c, DotStripe, 3, 1)
	if w2 != 5 {
		t.Errorf("leftover = %v wide, want 5 (a single dot)", w2)
	}
	if got := c.consumedCount(); got != 4 {
		t.Errorf("consumed %d modules in total, want 4", got)
	}
}

// This is the reference library's bug, made impossible rather than merely
// fixed. There, the run test asked only whether a module was dark, so a run
// could grow straight through a finder pattern or the logo safe zone. Here
// '.' stands for every reason a module is unavailable, and the run stops.
func TestStripeRunStopsAtAnUnavailableModule(t *testing.T) {
	c := newFakeContext([]string{
		".....",
		"##.##",
		".....",
		".....",
		".....",
	}, 10)

	// Modules 0 and 1 are dark, 2 is not. The longest run is two, not three.
	w, _ := bounds(t, c, DotStripe, 0, 1)
	if w != 15 { // 1*10 + 5
		t.Errorf("bar width = %v, want 15 (a run of two, stopped by the gap)", w)
	}
	if got := c.consumedCount(); got != 2 {
		t.Errorf("consumed %d modules, want 2", got)
	}
}

func TestStripeShapesAreRegistered(t *testing.T) {
	for _, name := range []string{"stripe", "stripe-row", "stripe-column"} {
		if _, err := ParseDotType(name); err != nil {
			t.Errorf("ParseDotType(%q): %v", name, err)
		}
	}
	names := strings.Join(DotTypeNames(), ",")
	for _, want := range []string{"stripe", "stripe-row", "stripe-column"} {
		if !strings.Contains(names, want) {
			t.Errorf("DotTypeNames() = %v, missing %q", names, want)
		}
	}
}

// The whole point: a styled code that does not scan is a bug, not a style.
func TestStripeShapesDecode(t *testing.T) {
	for _, typ := range []DotType{DotStripe, DotStripeRow, DotStripeColumn} {
		t.Run(typ.String(), func(t *testing.T) {
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
