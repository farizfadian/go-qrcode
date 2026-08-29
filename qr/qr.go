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

	"github.com/HugoSmits86/nativewebp"

	"github.com/farizfadian/go-qrcode/internal/render"
)

// QR is a finished, immutable QR code. It is safe for concurrent use: every
// render method reads the same precomputed scene and writes nothing.
type QR struct {
	sc      render.Scene
	content string
	ecc     ECCLevel
	modules int

	// hiddenModules is how many modules the logo covers, and logoBudget how
	// many the error-correction level allows. Both are zero without a logo.
	hiddenModules int
	logoBudget    int
}

// New validates opts, encodes the content and builds the drawing. All
// configuration errors surface here; the render methods only ever report I/O
// failures.
func New(opts Options) (*QR, error) {
	if opts.Content == "" {
		return nil, ErrNoContent
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

	// Refuse a colour scheme that cannot be read. Both checks are measured:
	// decoding fails below a contrast ratio of about 3, and an inverted code
	// fails at any contrast at all.
	if err := checkContrast(bg, map[string]color.RGBA{
		"dots":    dotCol,
		"corners": cornerCol,
	}); err != nil {
		return nil, err
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

	// The logo is planned before anything is drawn, because the modules it
	// covers must never be rendered in the first place. Painting over them
	// would leave dark fringes around a rounded corner and would spend work on
	// figures nobody sees.
	var (
		plan     *logoPlan
		excluded func(x, y int) bool
		hidden   int
		budget   int
	)
	if o.Logo != nil {
		plan, err = planLogo(*o.Logo, o.Background, l, ecc, m)
		if err != nil {
			return nil, err
		}
		excluded = plan.hides(l)
		hidden = countHidden(m, excluded)
		budget = plan.maxHidden
		if hidden > budget {
			return nil, fmt.Errorf(
				"%w: it covers %d of %d modules but %v error correction allows only %d; "+
					"reduce Logo.Size or raise the ECC level",
				ErrLogoTooLarge, hidden, m.Size()*m.Size(), ecc, budget)
		}
	}

	ctx := newShapeContext(m, l, excluded)
	sc := render.Scene{Width: o.Width, Height: o.Width, Background: bg}

	// Subpaths are accumulated into one slice and wrapped in a Path at the end.
	//
	// The obvious loop — `dots = dots.Append(dotPath(...))` — is quadratic:
	// Path.Append copies the whole slice each call, which is what makes it safe
	// to share a Path, so calling it once per module copies the accumulated
	// work N times over. A version 30 symbol has around nine thousand dark
	// modules, and that cost 1.36 seconds and 1.2 GB before this was measured.
	// Every test passed throughout; only a benchmark showed it.
	n := m.Size()
	dotSubs := make([]render.SubPath, 0, n*n/2)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if !ctx.Dark(x, y) {
				continue
			}
			dotSubs = append(dotSubs, dotPath(o.Dots.Type, ctx, x, y).SubPaths...)
		}
	}
	if len(dotSubs) > 0 {
		sc.Items = append(sc.Items, render.PathItem{
			Path: render.Path{SubPaths: dotSubs},
			Fill: dotCol,
		})
	}

	cornerSubs := make([]render.SubPath, 0, 9)
	for _, c := range [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}} {
		px, py, s := l.Rect(c[0], c[1])
		cornerSubs = append(cornerSubs,
			cornerPath(o.Corners.Type, px, py, s, o.Corners.Radius).SubPaths...)
	}
	sc.Items = append(sc.Items, render.PathItem{
		Path: render.Path{SubPaths: cornerSubs},
		Fill: cornerCol,
	})

	// The logo is painted last so it covers whatever lies beneath.
	if plan != nil {
		sc.Items = append(sc.Items, plan.items()...)
	}

	return &QR{
		sc:            sc,
		content:       o.Content,
		ecc:           ecc,
		modules:       n,
		hiddenModules: hidden,
		logoBudget:    budget,
	}, nil
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

// WebP writes the code to w as a lossless WebP.
//
// WebP is typically 30 to 40 percent smaller than the equivalent PNG for an
// image like a QR code, which is flat colour with hard edges. The encoding is
// lossless, so the modules stay exact.
func (q *QR) WebP(w io.Writer) error {
	return nativewebp.Encode(w, q.Image(), nil)
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

// HiddenModules returns how many modules the logo covers, or zero when there is
// no logo. Together with LogoBudget it shows how much of the error-correction
// allowance a design is spending.
func (q *QR) HiddenModules() int { return q.hiddenModules }

// LogoBudget returns how many modules the chosen error-correction level allows
// a logo to cover, or zero when there is no logo.
func (q *QR) LogoBudget() int { return q.logoBudget }
