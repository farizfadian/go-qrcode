package qr

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// QR is a finished, immutable QR code. It is safe for concurrent use: every
// render method reads the same precomputed scene and writes nothing.
type QR struct {
	sc      render.Scene
	content string
	ecc     ECCLevel
	modules int
}

// New validates opts, encodes the content and builds the drawing. All
// configuration errors surface here; the render methods only ever report I/O
// failures.
func New(opts Options) (*QR, error) {
	if opts.Content == "" {
		return nil, ErrNoContent
	}
	if opts.Logo != nil {
		return nil, ErrLogoUnsupported
	}
	o := opts.withDefaults()

	bg, err := render.ParseColor(o.Background)
	if err != nil {
		return nil, fmt.Errorf("%w: background: %v", ErrBadColor, err)
	}
	dotCol, err := render.ParseColor(o.Dots.Color)
	if err != nil {
		return nil, fmt.Errorf("%w: dots: %v", ErrBadColor, err)
	}
	cornerCol, err := render.ParseColor(o.Corners.Color)
	if err != nil {
		return nil, fmt.Errorf("%w: corners: %v", ErrBadColor, err)
	}

	ecc := o.resolveECC()
	mods, err := defaultEncoder().Encode(o.Content, ecc)
	if err != nil {
		return nil, err
	}
	m, err := newMatrix(mods)
	if err != nil {
		return nil, err
	}
	l, err := newLayout(m.Size(), o.Margin, o.Width)
	if err != nil {
		return nil, err
	}

	ctx := newShapeContext(m, l, nil)
	sc := render.Scene{Width: o.Width, Height: o.Width, Background: bg}

	var dots render.Path
	for y := 0; y < m.Size(); y++ {
		for x := 0; x < m.Size(); x++ {
			if !ctx.Dark(x, y) {
				continue
			}
			dots = dots.Append(dotPath(o.Dots.Type, ctx, x, y))
		}
	}
	if !dots.IsEmpty() {
		sc.Items = append(sc.Items, render.PathItem{Path: dots, Fill: dotCol})
	}

	var corners render.Path
	n := m.Size()
	for _, c := range [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}} {
		px, py, s := l.Rect(c[0], c[1])
		corners = corners.Append(cornerPath(o.Corners.Type, px, py, s, o.Corners.Radius))
	}
	sc.Items = append(sc.Items, render.PathItem{Path: corners, Fill: cornerCol})

	return &QR{sc: sc, content: o.Content, ecc: ecc, modules: n}, nil
}

// Image returns the rendered raster image. Each call renders afresh, so the
// returned image is the caller's to modify.
func (q *QR) Image() image.Image { return render.Raster(q.sc) }

// PNG writes the code to w as a PNG.
func (q *QR) PNG(w io.Writer) error { return png.Encode(w, q.Image()) }

// JPEG writes the code to w as a JPEG at the given quality, 1 to 100. A
// transparent background is flattened onto white, since JPEG has no alpha.
func (q *QR) JPEG(w io.Writer, quality int) error {
	return jpeg.Encode(w, flatten(q.Image()), &jpeg.Options{Quality: quality})
}

// WritePNGFile renders the code and writes it to path.
func (q *QR) WritePNGFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := q.PNG(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Content returns the text the code encodes.
func (q *QR) Content() string { return q.content }

// Modules returns the symbol's side length in modules, excluding the quiet
// zone. It is useful for laying a code out beside other content.
func (q *QR) Modules() int { return q.modules }

// ECC returns the error-correction level actually used, which differs from the
// requested one when Options.ECC was ECCAuto.
func (q *QR) ECC() ECCLevel { return q.ecc }

// flatten composites src over opaque white. JPEG has no alpha channel, so a
// transparent background would otherwise encode as black.
func flatten(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
