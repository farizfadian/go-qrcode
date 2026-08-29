package qr

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update", false, "regenerate the golden images in testdata/golden")

// goldenCases are deliberately short-content and small, so the stored files stay
// a few kilobytes and a human can open one and see what changed.
var goldenCases = []struct {
	name string
	opts Options
}{
	{"default", Options{Content: "HELLO", Width: 260}},
	{"dot", Options{Content: "HELLO", Width: 260, Dots: DotOptions{Type: DotDot}}},
	{"rounded", Options{Content: "HELLO", Width: 260, Dots: DotOptions{Type: DotRounded}}},
	{"star", Options{Content: "HELLO", Width: 260, Dots: DotOptions{Type: DotStar}}},
	{"fluid", Options{Content: "HELLO", Width: 260, Dots: DotOptions{Type: DotFluid}}},
	{"fluid-line", Options{Content: "HELLO", Width: 260, Dots: DotOptions{Type: DotFluidLine}}},
	{"stripe", Options{Content: "HELLO", Width: 260, Dots: DotOptions{Type: DotStripe}}},
	{"corner-circle", Options{Content: "HELLO", Width: 260, Corners: CornerOptions{Type: CornerCircle}}},
	{"corner-rounded", Options{Content: "HELLO", Width: 260, Corners: CornerOptions{Type: CornerRounded}}},
	{"corner-circle-diamond", Options{Content: "HELLO", Width: 260, Corners: CornerOptions{Type: CornerCircleDiamond}}},
	{"corner-circle-star", Options{Content: "HELLO", Width: 260, Corners: CornerOptions{Type: CornerCircleStar}}},
	{"coloured", Options{
		Content: "HELLO", Width: 260,
		Dots:    DotOptions{Color: "#1f2937"},
		Corners: CornerOptions{Color: "#dc2626"},
	}},
	{"transparent", Options{Content: "HELLO", Width: 260, Background: "#00000000"}},
	{"margin-8", Options{Content: "HELLO", Width: 260, Margin: 8}},
}

// TestGolden pins the rendered geometry. It compares decoded pixels rather than
// file bytes, because a change in Go's PNG encoder would break a byte
// comparison without anything about this library having changed.
//
// Regenerate with:
//
//	go test ./qr -run TestGolden -update
//
// and read the resulting diff as a picture: an unintended change is usually
// obvious at a glance, which is the whole value of these files.
func TestGolden(t *testing.T) {
	dir := filepath.Join("testdata", "golden")
	if *updateGolden {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, tc := range goldenCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := New(tc.opts)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			got := q.Image()
			path := filepath.Join(dir, tc.name+".png")

			if *updateGolden {
				var buf bytes.Buffer
				if err := png.Encode(&buf, got); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("wrote %s", path)
				return
			}

			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("%v\nrun: go test ./qr -run TestGolden -update", err)
			}
			defer f.Close()
			want, err := png.Decode(f)
			if err != nil {
				t.Fatalf("decoding %s: %v", path, err)
			}

			if diff := comparePixels(got, want); diff != "" {
				t.Errorf("%s differs from its golden image: %s\n"+
					"if the change is intended, run: go test ./qr -run TestGolden -update",
					tc.name, diff)
			}
		})
	}
}

// Golden comparison tolerances.
//
// Exact pixel equality is not portable. golang.org/x/image/vector ships an
// amd64 assembly rasteriser and a pure-Go fallback, so a curve's anti-aliased
// edge rounds a shade differently on arm64 than on amd64 — measured at 8 of
// 67,600 pixels, each off by one level, on the circular finder shapes. Straight
// edges are unaffected.
//
// The two regimes are far apart, so a tolerance sits comfortably between them:
// a real geometry change moves whole modules, which means thousands of pixels
// differing by hundreds of levels, not a dozen differing by one.
const (
	channelTolerance = 4     // per-channel levels (0-255) attributable to rounding
	maxDifferingFrac = 0.005 // at most 0.5% of pixels may differ at all
)

