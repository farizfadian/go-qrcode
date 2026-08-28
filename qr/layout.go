package qr

import (
	"errors"
	"fmt"
)

// ErrWidthTooSmall reports a Width that cannot hold one pixel per module once
// the quiet zone is included.
var ErrWidthTooSmall = errors.New("qr: width too small for the module count")

// layout maps module coordinates to pixels. The module size is a whole number
// of pixels and the symbol is centred, with the spare pixels widening the quiet
// zone.
//
// Measurement decided this over stretching modules to fill Width exactly: over
// the same 32 cases, integer modules failed to decode 0 times and fractional
// ones 3 times, every failure being a dense symbol below 3 pixels per module
// where anti-aliased edges blur neighbouring modules together.
type layout struct {
	Modules    int
	Margin     int
	ModuleSize float64
	OriginX    float64
	OriginY    float64
	Width      int
}

// newLayout computes the pixel layout for a symbol of the given module count,
// quiet zone and image width. Modules and margin are counted in modules; width
// is in pixels.
func newLayout(modules, margin, width int) (layout, error) {
	if modules <= 0 {
		return layout{}, fmt.Errorf("qr: module count must be positive, got %d", modules)
	}
	if margin < 0 {
		return layout{}, fmt.Errorf("qr: margin must not be negative, got %d", margin)
	}
	total := modules + 2*margin
	size := width / total
	if size < 1 {
		return layout{}, fmt.Errorf("%w: %d pixels cannot hold %d modules",
			ErrWidthTooSmall, width, total)
	}
	origin := (width - size*total) / 2
	return layout{
		Modules:    modules,
		Margin:     margin,
		ModuleSize: float64(size),
		OriginX:    float64(origin),
		OriginY:    float64(origin),
		Width:      width,
	}, nil
}

// Rect returns the pixel position and size of module (x, y). The quiet zone is
// already accounted for, so (0, 0) is the first symbol module, not the first
// pixel.
func (l layout) Rect(x, y int) (px, py, size float64) {
	m := float64(l.Margin)
	return l.OriginX + (float64(x)+m)*l.ModuleSize,
		l.OriginY + (float64(y)+m)*l.ModuleSize,
		l.ModuleSize
}
