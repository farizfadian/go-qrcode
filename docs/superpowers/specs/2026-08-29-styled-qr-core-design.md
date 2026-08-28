# Styled QR core — design

Date: 2026-08-29
Status: approved by Fariz, ready for implementation planning
Scope: CLAUDE.md phases 1–4 in a single branch

## 1. What we are building

A pure-Go QR generator whose distinguishing features are per-module dot shapes,
independently styled finder patterns, a decorated centred logo, and identical
output from a raster and an SVG renderer.

Delivered in one branch:

- Encoder interface backed by a third-party pure-Go encoder
- Module matrix with finder-pattern classification
- Shared vector path model
- Raster renderer (PNG, JPEG) and SVG renderer
- 12 dot types, 7 corner types, independent colours and corner radii
- Logo with border, radii, background, auto-sizing and safe-zone clearing

Out of scope: everything CLAUDE.md §1 lists as a non-goal, plus the CLI and the
built-in encoder (phases 5 and 6).

## 2. Architecture

### 2.1 Package layout

```
go.mod · LICENSE · README.md · CLAUDE.md
qr/
  qr.go            Options, QR, New, PNG/JPEG/SVG/SVGString/Image/WritePNGFile
  options.go       defaults, validation, sentinel errors
  encode.go        Encoder interface + piglig-backed implementation
  matrix.go        Matrix: modules plus classification
  layout.go        module -> pixel mapping
  logo.go          decode, resize, border, radii, safe zone, ECC budget
  shape.go         ShapeContext, DotType, CornerType, shape registry
  shape_dots.go    12 dot types -> render.Path
  shape_corners.go 7 corner types -> render.Path
internal/render/
  path.go          Path, SubPath, MoveTo/LineTo/CurveTo/Close, Bounds
  color.go         hex parsing, alpha, named passthrough
  scene.go         Scene: background plus ordered items
  raster.go        Scene -> *image.RGBA
  svg.go           Scene -> string
```

Two additions to the layout in CLAUDE.md §2, each with a reason.

**`qr/layout.go`** owns the module-to-pixel mapping. It lives in `qr/` because
CLAUDE.md §2 forbids `internal/render` from knowing anything QR-specific. The
consequence is load-bearing: **shape functions emit paths in absolute pixel
coordinates**, and no renderer ever learns what a module is.

**`internal/render/scene.go`** exists because the logo forces it. A `Path`
cannot carry a bitmap, yet SVG places the logo as an embedded `<image>` while
the rasteriser places it with `draw.Draw`. Without a shared scene description
each renderer would grow its own logo path — the duplication CLAUDE.md §2
forbids.

### 2.2 Scene

```go
type Scene struct {
    Width, Height int
    Background    color.Color // fully transparent means no background fill
    Items         []Item      // painted in order
}

type Item interface{ isItem() }

type PathItem struct {
    Path Path
    Fill color.Color
}

type ImageItem struct {
    Img        image.Image
    X, Y, W, H float64
    Clip       *Path // optional rounded-corner clip; nil means no clipping
}
```

`Clip` keeps the logo's corner radius renderer-agnostic: the rasteriser builds
an alpha mask from the path and uses `draw.DrawMask`; the SVG renderer emits a
`<clipPath>` and references it. Neither needs QR knowledge.

### 2.3 Path model

Fill-only. No stroke concept exists anywhere in the model.

```go
type Path struct{ SubPaths []SubPath }

type SubPath struct {
    Start  Point
    Segs   []Segment // Line or Quad
    Closed bool
}
```

Holes are made by winding direction: an outer subpath clockwise plus an inner
subpath counter-clockwise. Verified 2026-08-29 that
`golang.org/x/image/vector` applies `abs(acc)` clamped to 1
(`raster_floating.go:150`), which is identical to the non-zero winding rule for
integer windings and therefore identical to SVG's default `fill-rule`. A probe
confirmed it: an inner subpath drawn clockwise fills (alpha 255), drawn
counter-clockwise punches a hole (alpha 0).

This is what lets the finder-pattern rings — which the reference library draws
with `stroke` and `lineWidth = dotSize` — be expressed as ordinary fills that
both renderers reproduce with no renderer-specific code.

### 2.4 Data flow

