package render

import (
	"image/color"
	"strings"
	"testing"
)

func TestSVGWrapsTheSceneWithAViewBox(t *testing.T) {
	sc := Scene{
		Width: 40, Height: 40,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items: []Item{PathItem{
			Path: RoundRect(5, 5, 10, 10, 0, 0, 0, 0),
			Fill: color.RGBA{0, 0, 0, 0xff},
		}},
	}
	out, err := SVG(sc)
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	for _, want := range []string{
		`xmlns="http://www.w3.org/2000/svg"`,
		`viewBox="0 0 40 40"`,
		`width="40"`,
		`height="40"`,
		`<path `,
		`fill="#000000"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("SVG output missing %q\ngot: %s", want, out)
		}
	}
}

func TestSVGOmitsATransparentBackground(t *testing.T) {
	out, err := SVG(Scene{Width: 10, Height: 10, Background: color.RGBA{}})
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	if strings.Contains(out, `<rect`) {
		t.Errorf("transparent background emitted a rect:\n%s", out)
	}
}

// The parity test: every path the scene holds must survive serialisation to a
// d attribute and back unchanged. That is what proves the raster and SVG
// renderers draw the same geometry, without needing an SVG rasteriser whose own
// bugs could be mistaken for ours.
func TestSVGPathsRoundTripThroughTheDAttribute(t *testing.T) {
	ring := RoundRect(0, 0, 30, 30, 4, 0, 4, 0).
		Append(Circle(15, 15, 6).Reverse())
	sc := Scene{
		Width: 30, Height: 30,
		Items: []Item{PathItem{Path: ring, Fill: color.RGBA{0, 0, 0, 0xff}}},
	}
	out, err := SVG(sc)
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	ds := extractDAttributes(out)
	if len(ds) != 1 {
		t.Fatalf("found %d path elements, want 1", len(ds))
	}
	got, err := parsePathD(ds[0])
	if err != nil {
		t.Fatalf("parsePathD: %v", err)
	}
	assertPathsEqual(t, got, ring)
}

func TestSVGRoundTripsEverySegmentKind(t *testing.T) {
	var b Builder
	b.MoveTo(1, 1)
	b.LineTo(9, 1)
	b.QuadTo(9, 5, 5, 9)
	b.CubeTo(3, 9, 1, 7, 1, 5)
	b.Close()
	want := b.Path()

	out, err := SVG(Scene{Width: 10, Height: 10,
		Items: []Item{PathItem{Path: want, Fill: color.RGBA{0, 0, 0, 0xff}}}})
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	got, err := parsePathD(extractDAttributes(out)[0])
	if err != nil {
		t.Fatalf("parsePathD: %v", err)
	}
	assertPathsEqual(t, got, want)
}
