package qr

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// Errors reported when a colour scheme cannot be scanned. Compare them with
// errors.Is.
var (
	// ErrLowContrast reports foreground and background colours too close in
	// luminance for a reader to tell apart.
	ErrLowContrast = errors.New("qr: foreground/background contrast too low to scan")

	// ErrInvertedPolarity reports a foreground lighter than its background.
	// This is a distinct failure from low contrast: an inverted code can have
	// perfect contrast and still be unreadable.
	ErrInvertedPolarity = errors.New("qr: foreground must be darker than background")
)

// MinContrastRatio is the lowest WCAG contrast ratio this library will render.
//
// Derived by measurement, not judgement. Sweeping a grey foreground against
// white, the lowest ratio that still decoded was 3.54 and the highest that
// failed was 2.96; coloured pairs behaved the same as greys at matching ratios,
// confirming luminance is the right metric. The floor is set at the lowest
// value observed to work.
//
// That measurement is a best case: a clean 640-pixel render read by software.
// A camera, at an angle, in poor light, needs far more headroom, so aim for 4.5
// or better in print. This constant is where the library refuses, not where
// good design begins.
const MinContrastRatio = 3.5

// checkContrast rejects colour schemes that cannot be read.
//
// A transparent background is judged against white, since that is what a code
// with no background of its own will usually end up sitting on.
func checkContrast(background color.RGBA, layers map[string]color.RGBA) error {
	bg := render.Over(background, color.RGBA{0xff, 0xff, 0xff, 0xff})
	bgLum := render.Luminance(bg)

	// Deterministic order, so the same options always report the same problem.
	for _, name := range []string{"foreground", "dots", "corners"} {
		fg, ok := layers[name]
		if !ok {
			continue
		}
		solid := render.Over(fg, bg)

		if render.Luminance(solid) > bgLum {
			return fmt.Errorf(
				"%w: the %s colour is lighter than the background; "+
					"readers look for dark modules on a light field, so an inverted "+
					"code will not scan however good its contrast is",
				ErrInvertedPolarity, name)
		}

		if r := render.ContrastRatio(solid, bg); r < MinContrastRatio {
			return fmt.Errorf(
				"%w: the %s colour has a contrast ratio of %.2f against the "+
					"background, below the measured floor of %.2f",
				ErrLowContrast, name, r, MinContrastRatio)
		}
	}
	return nil
}
