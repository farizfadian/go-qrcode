package qr

import (
	"errors"
	"fmt"

	goqr "github.com/piglig/go-qr"
)

// ErrContentTooLong reports content that does not fit in a version 40 symbol at
// the requested error-correction level.
var ErrContentTooLong = errors.New("qr: content does not fit at the requested ECC level")

// Encoder turns text into a QR module matrix. It exists so the third-party
// encoder can be replaced by a built-in one without a breaking change.
type Encoder interface {
	// Encode returns a square matrix indexed as [y][x]; true means a dark
	// module. ecc must be a concrete level, never ECCAuto.
	Encode(content string, ecc ECCLevel) ([][]bool, error)
}

// defaultEncoder returns the encoder New uses when Options names none.
func defaultEncoder() Encoder { return pigligEncoder{} }

// pigligEncoder adapts github.com/piglig/go-qr, a port of Nayuki's reference
// implementation.
type pigligEncoder struct{}

func (pigligEncoder) Encode(content string, ecc ECCLevel) ([][]bool, error) {
	level, ok := pigligLevel(ecc)
	if !ok {
		return nil, fmt.Errorf("qr: ECC level %v must be resolved before encoding", ecc)
	}
	segs, err := goqr.MakeSegments(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContentTooLong, err)
	}
	// boostEcl is pinned false on purpose. EncodeText would silently raise the
	// level whenever the data still fits, making the effective level differ
	// from the requested one and breaking both golden determinism and the logo
	// occlusion budget.
	code, err := goqr.EncodeSegments(segs, level, 1, 40, -1, false)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContentTooLong, err)
	}
	return toMatrix(code), nil
}

func pigligLevel(e ECCLevel) (goqr.Ecc, bool) {
	switch e {
	case ECCLow:
		return goqr.Low, true
	case ECCMedium:
		return goqr.Medium, true
	case ECCQuartile:
		return goqr.Quartile, true
	case ECCHigh:
		return goqr.High, true
	}
	return goqr.Low, false
}

func toMatrix(code *goqr.QrCode) [][]bool {
	n := code.Size()
	out := make([][]bool, n)
	for y := 0; y < n; y++ {
		row := make([]bool, n)
		for x := 0; x < n; x++ {
			row[x] = code.Module(x, y)
		}
		out[y] = row
	}
	return out
}

// boostedMatrixForTest exposes the boosted encoding path so a test can prove
// our own path differs from it. It is unexported and used only by tests.
func boostedMatrixForTest(content string, ecc ECCLevel) ([][]bool, error) {
	level, ok := pigligLevel(ecc)
	if !ok {
		return nil, fmt.Errorf("qr: bad level %v", ecc)
	}
	code, err := goqr.EncodeText(content, level)
	if err != nil {
		return nil, err
	}
	return toMatrix(code), nil
}
