package qr

import (
	"errors"
	"testing"
)

// The thresholds here come from a sweep run against a real decoder: a grey
// foreground on white decoded down to a contrast ratio of 3.54 and failed from
// 2.96 down. These cases sit either side of that boundary.
func TestContrastAcceptsWhatDecodesAndRejectsWhatDoesNot(t *testing.T) {
	for _, tc := range []struct {
		fg, bg string
		accept bool
		why    string
	}{
		{"#000000", "#ffffff", true, "ratio 21, the maximum"},
		{"#1f2937", "#ffffff", true, "ratio 14.7, a common dark slate"},
		{"#dc2626", "#ffffff", true, "ratio 4.8, a saturated red"},
		{"#0f766e", "#ffffff", true, "ratio 5.5, a teal"},
		{"#666666", "#ffffff", true, "ratio 5.7"},
		{"#878787", "#ffffff", true, "ratio 3.59, the lowest measured to decode"},
		{"#969696", "#ffffff", false, "ratio 2.96, measured to fail"},
		{"#aaaaaa", "#ffffff", false, "ratio 2.3"},
		{"#ffffff", "#ffffff", false, "ratio 1, invisible"},
	} {
		t.Run(tc.fg, func(t *testing.T) {
			_, err := New(Options{Content: testURL, Foreground: tc.fg, Background: tc.bg})
			switch {
			case tc.accept && err != nil:
				t.Errorf("rejected %s (%s): %v", tc.fg, tc.why, err)
			case !tc.accept && !errors.Is(err, ErrLowContrast):
				t.Errorf("accepted %s (%s); error = %v, want ErrLowContrast", tc.fg, tc.why, err)
			}
		})
	}
}

// Inverted codes have excellent contrast and still do not scan, so they need a
// check of their own. All four of these were measured failing to decode.
func TestInvertedPolarityIsRejectedRegardlessOfContrast(t *testing.T) {
	for _, tc := range []struct{ fg, bg string }{
		{"#ffffff", "#000000"},
		{"#ffffff", "#1f2937"},
		{"#f8fafc", "#0f172a"},
		{"#ffff00", "#000080"},
	} {
		t.Run(tc.fg+"-on-"+tc.bg, func(t *testing.T) {
			_, err := New(Options{Content: testURL, Foreground: tc.fg, Background: tc.bg})
			if !errors.Is(err, ErrInvertedPolarity) {
				t.Errorf("error = %v, want ErrInvertedPolarity", err)
			}
		})
	}
}

// Dots and corners can be coloured separately, so each must be checked. A
// scheme whose dots are fine but whose finder patterns vanish is still broken.
func TestContrastChecksDotsAndCornersSeparately(t *testing.T) {
	_, err := New(Options{
		Content: testURL,
		Corners: CornerOptions{Color: "#eeeeee"},
	})
	if !errors.Is(err, ErrLowContrast) {
		t.Fatalf("error = %v, want ErrLowContrast for washed-out corners", err)
	}

	_, err = New(Options{
		Content: testURL,
		Dots:    DotOptions{Color: "#eeeeee"},
	})
	if !errors.Is(err, ErrLowContrast) {
		t.Fatalf("error = %v, want ErrLowContrast for washed-out dots", err)
	}
}

// A transparent background is judged against white, because that is what a code
// with no background of its own will usually sit on.
func TestTransparentBackgroundIsJudgedAgainstWhite(t *testing.T) {
	if _, err := New(Options{
		Content: testURL, Foreground: "#000000", Background: "#00000000",
	}); err != nil {
		t.Errorf("black on transparent was rejected: %v", err)
	}

	_, err := New(Options{
		Content: testURL, Foreground: "#eeeeee", Background: "#00000000",
	})
	if !errors.Is(err, ErrLowContrast) {
		t.Errorf("error = %v; pale on transparent should fail against white", err)
	}
}

// The check must not fire on the default scheme, which is the maximum contrast
// available.
func TestDefaultColoursPassTheContrastCheck(t *testing.T) {
	if _, err := New(Options{Content: testURL}); err != nil {
		t.Fatalf("the default black-on-white scheme was rejected: %v", err)
	}
}
