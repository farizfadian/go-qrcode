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
	var fx, fy int
	found := false
	for y := 8; y < m.Size()-8 && !found; y++ {
		for x := 8; x < m.Size()-8; x++ {
			if c.Dark(x, y) {
				fx, fy, found = x, y, true
				break
			}
		}
	}
	if !found {
		t.Fatal("no drawable dark module found")
	}
	c.Consume(fx, fy)
	if c.Dark(fx, fy) {
		t.Error("Dark returned true for a consumed module")
	}
}

func TestShapeContextHonoursTheExcludedCallback(t *testing.T) {
	m := mustMatrix(t, "HELLO", ECCHigh)
	l, err := newLayout(m.Size(), 4, 380)
	if err != nil {
		t.Fatalf("newLayout: %v", err)
	}
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

func TestCornerTypeStringRoundTrips(t *testing.T) {
	for _, tc := range []struct {
		c    CornerType
		want string
	}{{CornerSquare, "square"}, {CornerCircleStar, "circle-star"}} {
		if got := tc.c.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", int(tc.c), got, tc.want)
		}
	}
}