```
Options
  | validate                        errors surface here, never at render time
  v
Encoder.Encode(content, ecc)        piglig, boostEcl pinned false
  v
Matrix                              modules plus finder classification
  v
Layout                              moduleSize, origin
  v
Logo                                occlusion rect -> excluded modules
  v
ShapeContext                        matrix + layout + exclusions + consumption
  |- dot shapes    -> []PathItem
  |- corner shapes -> []PathItem
  '- logo          -> PathItem (border, background) + ImageItem
  v
Scene --+-- raster.Render -> *image.RGBA -> PNG / JPEG
        '-- svg.Render    -> string
```

`*QR` stores the finished `Scene`. It is therefore immutable after `New`, and
`PNG`, `SVG` and `Image` are safe to call concurrently, as CLAUDE.md §6
requires.

## 3. Public API

Exactly as CLAUDE.md §3 specifies. Struct `Options` only — no fluent builder
and no functional options. The Java builder pattern that `kenglxn/QRGen` uses
is a workaround for language limits Go does not have; Go composite literals
with field names already give the same readability. The standard library agrees
(`http.Server`, `tls.Config`, `net.Dialer`, `jpeg.Options`). Practically:
errors surface at `New()` rather than at a deferred terminal call, `Options`
unmarshals from JSON or YAML which is what makes the phase-5 CLI easy, and one
surface means one compatibility promise after v1.0.0.

## 4. Geometry

Integer module size, symbol centred. Decided by measurement, not preference.

```
moduleSize = floor(Width / (modules + 2*Margin))
origin     = floor((Width - moduleSize*(modules + 2*Margin)) / 2)
```

Leftover pixels widen the quiet zone, which only helps scanning. `Width` is the
image size; the symbol occupies a slightly smaller centred square.

Control-guarded comparison across 4 symbols (v2, v5, v10, v20) and 8 widths
(120–1000 px):

| geometry | failures |
|---|---|
| integer module | 0 / 32 |
| exact `Width`, fractional module | 3 / 32 |

Every fractional failure was the dense v20 symbol at 1.24, 2.06 and 2.64
px/module, where anti-aliased edges blur neighbouring modules together.

Also verified: `x/image/vector` output is pixel-identical (`diffPx = 0`) to a
nearest-neighbour pixel fill at 4, 8, 11 and 22 px/module. The rasteriser
introduces no error of its own.

`New` returns `ErrWidthTooSmall` when `Width < modules + 2*Margin`, because the
module size would round to zero.

**Margin is measured in modules, default 4.** This deliberately diverges from
the reference library, which treats `margin` as pixels and therefore ships an
effective quiet zone of roughly 0.65 modules at its default width of 380 — far
below what ISO/IEC 18004 requires. The README must state the divergence.

## 5. Encoder

`github.com/piglig/go-qr` v1.1.0, MIT, a Nayuki port. It exposes `Size() int`
and `Module(x, y) bool`, and after `go mod tidy` a consumer's `go.mod` gains
exactly one line with no transitive runtime dependencies.

**Do not call `EncodeText`.** It delegates to `EncodeStandardSegments`, which
loops over Medium, Quartile and High to upgrade the level whenever the data
still fits (`encode.go:63`). That would make the actual ECC level differ from
the requested one, breaking golden-test determinism and the logo occlusion
budget. Pin it:

```go
segs, err := goqr.MakeSegments(content)
q, err := goqr.EncodeSegments(segs, ecl, 1, 40, -1, false) // boostEcl = false
```

The package exposes no version accessor; derive it as `(q.Size() - 17) / 4`.

The `Encoder` interface stays as CLAUDE.md §2 defines it so the phase-6 swap to
a built-in encoder is non-breaking.

## 6. Module classification

`matrix.go` classifies every module as finder, separator, timing, alignment,
format, version or data.

Only "is this module inside one of the three 7x7 finder patterns" is
load-bearing today: it drives the exclusion of finder modules from dot
rendering. The remaining categories are cheap, independently testable against
known QR versions, and useful for debugging, but they must not grow features of
their own. Alignment patterns render as ordinary dots, and timing and format
modules are data, both per CLAUDE.md §4.

The reference library has no classification at all — it hardcodes two 7x7 masks
at the three corners. Ours is cleaner and equivalent for the cases that matter.

