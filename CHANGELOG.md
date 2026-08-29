# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

Nothing has been tagged yet. Everything below ships in the first release.

### Added

#### Library

- `qr.New(qr.Options)` returning an immutable `*QR` that is safe to render from
  many goroutines at once.
- Output as PNG, JPEG, SVG and `image.Image`, all from one shared scene
  description so the raster and vector forms cannot drift apart.
- All twelve dot shapes: `square`, `dot`, `dot-small`, `tile`, `rounded`,
  `diamond`, `star`, `fluid`, `fluid-line`, `stripe`, `stripe-row` and
  `stripe-column`. The last five are neighbour-aware, merging runs of modules
  into continuous strokes.
- All seven finder-pattern shapes: `square`, `rounded`, `circle`,
  `rounded-circle`, `circle-rounded`, `circle-star` and `circle-diamond`, with
  independently configurable inner and outer radii.
- A centred logo from an `image.Image`, a file path or an `io.Reader`, with a
  frame, rounded corners on both the frame and the image, automatic sizing, and
  an enforced occlusion budget. The modules beneath it are excluded before
  rendering rather than painted over, so nothing shows around a rounded corner.
  `QR.HiddenModules` and `QR.LogoBudget` report how much of the
  error-correction allowance a design spends.
- Independent foreground, background, dot and finder colours; hex with or
  without `#`, and `#00000000` for a transparent background.
- All four error-correction levels plus automatic selection driven by content
  length.
- `ParseDotType`, `ParseCornerType`, `ParseECCLevel` and the matching `…Types`
  and `…Names` listings, derived from the shape registry so configuration-driven
  callers stay correct as shapes are added.

#### CLI

- `qrgen`, rendering to PNG, JPEG or SVG. The format follows the `-out`
  extension unless `-format` overrides it, and the flag help lists exactly the
  shapes the build can draw.
- `-logo` and its companions on the CLI, reporting how much of the
  error-correction budget the logo spent.

#### Project

- CI across Go 1.22, 1.24 and 1.25 on Linux, Windows and macOS, with the race
  detector, a `gofmt` check, a `go mod tidy` check and a dependency-footprint
  assertion.
- Release workflow triggered by a `v*` tag, publishing eleven cross-compiled
  `qrgen` binaries with checksums.

### Notes on behaviour that differs from the reference library

The visual behaviour is modelled on
[`zxpsuper/qrcode-with-logos`](https://github.com/zxpsuper/qrcode-with-logos).
Three deviations are deliberate:

- **`Margin` is measured in modules, not pixels.** The reference treats its
  margin as pixels, which yields an effective quiet zone of roughly 0.65 modules
  at its default width — far below the four modules ISO/IEC 18004 requires.
- **Run-merging shapes cannot grow into excluded regions.** In the reference a
  `stripe` run can extend across a finder pattern, because the run test checks
  only whether a module is dark.
- **An unimplemented shape is an error.** Rather than silently falling back to
  `square`, `New` reports which shapes the build supports.
- **`circle-diamond` is inscribed in the 3x3 core.** The reference rotates a
  full three-module square, giving a diagonal of 4.24 modules that nearly fills
  the ring's five-module gap. A reader looks for dark:light:dark:light:dark in
  the ratio 1:1:3:1:1 through the finder's centre; the reference geometry
  measures 1:0.38:4.24:0.38:1 and is not recognised as a finder pattern at all.
  Inscribing the diamond restores exactly 1:1:3:1:1. Measured, not assumed: the
  reference proportions fail to decode and these succeed.

### Known limitation

Go 1.22 cannot produce a loadable test binary for the `qr` package on current
macOS; it aborts with `missing LC_UUID load command`. The cause is inside the Go
toolchain — the other packages built by the same job run fine, and every other
Go version and operating system passes, including Go 1.22 on Linux and Windows.
The declared floor stays at 1.22; macOS users need Go 1.23 or later.

- Lossless WebP output, `QR.WebP` and `qrgen -out x.webp`. Measured 19 to 48
  percent smaller than the equivalent PNG depending on dot shape. This adds a
  third runtime dependency; the reasoning is in
  [ADR 0001](docs/adr/0001-webp-output-dependency.md).
- Reading QR codes: `Scan`, `ScanFile` and `ScanReader`, returning content plus
  version, module count, ECC level, mask and per-segment encoding modes. It uses
  the decoder already inside `piglig/go-qr`, so it costs no new dependency.
  `qrgen -scan` and `-scan-details` expose it on the command line.
- `LogoOptions.SVGMarkup`, an optional vector version of the logo embedded
  directly into SVG output as a nested `<svg>`, so gradients, rounded corners,
  strokes and text survive exactly as authored. It sits alongside the raster
  logo rather than replacing it, which keeps every output format working and
  every configuration error at `New`.
- Colour-scheme validation. `ErrLowContrast` fires below a WCAG contrast ratio
  of 3.5, and `ErrInvertedPolarity` fires whenever the foreground is lighter
  than the background. Both thresholds were measured against a real decoder:
  3.54 was the lowest ratio that still decoded and 2.96 the highest that failed,
  while inverted schemes failed at every ratio up to 21. Dots and finder
  patterns are checked separately.

### Fixed

- `New` built its dot and corner paths with a quadratic append. `Path.Append`
  copies the whole subpath slice on every call — which is what makes a `Path`
  safe to share — so calling it once per module copied the accumulated work
  repeatedly. A version 30 symbol took 1.36 seconds and allocated 1.2 GB. It now
  accumulates into one slice: 4.4 ms and 4.0 MB, and the default path improved
  4.4x in time and 19x in allocation as well. Found by a benchmark; every test
  passed before and after.

### Not yet implemented

---

## Releasing

1. Update `Version` in `qr/doc.go` to the new number, without the `v`.
2. Update this file: move `[Unreleased]` entries under the new version heading.
3. Commit, then tag and push:

   ```bash
   git tag v0.1.0
   git push origin main --tags
   ```

The release workflow refuses to publish when the tag and `qr.Version` disagree.
That check matters because the Go module proxy caches a tag permanently: a
published version can never be corrected, only superseded.
