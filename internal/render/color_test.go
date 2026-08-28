package render

import (
	"errors"
	"image/color"
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
