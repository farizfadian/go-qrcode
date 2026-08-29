package qr

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	// Registered so a code can be scanned from a file or reader without the
	// caller decoding the image first.
	_ "image/jpeg"
	_ "image/png"

	goqr "github.com/piglig/go-qr"
	_ "golang.org/x/image/webp" // decode-only; pairs with QR.WebP output
)

// ErrNoQRCode reports an image with no readable QR code in it.
var ErrNoQRCode = errors.New("qr: no readable QR code in the image")

// SegmentMode is how one part of a scanned code's content was encoded. QR packs
// digits and uppercase text more densely than arbitrary bytes, and a single
// symbol may switch modes partway through.
type SegmentMode int

// The encoding modes defined by ISO/IEC 18004.
const (
	ModeUnknown SegmentMode = iota
	ModeNumeric
	ModeAlphanumeric
	ModeByte
	ModeKanji
	ModeECI
)

// String returns the mode's specification name.
func (m SegmentMode) String() string {
	switch m {
	case ModeNumeric:
		return "numeric"
	case ModeAlphanumeric:
		return "alphanumeric"
	case ModeByte:
		return "byte"
	case ModeKanji:
		return "kanji"
	case ModeECI:
		return "eci"
	default:
		return "unknown"
	}
}

// Segment describes one encoded run within a scanned code.
type Segment struct {
	// Mode is how this run was encoded.
	Mode SegmentMode
	// Chars is the character count declared in the segment header.
	Chars int
	// Bytes is what this segment contributed to the content.
	Bytes []byte
}

// ScanResult describes a QR code read from an image.
//
// Note what is absent: the module matrix. The decoder does not expose it, and
// re-encoding the content would produce *a* matrix for the same text rather
// than *the* matrix that was scanned — the segmentation could differ. Returning
// that as the scanned symbol would be a quiet lie, so it is left out.
type ScanResult struct {
	// Content is the decoded text.
	Content string
	// Version is the symbol version, 1 to 40.
	Version int
	// Modules is the side length in modules, excluding the quiet zone.
	Modules int
	// ECC is the error-correction level the symbol was encoded at.
	ECC ECCLevel
	// Mask is the mask pattern applied, 0 to 7.
	Mask int
	// Segments is how the content was split across encoding modes.
	Segments []Segment
}

// Scan reads the first QR code found in img.
//
// It handles the images this library produces, including styled ones, and
// tolerates moderate rotation and downscaling. It has not been proven against
// camera photographs with perspective distortion, uneven lighting or blur; if
// that is your input, measure before relying on it.
func Scan(img image.Image) (*ScanResult, error) {
	if img == nil {
		return nil, fmt.Errorf("%w: no image given", ErrNoQRCode)
	}
	res, err := goqr.DecodeDetailed(img)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoQRCode, err)
	}

	out := &ScanResult{
		Content: res.Text,
		Version: res.Version,
		Modules: 4*res.Version + 17,
		ECC:     fromPigligEcc(res.Ecc),
		Mask:    res.Mask,
	}
	for _, s := range res.Segments {
		out.Segments = append(out.Segments, Segment{
			Mode:  segmentMode(s.Mode),
			Chars: s.NumChars,
			Bytes: s.Bytes,
		})
	}
	return out, nil
}

// ScanFile reads the first QR code in the image at path. PNG, JPEG and WebP are
// understood.
func ScanFile(path string) (*ScanResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("qr: opening image: %w", err)
	}
	defer f.Close()
	return ScanReader(f)
}

// ScanReader reads the first QR code in the image read from r. PNG, JPEG and
// WebP are understood.
func ScanReader(r io.Reader) (*ScanResult, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("qr: decoding image: %w", err)
	}
	return Scan(img)
}

// segmentMode maps the four-bit mode indicator from the specification onto the
// named constant.
func segmentMode(indicator int) SegmentMode {
	switch indicator {
	case 1:
		return ModeNumeric
	case 2:
		return ModeAlphanumeric
	case 4:
		return ModeByte
	case 7:
		return ModeECI
	case 8:
		return ModeKanji
	default:
		return ModeUnknown
	}
}

func fromPigligEcc(e goqr.Ecc) ECCLevel {
	switch e {
	case goqr.Low:
		return ECCLow
	case goqr.Medium:
		return ECCMedium
	case goqr.Quartile:
		return ECCQuartile
	case goqr.High:
		return ECCHigh
	}
	return ECCAuto
}
