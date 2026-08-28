package render

import (
	"image"
	"image/draw"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/vector"
)

// Raster renders sc into a new RGBA image. It never fails: a Scene is already
// validated by the time it reaches a renderer.
func Raster(sc Scene) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, sc.Width, sc.Height))
	if sc.Background != nil {
		if _, _, _, a := sc.Background.RGBA(); a > 0 {
			draw.Draw(dst, dst.Bounds(), image.NewUniform(sc.Background), image.Point{}, draw.Src)
		}
	}
	for _, it := range sc.Items {
		switch v := it.(type) {
		case PathItem:
			mask := rasterize(v.Path, sc.Width, sc.Height)
			draw.DrawMask(dst, dst.Bounds(), image.NewUniform(v.Fill), image.Point{},
				mask, image.Point{}, draw.Over)
		case ImageItem:
			drawImage(dst, v, sc.Width, sc.Height)
		}
	}
	return dst
}

// rasterize converts a path to a coverage mask. x/image/vector accumulates
// signed area and takes its absolute value clamped to 1, which is equivalent to
// the non-zero winding rule and therefore to SVG's default fill-rule. That is
// what makes a reversed inner subpath punch a hole.
func rasterize(p Path, w, h int) *image.Alpha {
	r := vector.NewRasterizer(w, h)
	for _, sp := range p.SubPaths {
		r.MoveTo(float32(sp.Start.X), float32(sp.Start.Y))
		for _, s := range sp.Segs {
			switch s.Kind {
			case SegLine:
				r.LineTo(float32(s.To.X), float32(s.To.Y))
			case SegQuad:
				r.QuadTo(float32(s.C1.X), float32(s.C1.Y), float32(s.To.X), float32(s.To.Y))
			case SegCube:
				r.CubeTo(float32(s.C1.X), float32(s.C1.Y),
					float32(s.C2.X), float32(s.C2.Y),
					float32(s.To.X), float32(s.To.Y))
			}
		}
		r.ClosePath()
	}
	mask := image.NewAlpha(image.Rect(0, 0, w, h))
	r.Draw(mask, mask.Bounds(), image.Opaque, image.Point{})
	return mask
}

func drawImage(dst *image.RGBA, it ImageItem, w, h int) {
	rect := image.Rect(int(it.X), int(it.Y), int(it.X+it.W), int(it.Y+it.H))
	scaled := image.NewRGBA(rect)
	xdraw.CatmullRom.Scale(scaled, rect, it.Img, it.Img.Bounds(), xdraw.Over, nil)
	if it.Clip == nil {
		draw.Draw(dst, rect, scaled, rect.Min, draw.Over)
		return
	}
	draw.DrawMask(dst, dst.Bounds(), scaled, image.Point{},
		rasterize(*it.Clip, w, h), image.Point{}, draw.Over)
}
