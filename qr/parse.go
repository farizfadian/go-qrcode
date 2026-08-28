package qr

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrUnknownShape reports a dot or corner shape this build does not implement.
// The error message lists the names that are available.
var ErrUnknownShape = errors.New("qr: shape not implemented in this build")

// DotTypes returns every dot shape this build implements, in declaration order.
//
// The list is derived from the shape registry rather than written out
// separately, so it grows on its own as shapes are added. Callers driven by
// configuration - a CLI, a JSON config file - never need updating alongside the
// renderer.
func DotTypes() []DotType {
	out := make([]DotType, 0, len(dotFuncs))
	for i := range dotNames {
		d := DotType(i)
		if _, ok := dotFuncs[d]; ok {
			out = append(out, d)
		}
	}
	return out
}

// CornerTypes returns every finder-pattern shape this build implements, in
// declaration order.
func CornerTypes() []CornerType {
	out := make([]CornerType, 0, len(cornerFuncs))
	for i := range cornerNames {
		c := CornerType(i)
		if _, ok := cornerFuncs[c]; ok {
			out = append(out, c)
		}
	}
	return out
}

// ParseDotType returns the dot shape named by s, matching DotType.String
// case-insensitively. A name that exists in the specification but is not yet
// implemented is rejected rather than quietly falling back to square.
func ParseDotType(s string) (DotType, error) {
	want := strings.ToLower(strings.TrimSpace(s))
	for _, d := range DotTypes() {
		if d.String() == want {
			return d, nil
		}
	}
	return DotSquare, fmt.Errorf("%w: dot shape %q; available: %s",
		ErrUnknownShape, s, strings.Join(DotTypeNames(), ", "))
}

// ParseCornerType returns the finder-pattern shape named by s, matching
// CornerType.String case-insensitively.
func ParseCornerType(s string) (CornerType, error) {
	want := strings.ToLower(strings.TrimSpace(s))
	for _, c := range CornerTypes() {
		if c.String() == want {
			return c, nil
		}
	}
	return CornerSquare, fmt.Errorf("%w: corner shape %q; available: %s",
		ErrUnknownShape, s, strings.Join(CornerTypeNames(), ", "))
}

// DotTypeNames returns the names of every implemented dot shape, for help text.
func DotTypeNames() []string {
	out := make([]string, 0, len(dotFuncs))
	for _, d := range DotTypes() {
		out = append(out, d.String())
	}
	sort.Strings(out)
	return out
}

// CornerTypeNames returns the names of every implemented finder-pattern shape,
// for help text.
func CornerTypeNames() []string {
	out := make([]string, 0, len(cornerFuncs))
	for _, c := range CornerTypes() {
		out = append(out, c.String())
	}
	sort.Strings(out)
	return out
}

// ParseECCLevel returns the error-correction level named by s. It accepts the
// single letters L, M, Q and H in either case, and "auto" or the empty string
// for automatic selection.
func ParseECCLevel(s string) (ECCLevel, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "", "AUTO":
		return ECCAuto, nil
	case "L":
		return ECCLow, nil
	case "M":
		return ECCMedium, nil
	case "Q":
		return ECCQuartile, nil
	case "H":
		return ECCHigh, nil
	}
	return ECCAuto, fmt.Errorf("qr: unknown ECC level %q; want auto, L, M, Q or H", s)
}
