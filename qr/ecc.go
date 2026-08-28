package qr

// ECCLevel is a QR error-correction level. The zero value, ECCAuto, lets New
// choose: with a logo it selects ECCHigh so the occluded area fits the recovery
// budget, and without one it scales with content length.
type ECCLevel int

// The error-correction levels defined by ISO/IEC 18004, plus the automatic
// selection that is the zero value.
const (
	ECCAuto ECCLevel = iota
	ECCLow
	ECCMedium
	ECCQuartile
	ECCHigh
)

// String returns the single-letter name used by the QR specification, or
// "auto" for ECCAuto.
func (e ECCLevel) String() string {
	switch e {
	case ECCLow:
		return "L"
	case ECCMedium:
		return "M"
	case ECCQuartile:
		return "Q"
	case ECCHigh:
		return "H"
	default:
		return "auto"
	}
}

// recoveryFraction returns the share of codewords the level can restore. It
// drives the logo occlusion budget.
func (e ECCLevel) recoveryFraction() float64 {
	switch e {
	case ECCLow:
		return 0.07
	case ECCMedium:
		return 0.15
	case ECCQuartile:
		return 0.25
	case ECCHigh:
		return 0.30
	default:
		return 0
	}
}
