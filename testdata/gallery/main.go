// Command gallery renders one sample per dot shape and per finder shape into
// the directory given as its argument, for the README gallery. It lives under
// testdata so the go tool ignores it.
//
//	go run ./testdata/gallery docs/images
package main

import (
	"fmt"
	"os"

	"github.com/farizfadian/go-qrcode/qr"
)

const sample = "https://github.com/farizfadian/go-qrcode"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gallery <output-directory>")
		os.Exit(2)
	}
	out := os.Args[1]

	for _, d := range qr.DotTypes() {
		write(out+"/dot-"+d.String()+".png", qr.Options{
			Content: sample,
			Width:   240,
			Dots:    qr.DotOptions{Type: d, Color: "#1f2937"},
			Corners: qr.CornerOptions{Color: "#1f2937"},
		})
	}

	for _, c := range qr.CornerTypes() {
		write(out+"/corner-"+c.String()+".png", qr.Options{
			Content: sample,
			Width:   240,
			Dots:    qr.DotOptions{Color: "#94a3b8"},
			Corners: qr.CornerOptions{Type: c, Color: "#dc2626"},
		})
	}
}

func write(path string, opts qr.Options) {
	q, err := qr.New(opts)
	if err != nil {
		panic(err)
	}
	if err := q.WritePNGFile(path); err != nil {
		panic(err)
	}
	fmt.Println("wrote", path)
}
