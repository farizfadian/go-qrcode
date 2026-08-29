package qr

import (
	"bytes"
	"errors"
	"image/png"
	"strings"
	"sync"
	"testing"
)

const testURL = "https://github.com/farizfadian/go-qrcode"

func TestNewRejectsEmptyContent(t *testing.T) {
	if _, err := New(Options{}); !errors.Is(err, ErrNoContent) {
		t.Fatalf("error = %v, want ErrNoContent", err)
	}
}

func TestZeroValueOptionsProducesAScannableCode(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	b := q.Image().Bounds()
	if b.Dx() != 380 || b.Dy() != 380 {
		t.Errorf("size = %dx%d, want 380x380 by default", b.Dx(), b.Dy())
	}
	assertDecodes(t, q.Image(), testURL)
}

func TestAllFourECCLevelsDecode(t *testing.T) {
	for _, ecc := range []ECCLevel{ECCLow, ECCMedium, ECCQuartile, ECCHigh} {
		t.Run(ecc.String(), func(t *testing.T) {
			requireDecodableBaseline(t, testURL, ecc)
			q, err := New(Options{Content: testURL, ECC: ecc, Width: 512})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			assertDecodes(t, q.Image(), testURL)
		})
	}
}

func TestLongContentDecodes(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("abcdefghij", 70)
	requireDecodableBaseline(t, long, ECCMedium)
	q, err := New(Options{Content: long, ECC: ECCMedium, Width: 1000})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	assertDecodes(t, q.Image(), long)
}

func TestPNGWritesADecodableImage(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var buf bytes.Buffer
	if err := q.PNG(&buf); err != nil {
		t.Fatalf("PNG: %v", err)
	}
	img, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	assertDecodes(t, img, testURL)
}

func TestTransparentBackgroundDecodesWhenFlattened(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL, Background: "#00000000"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	assertDecodes(t, flattenOntoWhite(q.Image()), testURL)
}

func TestNewRejectsUnparseableColour(t *testing.T) {
	_, err := New(Options{Content: testURL, Foreground: "not-a-colour"})
	if !errors.Is(err, ErrBadColor) {
		t.Fatalf("error = %v, want ErrBadColor", err)
	}
}

func TestNewRejectsWidthTooSmall(t *testing.T) {
	long := strings.Repeat("a", 1200)
	if _, err := New(Options{Content: long, Width: 20}); !errors.Is(err, ErrWidthTooSmall) {
		t.Fatalf("error = %v, want ErrWidthTooSmall", err)
	}
}

func TestQRIsSafeForConcurrentUse(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var wg sync.WaitGroup
	outs := make([][]byte, 8)
	for i := range outs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var buf bytes.Buffer
			if err := q.PNG(&buf); err != nil {
				t.Errorf("PNG: %v", err)
				return
			}
			outs[i] = buf.Bytes()
		}(i)
	}
	wg.Wait()
	for i := 1; i < len(outs); i++ {
		if !bytes.Equal(outs[0], outs[i]) {
			t.Fatalf("goroutine %d produced different bytes; *QR is not immutable", i)
		}
	}
}

func TestSVGAndPNGDescribeTheSameGeometry(t *testing.T) {
	requireDecodableBaseline(t, testURL, ECCAuto)
	q, err := New(Options{Content: testURL, Width: 512})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	svg, err := q.SVGString()
	if err != nil {
		t.Fatalf("SVGString: %v", err)
	}
	if !strings.Contains(svg, `viewBox="0 0 512 512"`) {
		t.Errorf("viewBox missing or wrong:\n%s", svg[:min(200, len(svg))])
	}
	// Two path elements: one for all dots, one for all corners.
	if n := strings.Count(svg, "<path "); n != 2 {
		t.Errorf("path elements = %d, want 2", n)
	}
	// The rasterised form must still decode, proving the shared scene is sound.
	assertDecodes(t, q.Image(), testURL)
}

func TestSVGWritesToAWriter(t *testing.T) {
	q, err := New(Options{Content: testURL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var buf bytes.Buffer
	if err := q.SVG(&buf); err != nil {
		t.Fatalf("SVG: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "<svg ") || !strings.HasSuffix(buf.String(), "</svg>") {
		t.Errorf("output is not a standalone svg document: %.60s...", buf.String())
	}
}

func TestAutoECCScalesWithContentLength(t *testing.T) {
	for _, tc := range []struct {
		content string
		want    ECCLevel
	}{
		{strings.Repeat("a", 10), ECCHigh},
		{strings.Repeat("a", 20), ECCQuartile},
		{strings.Repeat("a", 50), ECCMedium},
	} {
		q, err := New(Options{Content: tc.content})
		if err != nil {
			t.Fatalf("New(%d chars): %v", len(tc.content), err)
		}
		if q.ECC() != tc.want {
			t.Errorf("%d chars: ECC = %v, want %v", len(tc.content), q.ECC(), tc.want)
		}
	}
}
