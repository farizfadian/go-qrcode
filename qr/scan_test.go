package qr

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// roundTrip is the property that matters: whatever this library renders, it can
// read back.
func TestScanReadsWhatThisLibraryRenders(t *testing.T) {
	q, err := New(Options{Content: testURL, Width: 600})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := Scan(q.Image())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if res.Content != testURL {
		t.Errorf("Content = %q, want %q", res.Content, testURL)
	}
	if res.Modules != q.Modules() {
		t.Errorf("Modules = %d, want %d", res.Modules, q.Modules())
	}
	if res.ECC != q.ECC() {
		t.Errorf("ECC = %v, want %v", res.ECC, q.ECC())
	}
	if res.Version != (q.Modules()-17)/4 {
		t.Errorf("Version = %d, inconsistent with %d modules", res.Version, res.Modules)
	}
	if res.Mask < 0 || res.Mask > 7 {
		t.Errorf("Mask = %d, want 0 to 7", res.Mask)
	}
	if len(res.Segments) == 0 {
		t.Error("no segments reported")
	}
}

// Styling must not stop the code being readable. Every shape is rendered and
// scanned back, which is the strongest statement this library can make about
// its own output.
func TestScanReadsEveryDotShape(t *testing.T) {
	for _, typ := range DotTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			q, err := New(Options{
				Content: testURL, Width: 700, ECC: ECCHigh,
				Dots: DotOptions{Type: typ},
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			res, err := Scan(q.Image())
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if res.Content != testURL {
				t.Errorf("Content = %q, want %q", res.Content, testURL)
			}
		})
	}
}

func TestScanReadsEveryCornerShape(t *testing.T) {
	for _, typ := range CornerTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			q, err := New(Options{
				Content: testURL, Width: 700, ECC: ECCHigh,
				Corners: CornerOptions{Type: typ},
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			res, err := Scan(q.Image())
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if res.Content != testURL {
				t.Errorf("Content = %q, want %q", res.Content, testURL)
			}
		})
	}
}