## 7. Shape model

### 7.1 ShapeContext

CLAUDE.md §4 specifies that shape functions receive a "neighbour mask". That is
wrong, and the correction is load-bearing.

Reading `src/core/QRDot.ts` shows two different mechanisms. `fluid` and
`fluid-line` need a local 4- or 8-neighbour mask. But `stripe`, `stripe-row`
and `stripe-column` perform **greedy run consumption with mutable state**: they
try runs in order and mark every module of a matched run as consumed so the
main loop skips it. A read-only 3x3 snapshot cannot express that — those shapes
must read arbitrary coordinates and write consumption state back.

```go
// ShapeContext gives a shape function everything it needs to decide its
// geometry: the module grid, the pixel layout, and the consumption state that
// run-merging shapes such as stripe require.
type ShapeContext interface {
    // Dark reports whether (x, y) is a dark module this shape may claim. It is
    // false when the coordinate is out of bounds, light, already consumed,
    // part of a finder pattern, or covered by the logo.
    Dark(x, y int) bool

    // Consume marks (x, y) as drawn so the main loop skips it. A shape that
    // merges several modules into one figure calls this for every module it
    // claims.
    Consume(x, y int)

    // Rect returns the pixel rectangle of module (x, y).
    Rect(x, y int) (px, py, size float64)

    Size() int // modules per side
}

func neighbours4(c ShapeContext, x, y int) (n, e, s, w bool)
func neighbours8(c ShapeContext, x, y int) mask8
```

The masks are helpers built on the context, never the signature itself.

`Dark` folds five conditions into one question — out of bounds, light, already
consumed, finder, under the logo. This makes the reference library's second
defect structurally impossible: there, `getRangeTrue` checked only darkness, so
a `stripe` run could grow straight through a finder pattern or the logo safe
zone. With a single accessor, no shape can ask "is it dark?" without also
asking "may I use it?".

### 7.2 Dot catalogue

Every shape reduces to two primitives: a rounded rectangle with per-corner
radii — optionally rotated — and a quadratic curve.

Dimensions below are exact, because "a rounded rect with radius 0.4·size" is
ambiguous: it could mean a full-size cell with rounded corners or a smaller
centred square whose radius makes it a circle, and those are different shapes.
Each entry states the side length and the radius, both relative to the module
size `s`, with every figure centred in its module unless stated otherwise.

| Type | Construction | Neighbour input |
|---|---|---|
| `square` | side `s`, r = 0 | none |
| `tile` | side `s − 1px`, anchored at the module's top-left | none |
| `dot` | side `0.8s`, r = `0.4s` (half its own side, so a circle) | none |
| `dot-small` | side `0.6s`, r = `0.3s` (a circle) | none |
| `rounded` | side `0.75s`, r = `0.1875s` (a quarter of its own side) | none |
| `diamond` | side `0.5s / sin 45° ≈ 0.7071s`, r = 0, rotated 45° | none |
| `star` | four quadratic curves from the cell corners with the control point at the centre, rotated 45° | none |
| `fluid` | side `s`; per-corner r = `0.5s` where both adjacent neighbours are light, otherwise 0 | 4-way |
| `fluid-line` | `fluid` plus concave fillets joining lower diagonal neighbours | 8-way |
| `stripe` | greedy runs 3x1, 1x3, 2x1, 1x2, 1x1 → bar of thickness `0.5s`, r = `0.25s` | consumption |
| `stripe-row` | greedy runs 3x1, 2x1, 1x1 | consumption |
| `stripe-column` | greedy runs 1x3, 1x2, 1x1 | consumption |

The important simplification is `fluid`. The reference draws a circle and then
patches four quadrants with `fillRect`, which cannot be one path. The result is
mathematically identical to a rounded rectangle whose per-corner radius is
either `size/2` or `0` — a single subpath with no patching.

### 7.3 Corner catalogue

Rings are built by reversed winding, never by stroking. Translating the
reference's stroked figures into fills is the step most likely to go quietly
wrong, so the arithmetic is written out here.

