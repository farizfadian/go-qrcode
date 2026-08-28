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

## Command line

```bash
go install github.com/farizfadian/go-qrcode/cmd/qrgen@latest

qrgen -out qr.png "https://example.com"
qrgen -out card.svg -width 512 -dot-color '#1f2937' -corner-color '#dc2626' "hello"
```

The output format follows the `-out` extension unless `-format` overrides it.
Run `qrgen -h` to see the shapes your build supports — the list is read from the
library, so it is always accurate.

## Status

Under construction. Working today: PNG, JPEG and SVG output, `square` dots and
`square` finder patterns, independent colours, transparency, all four
error-correction levels, and the CLI. Still to come: the remaining eleven dot
shapes, six finder shapes, and logo support.

The design and the implementation plan live in
[`docs/superpowers/`](docs/superpowers/); changes are tracked in
[CHANGELOG.md](CHANGELOG.md).

Runtime dependencies are limited to `github.com/piglig/go-qr` and
`golang.org/x/image`, both pinned so the module keeps a Go 1.22 floor. CI
asserts that footprint on every push, so it cannot grow by accident.

## Credits

The feature surface and visual behaviour are modelled on
[`zxpsuper/qrcode-with-logos`](https://github.com/zxpsuper/qrcode-with-logos)
(MIT). No source was copied; the behaviour was reimplemented in Go.

## Licence

MIT. See [LICENSE](LICENSE).
