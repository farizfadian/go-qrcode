package render

import (
	"image"
	"image/color"
	"testing"
)

func TestRasterFillsAPath(t *testing.T) {
	sc := Scene{
		Width: 20, Height: 20,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items: []Item{PathItem{
			Path: RoundRect(5, 5, 10, 10, 0, 0, 0, 0),
			Fill: color.RGBA{0, 0, 0, 0xff},
		}},
	}
	img := Raster(sc)
	if got := img.RGBAAt(10, 10); got.R != 0 || got.A != 0xff {
		t.Errorf("centre = %v, want opaque black", got)
	}
	if got := img.RGBAAt(1, 1); got.R != 0xff {
		t.Errorf("corner = %v, want white background", got)
	}
}

// This is the architectural test: a ring must be a hole, not a filled square.
// If it fails, the fill-only path model does not work and nothing built on it
// will either.
func TestRasterPunchesHoleWithReversedWinding(t *testing.T) {
	outer := RoundRect(0, 0, 30, 30, 0, 0, 0, 0)
	inner := RoundRect(10, 10, 10, 10, 0, 0, 0, 0).Reverse()
	sc := Scene{
		Width: 30, Height: 30,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items:      []Item{PathItem{Path: outer.Append(inner), Fill: color.RGBA{0, 0, 0, 0xff}}},
	}
	img := Raster(sc)
	if got := img.RGBAAt(15, 15); got.R != 0xff {
		t.Errorf("ring centre = %v, want the white background showing through", got)
	}
	if got := img.RGBAAt(4, 15); got.R != 0 {
		t.Errorf("ring body = %v, want opaque black", got)
	}
}

func TestRasterKeepsTransparentBackground(t *testing.T) {
	sc := Scene{Width: 8, Height: 8, Background: color.RGBA{}}
	if got := Raster(sc).RGBAAt(4, 4); got.A != 0 {
		t.Errorf("alpha = %d, want 0", got.A)
	}
}

func TestRasterDrawsClippedImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			src.Set(x, y, color.RGBA{0, 0xff, 0, 0xff})
		}
	}
	clip := Circle(10, 10, 5)
	sc := Scene{
		Width: 20, Height: 20,
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
		Items:      []Item{ImageItem{Img: src, X: 5, Y: 5, W: 10, H: 10, Clip: &clip}},
	}
	img := Raster(sc)
	if got := img.RGBAAt(10, 10); got.G != 0xff {
		t.Errorf("clip centre = %v, want green", got)
	}
	// Pick a pixel wholly outside the circle, not merely one whose centre is.
	// Pixel (6,6) spans [6,7]x[6,7] and its far corner (7,7) is 4.24 from the
	// centre, inside the radius of 5, so it is legitimately part-covered and
	// anti-aliases to a pale green. Pixel (5,5)'s nearest corner is 5.66 away,
	// so it is fully outside.
	if got := img.RGBAAt(5, 5); got.R != 0xff || got.G != 0xff || got.B != 0xff {
		t.Errorf("just outside the clip = %v, want white background", got)
	}
	if got := img.RGBAAt(18, 18); got.R != 0xff || got.G != 0xff || got.B != 0xff {
		t.Errorf("beyond the image rect = %v, want white background", got)
	}
}
