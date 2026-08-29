package render

import (
	"errors"
	"image/color"
	"math"
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

func TestLuminanceMatchesTheWCAGReferencePoints(t *testing.T) {
	for _, tc := range []struct {
		c    color.RGBA
		want float64
	}{
		{color.RGBA{0, 0, 0, 0xff}, 0.0},
		{color.RGBA{0xff, 0xff, 0xff, 0xff}, 1.0},
		{color.RGBA{0xff, 0, 0, 0xff}, 0.2126},
		{color.RGBA{0, 0xff, 0, 0xff}, 0.7152},
		{color.RGBA{0, 0, 0xff, 0xff}, 0.0722},
	} {
		if got := Luminance(tc.c); math.Abs(got-tc.want) > 1e-4 {
			t.Errorf("Luminance(%v) = %.4f, want %.4f", tc.c, got, tc.want)
		}
	}
}

func TestContrastRatioIsSymmetricAndBounded(t *testing.T) {
	black := color.RGBA{0, 0, 0, 0xff}
	white := color.RGBA{0xff, 0xff, 0xff, 0xff}

	if got := ContrastRatio(black, white); math.Abs(got-21) > 1e-6 {
		t.Errorf("black against white = %.4f, want 21", got)
	}
	if a, b := ContrastRatio(black, white), ContrastRatio(white, black); a != b {
		t.Errorf("not symmetric: %v vs %v", a, b)
	}
	if got := ContrastRatio(white, white); math.Abs(got-1) > 1e-9 {
		t.Errorf("a colour against itself = %.4f, want 1", got)
	}
}

func TestOverCompositesTransparencyOntoABackdrop(t *testing.T) {
	white := color.RGBA{0xff, 0xff, 0xff, 0xff}

	// A fully transparent source leaves the backdrop untouched.
	if got := Over(color.RGBA{}, white); got != white {
		t.Errorf("transparent over white = %v, want white", got)
	}
	// An opaque source replaces it.
	red := color.RGBA{0xff, 0, 0, 0xff}
	if got := Over(red, white); got != red {
		t.Errorf("opaque red over white = %v, want red", got)
	}
	// The result is always opaque, so it can be measured.
	half, err := ParseColor("#00000080")
	if err != nil {
		t.Fatal(err)
	}
	if got := Over(half, white); got.A != 0xff {
		t.Errorf("alpha = %d, want 255", got.A)
	}
}
