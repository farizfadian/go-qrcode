package qr

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	// Registered so a logo can be loaded from a file or reader without the
	// caller decoding it first.
	_ "image/jpeg"
	_ "image/png"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// Errors a logo can produce. Compare them with errors.Is.
var (
	ErrLogoSource   = errors.New("qr: logo needs exactly one of Image, Path or Reader")
	ErrLogoTooLarge = errors.New("qr: logo occludes more than the error-correction budget allows")
)

// Logo defaults, matching the reference library.
const (
	defaultLogoBorderWidth  = 10.0
	defaultLogoBorderRadius = 8.0
	defaultLogoBgColor      = "#ffffff"
)

// logoPlan is a positioned logo: where its frame and image sit in pixels, which
// modules it hides, and how many it is allowed to hide.
type logoPlan struct {
	img image.Image

	frameX, frameY, frameW, frameH float64 // the whole block, border included
	imgX, imgY, imgW, imgH         float64 // where the image is drawn

	borderRadius float64
	imageRadius  float64
	borderColor  color.RGBA
	bgColor      color.RGBA
	svgMarkup    string

	maxHidden int
}

// planLogo works out where the logo goes and how much occlusion the chosen
// error-correction level can afford. It does not yet know how many modules are
// actually hidden; New counts those and compares against maxHidden.
func planLogo(lo LogoOptions, background string, l layout, ecc ECCLevel, m *Matrix) (*logoPlan, error) {
	modules := m.Size()
	img, err := loadLogoImage(lo)
	if err != nil {
		return nil, err
	}
	// Validated here, not at render time, so a malformed vector logo fails
	// where every other configuration error does.
	if lo.SVGMarkup != "" {
		if err := render.ValidateSVG(lo.SVGMarkup); err != nil {
			return nil, fmt.Errorf("qr: logo SVGMarkup: %w", err)
		}
	}
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return nil, fmt.Errorf("qr: logo image is empty")
	}
	aspect := float64(b.Dx()) / float64(b.Dy())

	borderWidth := lo.BorderWidth
	if borderWidth == 0 {
		borderWidth = defaultLogoBorderWidth
	}
	borderRadius := lo.BorderRadius
	if borderRadius == 0 {
		borderRadius = defaultLogoBorderRadius
	}

	// The budget. The error-correction level names a fraction of *codewords*
	// it can restore, not a fraction of area that is safe to cover, so the
	// reference library squares it: High allows 9% of modules, not 30%. That is
	// deliberately conservative, and the exact count is validated afterwards.
	cover := ecc.recoveryFraction()
	maxHidden := int(cover * cover * float64(modules) * float64(modules))

	var frameW, frameH float64
	if lo.Size > 0 {
		// An explicit size is honoured exactly. If it exceeds the budget the
		// caller is told, rather than silently given something smaller than
		// they asked for.
		frameW = lo.Size * float64(l.Width)
		frameH = (frameW-2*borderWidth)/aspect + 2*borderWidth
	} else {
		frameW, frameH = autoFitLogo(l, m, aspect, maxHidden)
	}

	innerW := frameW - 2*borderWidth
	innerH := frameH - 2*borderWidth
	if innerW <= 0 || innerH <= 0 {
		return nil, fmt.Errorf(
			"qr: logo border of %gpx leaves no room inside a %gx%g block; reduce BorderWidth or raise Size",
			borderWidth, frameW, frameH)
	}

	// Fit the image inside the border, preserving its aspect ratio.
	imgW, imgH := innerW, innerH
	if innerW/innerH > aspect {
		imgW = innerH * aspect
	} else {
		imgH = innerW / aspect
	}

	centre := float64(l.Width) / 2
	borderColor, err := resolveLogoColor(lo.BorderColor, background)
	if err != nil {
		return nil, fmt.Errorf("%w: logo border: %v", ErrBadColor, err)
	}
	bgColor, err := resolveLogoColor(lo.BgColor, defaultLogoBgColor)
	if err != nil {
		return nil, fmt.Errorf("%w: logo background: %v", ErrBadColor, err)
	}

	return &logoPlan{
		img:          img,
		frameX:       centre - frameW/2,
		frameY:       centre - frameH/2,
		frameW:       frameW,
		frameH:       frameH,
		imgX:         centre - imgW/2,
		imgY:         centre - imgH/2,
		imgW:         imgW,
		imgH:         imgH,
		borderRadius: borderRadius,
		imageRadius:  lo.Radius,
		borderColor:  borderColor,
		bgColor:      bgColor,
		svgMarkup:    lo.SVGMarkup,
		maxHidden:    maxHidden,
	}, nil
}

// hides reports whether module (x, y) lies under the logo's frame.
func (p *logoPlan) hides(l layout) func(x, y int) bool {
	return hidesRect(l, p.frameX, p.frameY, p.frameW, p.frameH)
}