The reference strokes a centreline figure with `lineWidth = dotSize`. Offsetting
a rounded rectangle outward by `d` grows its side by `2d` and its corner radius
by `d`. With `s` the module size and a stroke half-width of `0.5s`, a stroked
centreline figure of side `L` and radius `R` becomes two filled subpaths:

```
outer subpath (clockwise):         side L + s,  radius R + 0.5s
inner subpath (counter-clockwise): side L - s,  radius max(0, R - 0.5s)
```

For the 7x7 finder ring the centreline side is `6s`, so the outer subpath is
exactly the `7s` box and the inner subpath is the `5s` box — the ring is one
module thick with its outer edge flush to the finder pattern, as required.

The inner 3x3 core is stroked *and* filled in the reference, so its union is
simply a solid figure of side `3s` with radius `Radius.Inner + 0.5s`.

The `R ± 0.5s` rule applies only where `R > 0`. When `R = 0` the centreline has
a real corner join, and the canvas default `lineJoin` is `miter`, so the stroked
outline stays sharp — the offset does **not** introduce a radius. Where `R > 0`
the corner curve meets the straight segments tangentially, there is no join, and
the rule holds.

| Type | Outer 7x7 ring (two subpaths) | Inner 3x3 core (solid) |
|---|---|---|
| `square` | side `7s` / `5s`, r = 0 (sharp) | side `3s`, r = 0 |
| `rounded` | r = `Radius.Outer ± 0.5s`, default `Radius.Outer = 0.5s` → `1.0s` / `0` | side `3s`, r = `Radius.Inner + 0.5s`, default `Radius.Inner = 0.25s` → `0.75s` |
| `circle` | circles r = `3.5s` and `2.5s` | disc r = `1.5s` |
| `rounded-circle` | rounded rect as `rounded` | disc r = `1.5s` |
| `circle-rounded` | circles r = `3.5s` and `2.5s` | rounded rect side `3s`, r = `Radius.Inner + 0.5s` |
| `circle-diamond` | circles r = `3.5s` and `2.5s` | square side `3s` rotated 45°, r = 0 |
| `circle-star` | circles r = `3.5s` and `2.5s` | four quadratic curves spanning `3s` |

A golden-image test must pin each of these, because an error of half a module
in a ring is visible but easy to mistake for an intended style.

## 8. Logo

### 8.1 Source and validation

Exactly one of `Image`, `Path` or `Reader` must be set, otherwise
`ErrLogoSource`. Decoding supports PNG and JPEG through the standard library.
Resampling uses `draw.CatmullRom` from `golang.org/x/image`.

### 8.2 Sizing

`Size` is a fraction of `Width` and describes the **outer** block including the
border, because that is the dimension a user reasons about. `Size == 0` means
auto, using the reference formula:

```
coverLevel    = ECC percent (L .07, M .15, Q .25, H .30)
maxHiddenDots = floor(coverLevel^2 * modules^2)
```

The squared factor is a deliberate safety margin — at High it yields 9% of
modules, not 30% — because the ECC percentage describes codeword recovery, not
a safe occlusion area. It is what the reference ships and what our decode tests
will validate.

### 8.3 Safe zone and budget

Modules whose pixel rect intersects the logo's outer rect are excluded from
drawing, through the same `ShapeContext.Dark` accessor that excludes finder
modules. The count of excluded modules is compared against `maxHiddenDots`; if
it exceeds the budget, `New` returns `ErrLogoTooLarge` rather than emitting an
unscannable code, per CLAUDE.md §5.1.

### 8.4 Paint order

Background, dots, corners, logo border and background, logo image. The logo is
painted last so it covers whatever lies beneath.

## 9. ECC policy

- ECC set explicitly: honour it, then validate the logo against its budget.
- Auto without a logo: the reference heuristic by content length — over 36
  characters gives M, over 16 gives Q, otherwise H. Short content gets maximum
  protection because there is room for it.
- Auto with a logo: High, per CLAUDE.md §5.1, because the logo needs the
  occlusion budget.

Forcing High has a real cost worth stating: 259 characters encode to v17 at
High but v15 at Quartile, and a larger symbol at a fixed pixel width means
smaller modules and therefore harder scanning. That is why High is forced only
when a logo is present, never unconditionally.

## 10. Colour and contrast

