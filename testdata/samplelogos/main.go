// Command samplelogos writes the neutral logo marks used by the README gallery,
// as matching PNG and SVG pairs, and renders a QR code with each.
//
// The marks are generated rather than downloaded on purpose. Real brand logos —
// Google, WhatsApp, Instagram — are registered trademarks, and putting one in an
// unaffiliated public repository invites a takedown for no benefit. Geometric
// marks make the same point about the feature and belong to this project.
//
// The marks land in <dir>/logos and the QR codes carrying them in <dir>, because
// CI decodes every image directly under docs/images and a logo is not a QR code.
//
//	go run ./testdata/samplelogos docs/images
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/farizfadian/go-qrcode/qr"
)

const sample = "https://github.com/farizfadian/go-qrcode"

type mark struct {
	name  string
	svg   string
	paint func(x, y, n float64) bool
	fill  color.RGBA
}

var marks = []mark{
	{
		name: "logo-letter",
		fill: color.RGBA{0x0f, 0x76, 0x6e, 0xff},
		svg: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
			`<rect x="6" y="6" width="88" height="88" rx="20" fill="#0f766e"/>` +
			`<path d="M30 28 L72 28 L72 41 L44 41 L44 53 L67 53 L67 66 L44 66 L44 78 L30 78 Z" fill="#ffffff"/>` +
			`</svg>`,
		paint: func(x, y, n float64) bool {
			// A rounded panel with a blocky letter F knocked out of it.
			if !roundedSquare(x, y, n, 0.06, 0.20) {
				return false
			}
			stem := between(x, n, 0.30, 0.44) && between(y, n, 0.28, 0.78)
			top := between(y, n, 0.28, 0.41) && between(x, n, 0.30, 0.72)
			mid := between(y, n, 0.53, 0.66) && between(x, n, 0.30, 0.67)
			return !(stem || top || mid)
		},
	},
	{
		name: "logo-target",
		fill: color.RGBA{0xdc, 0x26, 0x26, 0xff},
		svg: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
			`<circle cx="50" cy="50" r="44" fill="none" stroke="#dc2626" stroke-width="14"/>` +
			`<circle cx="50" cy="50" r="14" fill="#dc2626"/>` +
			`</svg>`,
		paint: func(x, y, n float64) bool {
			d := dist(x, y, n)
			return (d > 0.37 && d < 0.44) || d < 0.14
		},
	},
	{
		name: "logo-chevron",
		fill: color.RGBA{0x4f, 0x46, 0xe5, 0xff},
		svg: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
			`<path d="M50 12 L88 76 L68 76 L50 46 L32 76 L12 76 Z" fill="#4f46e5"/>` +
			`</svg>`,
		paint: func(x, y, n float64) bool {
			// An outlined chevron: a wide triangle minus a narrower one.
			return triangle(x, y, n, 0.12, 0.76) && !triangle(x, y, n, 0.32, 0.94)
		},
	},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: samplelogos <output-directory>")
		os.Exit(2)
	}
	out := os.Args[1]

	// The source art goes in its own directory. Everything directly under
	// docs/images is a QR code, and CI decodes every one of them; a logo file
	// sitting there would fail that check for the obvious reason.
	art := out + "/logos"
	if err := os.MkdirAll(art, 0o755); err != nil {
		panic(err)
	}

	for _, m := range marks {
		const n = 512
		img := image.NewRGBA(image.Rect(0, 0, n, n))
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				if m.paint(float64(x), float64(y), n) {
					img.Set(x, y, m.fill)
				}
			}
		}
		writePNG(art+"/"+m.name+".png", img)
		writeFile(art+"/"+m.name+".svg", m.svg)

		// A QR code carrying the mark, for the gallery.
		q, err := qr.New(qr.Options{
			Content: sample,
			Width:   400,
			Dots:    qr.DotOptions{Type: qr.DotFluid, Color: "#1f2937"},
			Corners: qr.CornerOptions{Type: qr.CornerCircle, Color: "#1f2937"},
			Logo: &qr.LogoOptions{
				Image:        img,
				SVGMarkup:    m.svg,
				Radius:       14,
				BorderWidth:  10,
				BorderRadius: 16,
			},
		})
		if err != nil {
			panic(err)
		}
		if err := q.WritePNGFile(out + "/qr-" + m.name + ".png"); err != nil {
			panic(err)
		}
		fmt.Printf("%-14s hides %d of %d allowed\n", m.name, q.HiddenModules(), q.LogoBudget())
	}
}

func between(v, n, lo, hi float64) bool { return v/n > lo && v/n < hi }

func dist(x, y, n float64) float64 {
	dx, dy := x/n-0.5, y/n-0.5
	return math.Hypot(dx, dy)
}

// roundedSquare reports whether (x, y) is inside a square inset by `inset` with
// corners of the given radius, all as fractions of n.
func roundedSquare(x, y, n, inset, radius float64) bool {
	fx, fy := x/n, y/n
	lo, hi := inset, 1-inset
	if fx < lo || fx > hi || fy < lo || fy > hi {
		return false
	}
	cx := clamp(fx, lo+radius, hi-radius)
	cy := clamp(fy, lo+radius, hi-radius)
	return math.Hypot(fx-cx, fy-cy) <= radius
}

// triangle reports whether (x, y) is inside an upward triangle whose base sits
// at `base` and whose sides start `spread` in from the edges.
func triangle(x, y, n, spread, base float64) bool {
	fx, fy := x/n, y/n
	if fy < 0.12 || fy > base {
		return false
	}
	t := (fy - 0.12) / (base - 0.12)
	halfWidth := t * (0.5 - spread)
	return math.Abs(fx-0.5) <= halfWidth
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		panic(err)
	}
}
