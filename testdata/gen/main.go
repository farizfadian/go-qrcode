// Command gen renders sample QR codes for eyeballing during development. It
// lives under testdata so `go build ./...` and `go test ./...` ignore it.
//
// Usage: go run ./testdata/gen <output-directory>
package main

import (
	"fmt"
	"os"

	"github.com/farizfadian/go-qrcode/qr"
)

const sample = "https://github.com/farizfadian/go-qrcode"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gen <output-directory>")
		os.Exit(2)
	}
	out := os.Args[1]

	// Exactly the two-line snippet the README promises.
	q, err := qr.New(qr.Options{Content: sample})
	if err != nil {
		panic(err)
	}
	if err := q.WritePNGFile(out + "/default.png"); err != nil {
		panic(err)
	}
	fmt.Printf("default:  %d modules, ECC %v\n", q.Modules(), q.ECC())

	q2, err := qr.New(qr.Options{
		Content: sample,
		Width:   512,
		Dots:    qr.DotOptions{Color: "#1f2937"},
		Corners: qr.CornerOptions{Color: "#dc2626"},
	})
	if err != nil {
		panic(err)
	}
	if err := q2.WritePNGFile(out + "/coloured.png"); err != nil {
		panic(err)
	}
	fmt.Printf("coloured: %d modules, ECC %v\n", q2.Modules(), q2.ECC())

	svg, err := q2.SVGString()
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(out+"/coloured.svg", []byte(svg), 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("svg:      %d bytes\n", len(svg))
}