Hex with or without `#`; a bare `ff0000` becomes `#ff0000`. Non-hex strings
pass through untouched to SVG. The raster renderer parses what it can and
returns `ErrBadColor` otherwise. `#00000000` means a transparent background.

`ErrLowContrast` (CLAUDE.md §5.5) uses a threshold **derived by measurement,
not chosen by intuition**: sweep foreground/background pairs from high to low
contrast ratio, find where decoding starts to fail, and set the threshold above
that point with margin. A transparent background is evaluated against white,
since that is what most often sits behind it.

## 11. Error handling

All validation happens in `New`. Render methods return only I/O errors from the
`io.Writer`. No panics anywhere in library code.

```go
var (
    ErrNoContent      = errors.New("qr: content is empty")
    ErrWidthTooSmall  = errors.New("qr: width too small for the module count")
    ErrBadColor       = errors.New("qr: cannot parse colour")
    ErrLowContrast    = errors.New("qr: foreground/background contrast too low to scan")
    ErrContentTooLong = errors.New("qr: content does not fit at the requested ECC level")
    ErrLogoSource     = errors.New("qr: logo needs exactly one of Image, Path or Reader")
    ErrLogoTooLarge   = errors.New("qr: logo occludes more than the error-correction budget allows")
)
```

Wrapped with `%w` plus context so `errors.Is` works and messages stay useful.

## 12. Testing

### 12.1 The decoder baseline problem

Measured 2026-08-29 over 344 content and ECC combinations per encoder, on
plain black-on-white integer-aligned renders with a 4-module quiet zone:

| encoder | gozxing failures |
|---|---|
| `piglig/go-qr` | 10 / 344 (2.9%) |
| `rsc.io/qr` | 6 / 344 (1.7%) |
| `boombuler/barcode` | 7 / 344 (2.0%) |

Three independent encoders failing on overlapping inputs — `L/len159/v8` in two
of them, `Q/len600/v23` in all three — means the common factor is gozxing, not
any encoder. Failures are deterministic and surface as
`gozxing.formatException`: the symbol is located but the data will not parse.
`piglig`'s own decoder reads every one of them correctly.

A related dead end, recorded so nobody repeats it: for each failing case the
mask gozxing rejects is exactly the mask the encoder auto-selected. That looks
like an encoder bug and is not — all three encoders implement the same spec
penalty rules and therefore choose the same mask. It is one observation, not
two. Do not switch encoders over it.

Not established: whether real-world scanners share the weakness. Do not claim
they do.

### 12.2 The control render

Left unhandled, a 2–3% baseline false-failure rate would make the matrix
CLAUDE.md §6 mandates flaky for reasons unrelated to our renderer, which
destroys trust in exactly the tests meant to be the safety net.

Every round-trip test therefore starts with a control:

```go
// requireDecodableBaseline renders content unstyled, integer-aligned, at 10px
// per module and decodes it. gozxing cannot read roughly 2-3% of valid QR
// symbols; when the plain render fails, a styled render proves nothing, so the
// case is skipped rather than reported as our bug.
func requireDecodableBaseline(t *testing.T, content string, ecc ECCLevel)
```

Only when the control passes does a styled failure indicate a real bug.

### 12.3 Matrix

| Test | Coverage |
|---|---|
| Shape matrix | 12 dots x 7 corners = 84, round-trip decode |
| Logo matrix | 12 dots with logo, 7 corners with logo, round-trip decode |
| Logo sizing | several sizes including one deliberately oversized that must be rejected |
| ECC | all four levels, short content (v1–2) and long (v25+) |
| Transparency | `#00000000` flattened onto white, then decoded |
| SVG parity | see below |
| Golden | `testdata/golden/` with a `-update` flag |
| Classification | `matrix.go` against known QR versions |
| Concurrency | one `*QR`, rendered from many goroutines, outputs compared |
| Fuzz | `New` on arbitrary content, must never panic |
| Benchmarks | default path, `fluid`, `stripe` |

All tests are table-driven. `go test ./... -race` must pass.

### 12.4 SVG parity without a dependency

Go has no SVG rasteriser in the standard library, and adding a third-party one
would introduce a test dependency whose own bugs could be mistaken for ours.

