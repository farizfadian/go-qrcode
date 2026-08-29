package qr

import (
	"bytes"
	"flag"
	"fmt"
	"image"
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

// comparePixels returns a description of the first difference, or "" when the
// images match.
func comparePixels(got, want image.Image) string {
	gb, wb := got.Bounds(), want.Bounds()
	if gb != wb {
		return fmt.Sprintf("bounds %v, want %v", gb, wb)
	}
	differing := 0
	first := ""
	for y := gb.Min.Y; y < gb.Max.Y; y++ {
		for x := gb.Min.X; x < gb.Max.X; x++ {
			gr, gg, gbl, ga := got.At(x, y).RGBA()
			wr, wg, wbl, wa := want.At(x, y).RGBA()
			if gr != wr || gg != wg || gbl != wbl || ga != wa {
				differing++
				if first == "" {
					first = fmt.Sprintf("first at (%d,%d): got %v, want %v",
						x, y, got.At(x, y), want.At(x, y))
				}
			}
		}
	}
	if differing == 0 {
		return ""
	}
	total := gb.Dx() * gb.Dy()
	return fmt.Sprintf("%d of %d pixels differ (%.2f%%); %s",
		differing, total, 100*float64(differing)/float64(total), first)
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
