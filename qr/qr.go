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
	"strings"

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
	// Reject a shape this build does not implement. Falling back to square
	// would hand back a code that silently does not look like what was asked
	// for, which is harder to notice than an error.
	if _, ok := dotFuncs[opts.Dots.Type]; !ok {
		return nil, fmt.Errorf("%w: dot shape %q; available: %s",
			ErrUnknownShape, opts.Dots.Type, strings.Join(DotTypeNames(), ", "))
	}
	if _, ok := cornerFuncs[opts.Corners.Type]; !ok {
		return nil, fmt.Errorf("%w: corner shape %q; available: %s",
			ErrUnknownShape, opts.Corners.Type, strings.Join(CornerTypeNames(), ", "))
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

// SVG writes the code to w as standalone SVG markup.
func (q *QR) SVG(w io.Writer) error {
	s, err := q.SVGString()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// SVGString returns the code as standalone SVG markup. Its geometry is the same
// as Image produces: both renderers consume one scene.
func (q *QR) SVGString() (string, error) { return render.SVG(q.sc) }

// flatten composites src over opaque white. JPEG has no alpha channel, so a
// transparent background would otherwise encode as black.
func flatten(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
