// Command checkimg decodes the PNG files given as arguments and reports whether
// each one scans. It lives under testdata so the go tool ignores it.
package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/makiuchi-d/gozxing"
	zxqr "github.com/makiuchi-d/gozxing/qrcode"
)

func main() {
	bad := 0
	for _, path := range os.Args[1:] {
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("%-40s OPEN ERROR %v\n", path, err)
			bad++
			continue
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			fmt.Printf("%-40s PNG ERROR %v\n", path, err)
			bad++
			continue
		}
		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			fmt.Printf("%-40s BITMAP ERROR %v\n", path, err)
			bad++
			continue
		}
		res, err := zxqr.NewQRCodeReader().Decode(bmp, nil)
		if err != nil {
			fmt.Printf("%-40s DOES NOT SCAN (%v)\n", path, err)
			bad++
			continue
		}
		fmt.Printf("%-40s scans -> %s\n", path, res.GetText())
	}
	if bad > 0 {
		os.Exit(1)
	}
}
