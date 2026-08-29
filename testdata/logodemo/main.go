// Command logodemo renders QR codes with a synthetic logo, for eyeballing.
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/farizfadian/go-qrcode/qr"
)

// mark draws a simple high-contrast glyph on a transparent field.
func mark(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	fg := color.RGBA{0x0f, 0x76, 0x6e, 0xff}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx, fy := float64(x)/float64(size), float64(y)/float64(size)
			inBar := fx > 0.18 && fx < 0.38 && fy > 0.18 && fy < 0.82
			inArm := fy > 0.4 && fy < 0.6 && fx > 0.18 && fx < 0.78
			inFoot := fy > 0.62 && fy < 0.82 && fx > 0.18 && fx < 0.68
			if inBar || inArm || inFoot {
				img.Set(x, y, fg)
			}
		}
	}
	return img
}

func main() {
	out := os.Args[1]
	logo := mark(256)

	for _, c := range []struct {
		name string
		opts qr.Options
	}{
		{"logo-auto", qr.Options{Content: "https://github.com/farizfadian/go-qrcode", Width: 720,
			Logo: &qr.LogoOptions{Image: logo}}},
		{"logo-fluid", qr.Options{Content: "https://github.com/farizfadian/go-qrcode", Width: 720,
			Dots:    qr.DotOptions{Type: qr.DotFluid, Color: "#1f2937"},
			Corners: qr.CornerOptions{Type: qr.CornerCircle, Color: "#dc2626"},
			Logo:    &qr.LogoOptions{Image: logo, Radius: 14, BorderWidth: 12, BorderRadius: 18}}},
		{"logo-stripe", qr.Options{Content: "https://github.com/farizfadian/go-qrcode", Width: 720,
			Dots:    qr.DotOptions{Type: qr.DotStripe, Color: "#4f46e5"},
			Corners: qr.CornerOptions{Type: qr.CornerCircleRounded, Color: "#4f46e5"},
			Logo:    &qr.LogoOptions{Image: logo, Size: 0.18, Radius: 10}}},
	} {
		q, err := qr.New(c.opts)
		if err != nil {
			fmt.Printf("%-12s ERROR %v\n", c.name, err)
			continue
		}
		if err := q.WritePNGFile(out + "/" + c.name + ".png"); err != nil {
			panic(err)
		}
		fmt.Printf("%-12s %d modules, ECC %v, hides %d of %d allowed\n",
			c.name, q.Modules(), q.ECC(), q.HiddenModules(), q.LogoBudget())
	}
}
