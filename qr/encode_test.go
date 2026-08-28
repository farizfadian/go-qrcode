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
