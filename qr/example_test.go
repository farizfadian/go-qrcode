package qr_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/farizfadian/go-qrcode/qr"
)

// The simplest useful call: only Content is required, and every other field
// has a working default.
func ExampleNew() {
	q, err := qr.New(qr.Options{Content: "https://example.com"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(q.Content())
	fmt.Println(q.Modules(), "modules")
	fmt.Println(q.Image().Bounds().Dx(), "pixels")
	// Output:
	// https://example.com
	// 25 modules
	// 380 pixels
}

// Dots and finder patterns take their colour from Foreground unless given one
// of their own.
func ExampleNew_colours() {
	q, err := qr.New(qr.Options{
		Content:    "https://example.com",
		Width:      512,
		Foreground: "#1f2937",
		Background: "#ffffff",
		Corners:    qr.CornerOptions{Color: "#dc2626"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(q.Image().Bounds().Dx())
	// Output: 512
}

// A transparent background is a zero-alpha colour. Composite it over whatever
// your design needs — but keep the foreground darker than what sits behind it.
func ExampleNew_transparentBackground() {
	q, err := qr.New(qr.Options{
		Content:    "https://example.com",
		Background: "#00000000",
	})
	if err != nil {
		log.Fatal(err)
	}

	_, _, _, alpha := q.Image().At(0, 0).RGBA()
	fmt.Println("corner alpha:", alpha)
	// Output: corner alpha: 0
}

// SVG comes from the same geometry as the raster output, so the two cannot
// drift apart.
func ExampleQR_SVGString() {
	q, err := qr.New(qr.Options{Content: "https://example.com", Width: 256})
	if err != nil {
		log.Fatal(err)
	}

	markup, err := q.SVGString()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.HasPrefix(markup, "<svg "))
	fmt.Println(strings.Contains(markup, `viewBox="0 0 256 256"`))
	// Output:
	// true
	// true
}

// Leaving ECC unset lets the library choose: short content gets the highest
// protection because there is room for it, longer content steps down so the
// symbol does not grow and shrink its own modules.
func ExampleQR_ECC() {
	short, _ := qr.New(qr.Options{Content: "hello"})
	long, _ := qr.New(qr.Options{Content: strings.Repeat("a", 100)})

	fmt.Println("short:", short.ECC())
	fmt.Println("long: ", long.ECC())
	// Output:
	// short: H
	// long:  M
}

// Ask the library which shapes the build supports rather than hard-coding a
// list that can fall out of step with it. This output grows as shapes land,
// and the test fails if the documentation does not keep up.
func ExampleDotTypeNames() {
	fmt.Println(qr.DotTypeNames())
	fmt.Println(qr.CornerTypeNames())
	// Output:
	// [diamond dot dot-small rounded square star stripe stripe-column stripe-row tile]
	// [square]
}

// A shape named in the specification but not implemented is rejected, rather
// than quietly falling back to square.
func ExampleParseDotType() {
	if _, err := qr.ParseDotType("square"); err == nil {
		fmt.Println("square: ok")
	}
	if _, err := qr.ParseDotType("fluid"); errors.Is(err, qr.ErrUnknownShape) {
		fmt.Println("fluid: rejected, not implemented in this build")
	}
	// Output:
	// square: ok
	// fluid: rejected, not implemented in this build
}

// Errors are sentinel values, so callers can branch on what went wrong.
func ExampleNew_errorHandling() {
	_, err := qr.New(qr.Options{Content: ""})

	switch {
	case errors.Is(err, qr.ErrNoContent):
		fmt.Println("content was empty")
	case errors.Is(err, qr.ErrWidthTooSmall):
		fmt.Println("width cannot hold the module count")
	case err != nil:
		fmt.Println("something else:", err)
	}
	// Output: content was empty
}

// Serving a QR code over HTTP is four lines, because rendering targets any
// io.Writer.
func ExampleQR_PNG() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		q, err := qr.New(qr.Options{Content: r.URL.Query().Get("url")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_ = q.PNG(w)
	}

	// Exercise it without starting a server.
	q, _ := qr.New(qr.Options{Content: "https://example.com"})
	var buf bytes.Buffer
	if err := q.PNG(&buf); err != nil {
		log.Fatal(err)
	}
	_ = handler

	fmt.Println("png bytes written:", buf.Len() > 0)
	// Output: png bytes written: true
}