Instead the parity test parses the `d` attributes of the emitted SVG back into
`render.Path` values and compares them against the `Path` values in the
`Scene`. That proves the SVG serialiser is lossless with respect to the shared
path model. `viewBox`, `width`, `height` and fill colours are asserted
directly. The rasteriser's faithfulness to the same `Scene` is already proven
by the decode tests, so "PNG and SVG come from the same geometry" follows.

The honest limitation: this does not catch a fault that only appears when the
SVG is rendered by some other engine. A test-only rasteriser can be added later
if that risk ever materialises.

## 13. Build order

One branch, `feat/styled-qr-core`, committed in logical chunks with the
conventional-commit prefixes and co-author trailer from CLAUDE.md §9.

```
 1  chore  go.mod, LICENSE, README stub, CI workflow
 2  feat   Encoder interface + piglig backing, boostEcl pinned false
 3  feat   matrix.go and classification
 4  feat   layout.go module -> pixel
 5  feat   render.Path and render/color
 6  feat   render.Scene and the raster renderer
 7  feat   square dot, square corner, Options/New/PNG
        == GATE: round-trip decode green, or stop and report ==
 8  feat   SVG renderer and parity tests
 9  feat   stripe, stripe-row, stripe-column
        == GATE: if ShapeContext must change, stop and report ==
10  feat   fluid and fluid-line
11  feat   the remaining seven dot types
12  feat   the seven corner types, radii, independent colours
13  feat   logo: decode, resize, border, radii, safe zone, ECC budget
14  test   full matrices, golden images, fuzz, benchmarks
15  docs   README with the rendered gallery
```

The stripe shapes are step 9 rather than step 11 deliberately. They stress
`ShapeContext` harder than anything else, and if the signature is wrong that
must surface before eleven other shapes are built on top of it.

Do not push until tests are green; then push and open a PR into `main`.

## 14. Decisions carried in from research

| Decision | Evidence |
|---|---|
| Encoder is `piglig/go-qr`, `boostEcl` pinned false | source read plus round-trip run |
| Fill-only paths, holes by reversed winding | `raster_floating.go:150` plus alpha probe |
| Integer module geometry | 0/32 versus 3/32 control-guarded failures |
| Struct `Options` only | Go stdlib precedent plus error-timing and marshalling |
| `ShapeContext`, not a neighbour mask | `QRDot.ts` run-consumption logic |
| Control render in every decode test | 2–3% gozxing baseline across three encoders |
| Margin in modules, not pixels | reference ships a 0.65-module quiet zone |

## 15. Consumability

Fariz's requirement: this must be easy to drop into other Go projects. Treated
here as testable criteria rather than good intentions, because "easy to use" is
only real if something fails when it stops being true.

| Criterion | How it is enforced |
|---|---|
| Import is one line and reads naturally | module `github.com/farizfadian/go-qrcode`, package `qr` under `qr/`, so consumers write `qr.New(...)` |
| Simplest useful call is two lines | `q, err := qr.New(qr.Options{Content: "..."})` then `q.WritePNGFile("out.png")` — a test asserts this exact snippet compiles and decodes |
| Zero-value `Options` is valid | only `Content` is required; a table test drives every field left at its zero value |
| Dependency footprint stays tiny | a test asserts `go list -deps` adds only `piglig/go-qr` and `golang.org/x/image`; adding a third needs an ADR under `docs/adr/` |
| No hidden setup | no `init()`, no package-level mutable state, no registration step; a `*QR` needs nothing but `New` |
| Safe to share | `*QR` is immutable after `New`, so concurrent renders are safe — proven by a `-race` test |
| Errors are actionable | sentinel errors comparable with `errors.Is`, wrapped with context; never a panic |
| Discoverable from godoc alone | runnable `Example` functions for the default code, each dot type, each corner type, a logo, and SVG output |

The runnable examples matter more than they look. They appear in godoc, they
are compiled by `go test`, and they are the first thing someone evaluating the
library reads. An example that stops compiling is a broken promise the test
suite catches immediately.

## 16. Deferred

The `soldair/node-qrcode` surface — terminal and UTF-8 output, manual segment
and mode control, explicit version and mask pinning — is not decided. It does
not block anything here: the phase 1–4 design is identical with or without it.
Revisit when the `Options` surface is being finalised for v1.0.0.
