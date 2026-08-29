package qr

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// testLogo builds a simple high-contrast mark, so the logo tests need no binary
// fixture in the repository and can vary the aspect ratio freely.
func testLogo(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.RGBA{0x22, 0x55, 0xdd, 0xff}
			if (x*4/w+y*4/h)%2 == 0 {
				c = color.RGBA{0xff, 0xcc, 0x00, 0xff}
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func writeTestLogo(t *testing.T, w, h int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "logo.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, testLogo(w, h)); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLogoNeedsExactlyOneSource(t *testing.T) {
	for _, tc := range []struct {
		name string
		logo LogoOptions
	}{
		{"none", LogoOptions{}},
		{"image and path", LogoOptions{Image: testLogo(64, 64), Path: "x.png"}},
		{"path and reader", LogoOptions{Path: "x.png", Reader: bytes.NewReader(nil)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := tc.logo
			_, err := New(Options{Content: testURL, Logo: &l})
			if !errors.Is(err, ErrLogoSource) {
				t.Fatalf("error = %v, want ErrLogoSource", err)
			}
		})
	}
}

func TestLogoAcceptsEverySourceForm(t *testing.T) {
	path := writeTestLogo(t, 128, 128)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		logo LogoOptions
	}{
		{"Image", LogoOptions{Image: testLogo(128, 128)}},
		{"Path", LogoOptions{Path: path}},
		{"Reader", LogoOptions{Reader: bytes.NewReader(raw)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := tc.logo
			q, err := New(Options{Content: testURL, Width: 640, Logo: &l})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if q.Image().Bounds().Dx() != 640 {
				t.Errorf("width = %d, want 640", q.Image().Bounds().Dx())
			}
		})
	}
}

func TestLogoReportsAMissingFile(t *testing.T) {
	_, err := New(Options{Content: testURL, Logo: &LogoOptions{Path: "no-such-file.png"}})
	if err == nil {
		t.Fatal("New accepted a logo path that does not exist")
	}
}

// A logo occludes data, so New forces the highest error-correction level when
// the caller did not pick one. The budget depends on knowing the real level.
func TestLogoForcesHighECCWhenLevelIsAuto(t *testing.T) {
	q, err := New(Options{
		Content: testURL,
		Width:   640,
		Logo:    &LogoOptions{Image: testLogo(64, 64)},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if q.ECC() != ECCHigh {
		t.Errorf("ECC = %v, want H", q.ECC())
	}
}

func TestLogoRespectsAnExplicitECCLevel(t *testing.T) {
	q, err := New(Options{
		Content: testURL,
		Width:   640,
		ECC:     ECCQuartile,
		Logo:    &LogoOptions{Image: testLogo(64, 64), Size: 0.15},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if q.ECC() != ECCQuartile {
		t.Errorf("ECC = %v, want Q; an explicit level must win", q.ECC())
	}
}

// The modules under the logo must not be drawn. Rendering them and painting
// over the top would waste work and, worse, leave dark fringes around a logo
// with rounded corners.
func TestLogoHidesTheModulesBeneathIt(t *testing.T) {
	q, err := New(Options{
		Content: testURL,
		Width:   640,
		Logo:    &LogoOptions{Image: testLogo(64, 64), Size: 0.2},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if q.hiddenModules == 0 {
		t.Fatal("no modules were hidden; the safe zone did nothing")
	}
	if q.hiddenModules > q.logoBudget {
		t.Errorf("hid %d modules, budget is %d", q.hiddenModules, q.logoBudget)
	}
}

// Emitting an unscannable code is worse than refusing. A logo that covers more
// than the error correction can recover is rejected.
func TestOversizedLogoIsRejected(t *testing.T) {
	_, err := New(Options{
		Content: testURL,
		Width:   640,
		Logo:    &LogoOptions{Image: testLogo(64, 64), Size: 0.8},
	})
	if !errors.Is(err, ErrLogoTooLarge) {
		t.Fatalf("error = %v, want ErrLogoTooLarge", err)
	}
}

func TestLogoBorderCannotSwallowTheImage(t *testing.T) {
	_, err := New(Options{
		Content: testURL,
		Width:   640,
		Logo: &LogoOptions{
			Image:       testLogo(64, 64),
			Size:        0.05,
			BorderWidth: 40,
		},
	})
	if err == nil {
		t.Fatal("New accepted a border wider than the logo block")
	}
}

// The scene must carry the logo as an image with a clip path, and the frame
// beneath it, in that order.
func TestLogoAddsFrameAndImageToTheScene(t *testing.T) {
	q, err := New(Options{
		Content: testURL,
		Width:   640,
		Logo:    &LogoOptions{Image: testLogo(64, 64), Size: 0.2, Radius: 8},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	items := q.sc.Items
	if len(items) < 3 {
		t.Fatalf("scene has %d items, want dots, corners, frame and image", len(items))
	}
	if _, isImage := items[len(items)-1].(render.ImageItem); !isImage {
		t.Errorf("last item is %T, want the logo image drawn on top", items[len(items)-1])
	}
}

// The whole point of the logo work: a code with a logo must still scan.
func TestLogoDecodesAcrossDotShapes(t *testing.T) {
	for _, typ := range []DotType{DotSquare, DotDot, DotFluid, DotStripe, DotStar} {
		t.Run(typ.String(), func(t *testing.T) {
			requireDecodableBaseline(t, testURL, ECCHigh)
			q, err := New(Options{
				Content: testURL,
				Width:   900,
				Dots:    DotOptions{Type: typ},
				Logo:    &LogoOptions{Image: testLogo(128, 128)},
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			assertDecodes(t, q.Image(), testURL)
		})
	}
}

func TestLogoDecodesAcrossCornerShapes(t *testing.T) {
	for _, typ := range CornerTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			requireDecodableBaseline(t, testURL, ECCHigh)
			q, err := New(Options{
				Content: testURL,
				Width:   900,
				Corners: CornerOptions{Type: typ},
				Logo:    &LogoOptions{Image: testLogo(128, 128)},
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			assertDecodes(t, q.Image(), testURL)
		})
	}
}

func TestLogoDecodesAtSeveralSizes(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCHigh)
	for _, size := range []float64{0, 0.12, 0.18, 0.22} {
		t.Run(sizeName(size), func(t *testing.T) {
			q, err := New(Options{
				Content: testURL,
				Width:   900,
				Logo:    &LogoOptions{Image: testLogo(128, 128), Size: size},
			})
			if err != nil {
				t.Fatalf("New(size=%v): %v", size, err)
			}
			assertDecodes(t, q.Image(), testURL)
		})
	}
}

func TestNonSquareLogoKeepsItsAspectRatio(t *testing.T) {
	q, err := New(Options{
		Content: testURL,
		Width:   900,
		Logo:    &LogoOptions{Image: testLogo(200, 100), Size: 0.2},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	requireDecodableBaseline(t, testURL, ECCHigh)
	assertDecodes(t, q.Image(), testURL)
}

func TestSVGWithLogoEmbedsTheImage(t *testing.T) {
	q, err := New(Options{
		Content: testURL,
		Width:   640,
		Logo:    &LogoOptions{Image: testLogo(64, 64), Size: 0.18, Radius: 6},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	markup, err := q.SVGString()
	if err != nil {
		t.Fatalf("SVGString: %v", err)
	}
	for _, want := range []string{"<image ", "data:image/png;base64,", "<clipPath ", `clip-path="url(#`} {
		if !strings.Contains(markup, want) {
			t.Errorf("SVG is missing %q", want)
		}
	}
}

func sizeName(f float64) string {
	if f == 0 {
		return "auto"
	}
	return "size-" + strconv.FormatFloat(f, 'f', -1, 64)
}
