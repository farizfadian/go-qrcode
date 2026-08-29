package qr

import (
	"errors"
	"image"
	"io"
)

// Default values applied when the corresponding Options field is left at its
// zero value.
const (
	DefaultWidth      = 380
	DefaultMargin     = 4
	DefaultForeground = "#000000"
	DefaultBackground = "#ffffff"
)

// Errors returned by New. Compare them with errors.Is.
var (
	ErrNoContent = errors.New("qr: content is empty")
	ErrBadColor  = errors.New("qr: cannot parse colour")
)

// Options configures a QR code. Only Content is required: every other field has
// a working default, so the zero value produces a conventional black-on-white
// code at 380 pixels.
type Options struct {
	// Content is the text to encode. Required.
	Content string
	// Width is the output image size in pixels. Zero selects DefaultWidth.
	Width int
	// Margin is the quiet zone in modules. Zero selects DefaultMargin, which is
	// the four modules ISO/IEC 18004 requires.
	Margin int
	// ECC is the error-correction level. The zero value lets New choose.
	ECC ECCLevel

	// Foreground is the module colour. Empty selects DefaultForeground.
	Foreground string
	// Background is the backdrop colour. Empty selects DefaultBackground;
	// "#00000000" makes it transparent.
	Background string

	// Dots styles the data modules.
	Dots DotOptions
	// Corners styles the three finder patterns.
	Corners CornerOptions
	// Logo places an image at the centre. Nil means no logo.
	Logo *LogoOptions
}

// DotOptions styles the data modules.
type DotOptions struct {
	// Type selects the figure. The zero value is DotSquare.
	Type DotType
	// Color overrides Options.Foreground for dots. Empty inherits it.
	Color string
}

// CornerOptions styles the three finder patterns.
type CornerOptions struct {
	// Type selects the figure. The zero value is CornerSquare.
	Type CornerType
	// Color overrides Options.Foreground for corners. Empty inherits it.
	Color string
	// Radius sets the corner radii for the rounding shapes. Zero fields fall
	// back to a fraction of the module size.
	Radius CornerRadius
}

// LogoOptions describes a centred logo. Exactly one of Image, Path or Reader
// must be set.
type LogoOptions struct {
	Image  image.Image
	Path   string
	Reader io.Reader

	// SVGMarkup is an optional vector version of the same logo, used only for
	// SVG output, where it is embedded as a nested <svg> and rendered by
	// whatever displays the page. Gradients, curves, strokes and text survive
	// exactly as authored, at any zoom.
	//
	// It does not replace the raster logo above; it sits alongside it. Go has no
	// SVG rasteriser, so PNG, JPEG and WebP still need Image, Path or Reader.
	// Requiring both is deliberate: the alternative was an option that made
	// New succeed and PNG fail later, and Image cannot report an error at all,
	// so it would have quietly produced a code with no logo on it.
	//
	// Export a PNG from the same source once, and set both.
	SVGMarkup string

	// Size is the logo block's width as a fraction of Options.Width, border
	// included. Zero selects automatic sizing from the ECC budget.
	Size float64
	// Radius rounds the logo image's own corners, in pixels.
	Radius float64
	// BorderWidth is the frame around the image, in pixels. Zero selects 10.
	BorderWidth float64
	// BorderRadius rounds the frame, in pixels. Zero selects 8.
	BorderRadius float64
	// BorderColor fills the frame. Empty inherits Options.Background.
	BorderColor string
	// BgColor fills behind the image. Empty selects "#ffffff".
	BgColor string
}

// withDefaults returns a copy of o with every zero field replaced by its
// default. It never mutates the caller's value.
func (o Options) withDefaults() Options {
	if o.Width == 0 {
		o.Width = DefaultWidth
	}
	if o.Margin == 0 {
		o.Margin = DefaultMargin
	}
	if o.Foreground == "" {
		o.Foreground = DefaultForeground
	}
	if o.Background == "" {
		o.Background = DefaultBackground
	}
	if o.Dots.Color == "" {
		o.Dots.Color = o.Foreground
	}
	if o.Corners.Color == "" {
		o.Corners.Color = o.Foreground
	}
	return o
}

// resolveECC turns ECCAuto into a concrete level. With a logo the occlusion
// budget matters more than symbol size, so High wins. Without one the level
// scales with content length, matching the reference library: short content has
// room for maximum protection, long content does not, and forcing High on long
// content only shrinks the modules and makes scanning harder.
func (o Options) resolveECC() ECCLevel {
	if o.ECC != ECCAuto {
		return o.ECC
	}
	if o.Logo != nil {
		return ECCHigh
	}
	switch n := len(o.Content); {
	case n > 36:
		return ECCMedium
	case n > 16:
		return ECCQuartile
	default:
		return ECCHigh
	}
}
