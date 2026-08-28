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
