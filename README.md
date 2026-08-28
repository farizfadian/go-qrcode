# go-qrcode

Styled QR codes for Go: twelve per-module dot shapes, seven independently styled
and coloured finder patterns, a decorated centre logo, and byte-identical
geometry from both a raster and an SVG renderer. Existing Go QR libraries give
you a logo or a colour scheme; this one gives you the whole visual surface, and
proves every shape still scans with a round-trip decode test.

```go
q, err := qr.New(qr.Options{Content: "https://example.com"})
if err != nil {
    return err
}
return q.WritePNGFile("qr.png")
```

Only `Content` is required. Every other field has a working default, so the zero
value produces a conventional black-on-white code at 380 pixels.

## Status

Under construction. The design and the implementation plan live in
[`docs/superpowers/`](docs/superpowers/).

Runtime dependencies are limited to `github.com/piglig/go-qr` and
`golang.org/x/image`, both pinned so the module keeps a Go 1.22 floor.

## Credits

The feature surface and visual behaviour are modelled on
[`zxpsuper/qrcode-with-logos`](https://github.com/zxpsuper/qrcode-with-logos)
(MIT). No source was copied; the behaviour was reimplemented in Go.

## Licence

MIT. See [LICENSE](LICENSE).
