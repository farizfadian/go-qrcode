// Package qr generates QR codes with styled modules, styled finder patterns and
// a decorated centre logo, rendered identically to PNG, JPEG and SVG.
//
// The simplest useful call is two lines:
//
//	q, err := qr.New(qr.Options{Content: "https://example.com"})
//	err = q.WritePNGFile("qr.png")
//
// Every field of Options other than Content has a working default, so the zero
// value produces a conventional black-on-white code.
package qr

// Version is the library version, set at release time.
const Version = "0.0.0-dev"