// hidesRect reports which modules a pixel rectangle covers. A module counts as
// hidden as soon as its own rectangle *touches* the frame, not merely when its
// centre falls inside: a half-covered dark module would otherwise still be
// drawn and leave a sliver poking out from under the logo.
//
// That choice has a consequence for sizing. A frame is not aligned to the
// module grid, so it touches up to one extra module on each side — a frame
// eleven modules wide can hide thirteen. Any budget arithmetic must count the
// modules actually touched rather than divide an area, which is what
// autoFitLogo does.
func hidesRect(l layout, fx, fy, fw, fh float64) func(x, y int) bool {
	return func(x, y int) bool {
		px, py, s := l.Rect(x, y)
		return px+s > fx && px < fx+fw && py+s > fy && py < fy+fh
	}
}

// autoFitLogo returns the largest frame of the given aspect ratio that stays
// inside the occlusion budget.
//
// It starts from the area the budget implies and shrinks until the measured
// count fits, because the touched-module count exceeds area/moduleSize² by the
// partially covered edge modules. Measuring beats estimating here: the error
// is a whole ring of modules, not a rounding difference.
func autoFitLogo(l layout, m *Matrix, aspect float64, maxHidden int) (w, h float64) {
	area := float64(maxHidden) * l.ModuleSize * l.ModuleSize
	h = math.Sqrt(area / aspect)
	w = h * aspect

	centre := float64(l.Width) / 2
	for i := 0; i < 200 && w > l.ModuleSize && h > l.ModuleSize; i++ {
		hides := hidesRect(l, centre-w/2, centre-h/2, w, h)
		if countHidden(m, hides) <= maxHidden {
			break
		}
		w *= 0.97
		h *= 0.97
	}
	return w, h
}

// items returns the scene items that draw the logo, in paint order: the frame,
// the background the image sits on, then the image itself.
func (p *logoPlan) items() []render.Item {
	frame := render.RoundRect(p.frameX, p.frameY, p.frameW, p.frameH,
		p.borderRadius, p.borderRadius, p.borderRadius, p.borderRadius)

	inner := render.RoundRect(p.imgX, p.imgY, p.imgW, p.imgH,
		p.imageRadius, p.imageRadius, p.imageRadius, p.imageRadius)

	out := []render.Item{
		render.PathItem{Path: frame, Fill: p.borderColor},
		render.PathItem{Path: inner, Fill: p.bgColor},
	}

	item := render.ImageItem{
		Img:       p.img,
		SVGMarkup: p.svgMarkup,
		X:         p.imgX, Y: p.imgY, W: p.imgW, H: p.imgH,
	}
	if p.imageRadius > 0 {
		clip := inner
		item.Clip = &clip
	}
	return append(out, item)
}

// loadLogoImage resolves exactly one of the three source forms.
func loadLogoImage(lo LogoOptions) (image.Image, error) {
	sources := 0
	if lo.Image != nil {
		sources++
	}
	if lo.Path != "" {
		sources++
	}
	if lo.Reader != nil {
		sources++
	}
	if sources != 1 {
		return nil, fmt.Errorf("%w: %d were set", ErrLogoSource, sources)
	}

	switch {
	case lo.Image != nil:
		return lo.Image, nil
	case lo.Path != "":
		f, err := os.Open(lo.Path)
		if err != nil {
			return nil, fmt.Errorf("qr: opening logo: %w", err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("qr: decoding logo %s: %w", lo.Path, err)
		}
		return img, nil
	default:
		img, _, err := image.Decode(lo.Reader)
		if err != nil {
			return nil, fmt.Errorf("qr: decoding logo: %w", err)
		}
		return img, nil
	}
}

func resolveLogoColor(value, fallback string) (color.RGBA, error) {
	if value == "" {
		value = fallback
	}
	return render.ParseColor(value)
}

// logoTooLargeAdvice says what the caller can actually do about an oversized
// logo, given what they already asked for.
//
// Suggesting a higher error-correction level when the code is already at High
// sends someone chasing a setting that does not exist, and they are as likely to
// conclude the library is broken as that the advice was wrong.
func logoTooLargeAdvice(o Options, ecc ECCLevel) string {
	switch {
	case o.Logo != nil && o.Logo.Size > 0 && ecc < ECCHigh:
		return "reduce Logo.Size, leave it unset to fit automatically, " +
			"or raise ECC to H for a larger budget"
	case o.Logo != nil && o.Logo.Size > 0:
		return "reduce Logo.Size, or leave it unset to fit the largest logo " +
			"this symbol can carry; ECC is already at H, the highest level"
	case ecc < ECCHigh:
		return "raise ECC to H for a larger budget, or set Logo.BorderWidth lower"
	default:
		return "the symbol is too small to carry a logo at all; " +
			"ECC is already at H, so encode more content or drop the logo"
	}
}

// countHidden returns how many of the symbol's modules the predicate hides.
// Both dark and light modules count: a light module carries data too, so
// covering one costs the same error-correction budget.
func countHidden(m *Matrix, hides func(x, y int) bool) int {
	if hides == nil {
		return 0
	}
	n := 0
	for y := 0; y < m.Size(); y++ {
		for x := 0; x < m.Size(); x++ {
			if hides(x, y) {
				n++
			}
		}
	}
	return n
}
