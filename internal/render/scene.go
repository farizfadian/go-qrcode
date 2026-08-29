package render

import (
	"image"
	"image/color"
)

// Item is one drawing operation in a Scene. Only PathItem and ImageItem
// implement it; the sealed interface keeps renderers exhaustive.
type Item interface{ isItem() }

// PathItem fills a path with a single colour.
type PathItem struct {
	Path Path
	Fill color.Color
}

func (PathItem) isItem() {}

// ImageItem draws an image scaled into the rectangle at (X, Y) with size
// (W, H). When Clip is non-nil the image is masked by that path, which is how a
// logo gets rounded corners without either renderer knowing what a logo is.
//
// SVGMarkup is an optional vector version of the same picture. The SVG renderer
// embeds it instead of the bitmap, so it stays sharp at any zoom and keeps
// gradients, curves and text exactly as authored. The rasteriser ignores it and
// uses Img, so both renderers always have something to draw and neither can
// fail for want of the other.
type ImageItem struct {
	Img        image.Image
	SVGMarkup  string
	X, Y, W, H float64
	Clip       *Path
}

func (ImageItem) isItem() {}

// Scene is a complete, renderer-independent description of one drawing. Both
// the raster and the SVG renderer consume it, which is what guarantees their
// output matches.
type Scene struct {
	Width, Height int
	Background    color.Color // a zero-alpha colour means no background fill
	Items         []Item      // painted in order, first to last
}
