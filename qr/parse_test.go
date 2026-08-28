package qr

import (
	"errors"
	"strings"
	"testing"
)

func TestDotTypesListsOnlyImplementedShapes(t *testing.T) {
	got := DotTypes()
	if len(got) == 0 {
		t.Fatal("DotTypes returned nothing")
	}
	for _, d := range got {
		if _, ok := dotFuncs[d]; !ok {
			t.Errorf("DotTypes includes %v, which has no implementation", d)
		}
	}
	if len(got) != len(dotFuncs) {
		t.Errorf("DotTypes returned %d shapes, but %d are registered", len(got), len(dotFuncs))
	}
}

func TestCornerTypesListsOnlyImplementedShapes(t *testing.T) {
	got := CornerTypes()
	if len(got) != len(cornerFuncs) {
		t.Errorf("CornerTypes returned %d shapes, but %d are registered", len(got), len(cornerFuncs))
	}
	for _, c := range got {
		if _, ok := cornerFuncs[c]; !ok {
			t.Errorf("CornerTypes includes %v, which has no implementation", c)
		}
	}
}

func TestParseDotType(t *testing.T) {
	got, err := ParseDotType("square")
	if err != nil {
		t.Fatalf("ParseDotType(\"square\"): %v", err)
	}
	if got != DotSquare {
		t.Errorf("got %v, want DotSquare", got)
	}
	if _, err := ParseDotType("SQUARE"); err != nil {
		t.Errorf("ParseDotType is case sensitive: %v", err)
	}
}

func TestParseDotTypeRejectsNonsenseAndListsWhatIsAvailable(t *testing.T) {
	// The rejection must say what the caller could have asked for, otherwise a
	// typo leaves them guessing. The list comes from the registry, so it stays
	// correct as shapes are added.
	_, err := ParseDotType("banana")
	if !errors.Is(err, ErrUnknownShape) {
		t.Fatalf("error = %v, want ErrUnknownShape", err)
	}
	if !strings.Contains(err.Error(), "square") {
		t.Errorf("error does not list the available shapes: %v", err)
	}
}

func TestParseCornerType(t *testing.T) {
	got, err := ParseCornerType("square")
	if err != nil {
		t.Fatalf("ParseCornerType: %v", err)
	}
	if got != CornerSquare {
		t.Errorf("got %v, want CornerSquare", got)
	}
	if _, err := ParseCornerType("banana"); !errors.Is(err, ErrUnknownShape) {
		t.Errorf("error = %v, want ErrUnknownShape", err)
	}
}

func TestParseECCLevel(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want ECCLevel
	}{
		{"auto", ECCAuto}, {"", ECCAuto}, {"L", ECCLow}, {"l", ECCLow},
		{"M", ECCMedium}, {"Q", ECCQuartile}, {"H", ECCHigh},
	} {
		got, err := ParseECCLevel(tc.in)
		if err != nil {
			t.Fatalf("ParseECCLevel(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseECCLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
	if _, err := ParseECCLevel("X"); err == nil {
		t.Error("ParseECCLevel accepted an unknown level")
	}
}

// An out-of-range enum value stands in for "a shape this build cannot draw".
// Naming a real-but-unimplemented shape here would work only until that shape
// landed, and then quietly stop testing anything.
func TestNewRejectsAnUnimplementedDotType(t *testing.T) {
	// Without this, New would silently fall back to square and hand back a code
	// that does not look like what was asked for.
	_, err := New(Options{Content: testURL, Dots: DotOptions{Type: DotType(99)}})
	if !errors.Is(err, ErrUnknownShape) {
		t.Fatalf("error = %v, want ErrUnknownShape", err)
	}
	if !strings.Contains(err.Error(), "DotType(99)") {
		t.Errorf("error does not name the offending value: %v", err)
	}
}

func TestNewRejectsAnUnimplementedCornerType(t *testing.T) {
	_, err := New(Options{Content: testURL, Corners: CornerOptions{Type: CornerType(99)}})
	if !errors.Is(err, ErrUnknownShape) {
		t.Fatalf("error = %v, want ErrUnknownShape", err)
	}
}
