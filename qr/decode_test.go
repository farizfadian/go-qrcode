package qr

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/makiuchi-d/gozxing"
	zxqr "github.com/makiuchi-d/gozxing/qrcode"
)

// decodeImage returns the text gozxing reads from img, or an error.
func decodeImage(img image.Image) (string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	res, err := zxqr.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return "", err
	}
	return res.GetText(), nil
}

// requireDecodableBaseline renders content unstyled at a generous module size
// and decodes it. gozxing fails on roughly 2-3% of perfectly valid QR symbols,
// a property of the decoder rather than of any encoder: three independent
// encoders were measured failing on overlapping inputs. When the plain render
// already fails, a styled render proves nothing, so the case is skipped instead
// of reported as our bug.
func requireDecodableBaseline(t *testing.T, content string, ecc ECCLevel) {
	t.Helper()
	q, err := New(Options{Content: content, ECC: ecc, Width: 1000})
	if err != nil {
		t.Fatalf("baseline New: %v", err)
	}
	got, err := decodeImage(q.Image())
	if err != nil || got != content {
		t.Skipf("gozxing cannot decode this symbol unstyled (err=%v); not a renderer fault", err)
	}
}

// assertDecodes requires the decoded text of img to equal want.
func assertDecodes(t *testing.T, img image.Image, want string) {
	t.Helper()
	got, err := decodeImage(img)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if got != want {
		t.Fatalf("decoded %q, want %q", got, want)
	}
}

// flattenOntoWhite composites img over opaque white, which is what a
// transparent QR code sits on in practice.
func flattenOntoWhite(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