// comparePixels returns a description of how two images differ, or "" when they
// match within the tolerances above.
func comparePixels(got, want image.Image) string {
	gb, wb := got.Bounds(), want.Bounds()
	if gb != wb {
		return fmt.Sprintf("bounds %v, want %v", gb, wb)
	}

	differing, worst := 0, 0
	first := ""
	for y := gb.Min.Y; y < gb.Max.Y; y++ {
		for x := gb.Min.X; x < gb.Max.X; x++ {
			gr, gg, gbl, ga := got.At(x, y).RGBA()
			wr, wg, wbl, wa := want.At(x, y).RGBA()
			d := maxOf(
				diff8(gr, wr), diff8(gg, wg),
				diff8(gbl, wbl), diff8(ga, wa),
			)
			if d == 0 {
				continue
			}
			differing++
			if d > worst {
				worst = d
			}
			if first == "" {
				first = fmt.Sprintf("first at (%d,%d): got %v, want %v",
					x, y, got.At(x, y), want.At(x, y))
			}
		}
	}
	if differing == 0 {
		return ""
	}

	total := gb.Dx() * gb.Dy()
	frac := float64(differing) / float64(total)
	if worst <= channelTolerance && frac <= maxDifferingFrac {
		return "" // rasteriser rounding, not a change in geometry
	}
	return fmt.Sprintf(
		"%d of %d pixels differ (%.3f%%), worst channel gap %d levels; %s",
		differing, total, 100*frac, worst, first)
}

// diff8 returns the absolute difference between two 16-bit channel values,
// scaled to the 0-255 range the tolerance is expressed in.
func diff8(a, b uint32) int {
	d := int(a) - int(b)
	if d < 0 {
		d = -d
	}
	return d >> 8
}

func maxOf(vs ...int) int {
	m := 0
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

// Every golden image must also still scan. A regression that changed the
// geometry *and* was accepted with -update would otherwise go unnoticed.
func TestGoldenImagesStillDecode(t *testing.T) {
	for _, tc := range goldenCases {
		if tc.opts.Background == "#00000000" {
			continue // covered separately, after flattening
		}
		t.Run(tc.name, func(t *testing.T) {
			requireDecodableBaseline(t, tc.opts.Content, tc.opts.ECC)
			q, err := New(tc.opts)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			assertDecodes(t, q.Image(), tc.opts.Content)
		})
	}
}

// A tolerance that accepts everything is worse than no test, so both ends of it
// are pinned: a real change must still fail, and rasteriser rounding must still
// pass.
func TestGoldenToleranceCatchesRealChangesAndIgnoresRounding(t *testing.T) {
	base, err := New(Options{Content: "HELLO", Width: 260})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("a different dot shape is caught", func(t *testing.T) {
		other, err := New(Options{Content: "HELLO", Width: 260, Dots: DotOptions{Type: DotDot}})
		if err != nil {
			t.Fatal(err)
		}
		if diff := comparePixels(base.Image(), other.Image()); diff == "" {
			t.Error("the tolerance accepted a different dot shape")
		}
	})

	t.Run("a one-pixel shift is caught", func(t *testing.T) {
		shifted, err := New(Options{Content: "HELLO", Width: 261})
		if err != nil {
			t.Fatal(err)
		}
		if diff := comparePixels(base.Image(), shifted.Image()); diff == "" {
			t.Error("the tolerance accepted a different image size or offset")
		}
	})

	t.Run("rasteriser rounding is ignored", func(t *testing.T) {
		// Nudge a handful of pixels by one level, which is the shape of the
		// difference measured between the amd64 and pure-Go rasterisers.
		nudged := image.NewRGBA(base.Image().Bounds())
		draw.Draw(nudged, nudged.Bounds(), base.Image(), image.Point{}, draw.Src)
		for i := 0; i < 8; i++ {
			x, y := 40+i, 40
			c := nudged.RGBAAt(x, y)
			c.R, c.G, c.B = c.R-1, c.G-1, c.B-1
			nudged.SetRGBA(x, y, c)
		}
		if diff := comparePixels(nudged, base.Image()); diff != "" {
			t.Errorf("the tolerance rejected rounding-scale noise: %s", diff)
		}
	})
}
