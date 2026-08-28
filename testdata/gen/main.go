// Command gen renders sample QR codes for eyeballing during development. It
// lives under testdata so `go build ./...` and `go test ./...` ignore it.
package main

import (
	"fmt"
	"os"

	"github.com/farizfadian/go-qrcode/qr"
)

func main() {
	out := os.Args[1]

	// Exactly the two-line snippet the README promises.
	q, err := qr.New(qr.Options{Content: "https://github.com/farizfadian/go-qrcode"})
	if err != nil {
		panic(err)
	}
	if err := q.WritePNGFile(out + "/default.png"); err != nil {
		panic(err)
	}
	fmt.Printf("default:  %d modules, ECC %v\n", q.Modules(), q.ECC())

	q2, err := qr.New(qr.Options{
		Content: "https://github.com/farizfadian/go-qrcode",
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
}