func TestScanReadsACodeWithALogo(t *testing.T) {
	q, err := New(Options{
		Content: testURL, Width: 900,
		Logo: &LogoOptions{Image: testLogo(128, 128)},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := Scan(q.Image())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if res.Content != testURL {
		t.Errorf("Content = %q, want %q", res.Content, testURL)
	}
}

// The reported metadata must match what was asked for, otherwise it is worse
// than not reporting it.
func TestScanReportsTheEncodingParameters(t *testing.T) {
	for _, ecc := range []ECCLevel{ECCLow, ECCMedium, ECCQuartile, ECCHigh} {
		t.Run(ecc.String(), func(t *testing.T) {
			q, err := New(Options{Content: testURL, Width: 700, ECC: ecc})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			res, err := Scan(q.Image())
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if res.ECC != ecc {
				t.Errorf("ECC = %v, want %v", res.ECC, ecc)
			}
		})
	}
}

// Numeric content should be encoded numerically, which is the visible sign that
// segment reporting is real rather than always saying "byte".
func TestScanReportsSegmentModes(t *testing.T) {
	for _, tc := range []struct {
		content string
		want    SegmentMode
	}{
		{"1234567890", ModeNumeric},
		{"HELLO WORLD", ModeAlphanumeric},
		{"hello world", ModeByte},
	} {
		t.Run(tc.want.String(), func(t *testing.T) {
			q, err := New(Options{Content: tc.content, Width: 500})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			res, err := Scan(q.Image())
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			found := false
			for _, s := range res.Segments {
				if s.Mode == tc.want {
					found = true
				}
			}
			if !found {
				var got []string
				for _, s := range res.Segments {
					got = append(got, s.Mode.String())
				}
				t.Errorf("modes = %v, want one to be %v", got, tc.want)
			}
		})
	}
}

func TestScanRejectsAnImageWithNoCode(t *testing.T) {
	blank := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.Draw(blank, blank.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	if _, err := Scan(blank); !errors.Is(err, ErrNoQRCode) {
		t.Fatalf("error = %v, want ErrNoQRCode", err)
	}
	if _, err := Scan(nil); !errors.Is(err, ErrNoQRCode) {
		t.Fatalf("error = %v, want ErrNoQRCode for a nil image", err)
	}
}

func TestScanFileAndScanReader(t *testing.T) {
	q, err := New(Options{Content: testURL, Width: 600})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	path := filepath.Join(t.TempDir(), "qr.png")
	if err := q.WritePNGFile(path); err != nil {
		t.Fatal(err)
	}

	res, err := ScanFile(path)
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	if res.Content != testURL {
		t.Errorf("ScanFile content = %q, want %q", res.Content, testURL)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	res, err = ScanReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("ScanReader: %v", err)
	}
	if res.Content != testURL {
		t.Errorf("ScanReader content = %q, want %q", res.Content, testURL)
	}

	if _, err := ScanFile("no-such-file.png"); err == nil {
		t.Error("ScanFile accepted a path that does not exist")
	}
}

// A WebP written by this library must be readable by it too, which is the
// round trip that proves the two new features fit together.
func TestScanReadsBackAWebP(t *testing.T) {
	q, err := New(Options{Content: testURL, Width: 600})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var buf bytes.Buffer
	if err := q.WebP(&buf); err != nil {
		t.Fatalf("WebP: %v", err)
	}
	res, err := ScanReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ScanReader on WebP: %v", err)
	}
	if res.Content != testURL {
		t.Errorf("content = %q, want %q", res.Content, testURL)
	}
}

// Modest rotation and downscaling are tolerated. These bounds are measured, not
// promised beyond what was tested: 45 degrees fails, and so it is not claimed.
func TestScanToleratesModestDistortion(t *testing.T) {
	q, err := New(Options{Content: testURL, Width: 600})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	clean := q.Image()

	t.Run("rotated 5 degrees", func(t *testing.T) {
		assertScans(t, rotateImage(clean, 5), testURL)
	})
	t.Run("rotated 15 degrees", func(t *testing.T) {
		assertScans(t, rotateImage(clean, 15), testURL)
	})
	t.Run("downscaled 3x", func(t *testing.T) {
		assertScans(t, shrinkImage(clean, 3), testURL)
	})
}

func assertScans(t *testing.T, img image.Image, want string) {
	t.Helper()
	res, err := Scan(img)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if res.Content != want {
		t.Errorf("content = %q, want %q", res.Content, want)
	}
}

func TestSegmentModeString(t *testing.T) {
	for _, tc := range []struct {
		m    SegmentMode
		want string
	}{
		{ModeNumeric, "numeric"}, {ModeAlphanumeric, "alphanumeric"},
		{ModeByte, "byte"}, {ModeKanji, "kanji"}, {ModeECI, "eci"},
		{ModeUnknown, "unknown"}, {SegmentMode(99), "unknown"},
	} {
		if got := tc.m.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", int(tc.m), got, tc.want)
		}
	}
}

// rotateImage turns src about its centre onto white, standing in for a code
// photographed at an angle.
func rotateImage(src image.Image, deg float64) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	rad := deg * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	cx, cy := float64(w)/2, float64(h)/2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			sx, sy := int(cx+dx*cos+dy*sin), int(cy-dx*sin+dy*cos)
			if sx >= 0 && sy >= 0 && sx < w && sy < h {
				dst.Set(x, y, src.At(sx+b.Min.X, sy+b.Min.Y))
			}
		}
	}
	return dst
}

// shrinkImage samples src down, standing in for a code photographed from a
// distance.
func shrinkImage(src image.Image, factor int) image.Image {
	b := src.Bounds()
	w, h := b.Dx()/factor, b.Dy()/factor
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, src.At(b.Min.X+x*factor, b.Min.Y+y*factor))
		}
	}
	return dst
}
