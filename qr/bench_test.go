package qr

import (
	"image"
	"image/color"
	"io"
	"strings"
	"testing"
)

const benchContent = "https://github.com/farizfadian/go-qrcode"

func benchLogo() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 0x80, 0xff})
		}
	}
	return img
}

// New covers encoding, classification, layout and path building — everything
// except turning paths into pixels.
func BenchmarkNewDefault(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := New(Options{Content: benchContent, Width: 512}); err != nil {
			b.Fatal(err)
		}
	}
}

// fluid is the most expensive shape per module: it queries four neighbours and
// builds a per-corner rounded rectangle rather than a plain one.
func BenchmarkNewFluid(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := New(Options{
			Content: benchContent, Width: 512,
			Dots: DotOptions{Type: DotFluid},
		}); err != nil {
			b.Fatal(err)
		}
	}
}

// fluid-line adds the diagonal fillets, so it is fluid plus extra subpaths.
func BenchmarkNewFluidLine(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := New(Options{
			Content: benchContent, Width: 512,
			Dots: DotOptions{Type: DotFluidLine},
		}); err != nil {
			b.Fatal(err)
		}
	}
}

// stripe does the most work per module that is not geometry: it probes runs and
// writes consumption state back.
func BenchmarkNewStripe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := New(Options{
			Content: benchContent, Width: 512,
			Dots: DotOptions{Type: DotStripe},
		}); err != nil {
			b.Fatal(err)
		}
	}
}

// The logo path adds image loading, the auto-fit search and the occlusion count.
func BenchmarkNewWithLogo(b *testing.B) {
	logo := benchLogo()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := New(Options{
			Content: benchContent, Width: 900,
			Logo: &LogoOptions{Image: logo},
		}); err != nil {
			b.Fatal(err)
		}
	}
}

// Rasterisation alone, with New's work hoisted out. This is what a server
// repeats per request when it caches the *QR.
func BenchmarkImage(b *testing.B) {
	q, err := New(Options{Content: benchContent, Width: 512})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = q.Image()
	}
}

func BenchmarkImageFluid(b *testing.B) {
	q, err := New(Options{Content: benchContent, Width: 512, Dots: DotOptions{Type: DotFluid}})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = q.Image()
	}
}

// Serialising to SVG should be far cheaper than rasterising, since it only
// formats numbers.
func BenchmarkSVGString(b *testing.B) {
	q, err := New(Options{Content: benchContent, Width: 512})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := q.SVGString(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPNG(b *testing.B) {
	q, err := New(Options{Content: benchContent, Width: 512})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := q.PNG(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

// A long payload forces a large symbol, so this measures how the cost scales
// with module count rather than with image size.
func BenchmarkNewLongContent(b *testing.B) {
	long := "https://example.com/" + strings.Repeat("abcdefghij", 90)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := New(Options{Content: long, Width: 1000, ECC: ECCMedium}); err != nil {
			b.Fatal(err)
		}
	}
}
