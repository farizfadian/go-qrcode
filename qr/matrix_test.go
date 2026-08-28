package qr

import (
	"testing"

	"github.com/makiuchi-d/gozxing/qrcode/decoder"
)

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
	for _, c := range [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}} {
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
		t.Fatalf("expected a version 2 symbol, got %d modules", m.Size())
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

// Do not hand-pick a handful of versions here. The derived spacing formula is
// wrong for exactly five of the forty (16, 19, 30, 36 and 39), and a
// six-version sample happens to miss all of them. gozxing is already a test
// dependency and carries the specification's own table, so compare against
// that for every version.
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
