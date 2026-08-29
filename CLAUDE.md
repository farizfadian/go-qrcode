# CLAUDE.md

Guidance for Claude Code working in this repository.

---

## 1. What this project is

`go-qrcode` is a pure-Go library for generating **styled QR codes with a
centred logo**.

- Module path: `github.com/farizfadian/go-qrcode`
- License: MIT
- Minimum Go: 1.22 (develop on latest stable)

The QR package is a Go port of the feature surface of the TypeScript library
[`zxpsuper/qrcode-with-logos`](https://github.com/zxpsuper/qrcode-with-logos)
(docs: <https://zxpsuper.github.io/qrcode-with-logos/guide/options.html>).
That library is the **behavioural specification** for `qr/`. It is MIT licensed,
so its ideas may be ported freely — but write original Go code, do not
transliterate TypeScript line by line.

### Why this exists

Go already has QR-with-logo libraries. They were evaluated and found insufficient:

| Library | What it has | What it lacks |
|---|---|---|
| `yeqown/go-qrcode/v2` | logo, gradient, transparent bg, circle cells, halftone, wasm | only 2 cell shapes, **no finder-pattern styling**, **no SVG output** |
| `piglig/go-qr` | pure Go, PNG + SVG, centred logo with ECC validation | no styling at all |
| `boombuler/barcode` | 11 symbologies, mature, colour schemes | no styling, **no SVG**, **no human-readable text rendering** |

**The gap this library fills:** per-module dot shapes, independently styled and
coloured finder patterns ("corners"), a decorated logo (border, radius,
background), and **identical output from both a raster and an SVG renderer**.

If a task can already be done by `yeqown/go-qrcode/v2` with three lines of code,
this library is not adding value — flag that in review rather than building it.

### Scope and sequencing — read this before creating any package

The repository is named `go-qrcode` and QR is its primary subject. It may later
grow additional symbologies — DataMatrix and Aztec (where dot styling genuinely
applies), and possibly 1D codes like Code 128 and EAN (where it does not — bar
widths are dictated by spec and cannot be restyled without breaking
scannability).

**But not yet.** Phases 1–6 are QR only.

> Do not create `code128/`, `ean/`, `datamatrix/`, or any other symbology
> package until `qr/` reaches v1.0.0. The renderer must be proven against one
> symbology before it is generalised. Building the folder structure now means
> designing an abstraction for requirements nobody understands yet.

The `qr/` subpackage layout in §2 exists to keep that door open at zero cost —
not as an invitation to walk through it now.

When 1D symbologies do arrive, **do not reimplement the encoders** — use
`boombuler/barcode` for the bar patterns and add the rendering layer on top.
The value-add there is SVG output and human-readable text, not encoding.

### Non-goals

- ~~Barcode **decoding**~~ — **reversed 2026-08-29.** QR *reading* now ships as
  `qr.Scan`. The reason the non-goal existed was dependency cost, and that
  turned out not to apply: `piglig/go-qr`, already the encoder, exposes
  `DecodeDetailed`, so reading came at zero new dependencies and matched
  gozxing on every input tested. Decoding *other symbologies* remains a
  non-goal
- Micro QR / rMQR
- Browser/WASM build in v1 (revisit after v1.0.0)
- Animated or video codes

---

## 2. Architecture

The single most important design decision:

> **Shapes are described as vector paths once, then handed to a renderer.**
> The SVG renderer serialises those paths to `d` attributes. The raster renderer
> rasterises the *same* paths. This guarantees PNG and SVG output match, and
> means adding a new dot shape requires touching exactly one file.

Do not write shape logic twice (once for canvas-style drawing, once for SVG).
If you find yourself doing that, stop and refactor to the shared path model.

```
go-qrcode/
├── go.mod
├── CLAUDE.md
├── README.md
├── LICENSE                     MIT
├── barcode.go                  shared types (keep minimal until a 2nd symbology exists)
│
├── qr/                         ← all Phase 1–6 work happens here
│   ├── qr.go                   public API: New, Options, QR, PNG/SVG/Image
│   ├── options.go              option structs, defaults, validation
│   ├── encode.go               text -> module matrix (Encoder interface + impl)
│   ├── matrix.go               module classification: finder / separator / timing /
│   │                           alignment / format / version / data
│   ├── layout.go               module -> pixel mapping (added: see note below)
│   ├── shape.go                ShapeContext, DotType/CornerType, shape registries
│   ├── shape_dots.go           12 dot shapes -> render.Path
│   ├── shape_corners.go        7 corner shapes -> render.Path
│   ├── parse.go                name parsing driven by the registries (added)
│   └── logo.go                 decode, resize, border, rounded corners, bg, safe zone
│
├── internal/render/            ← shared by every future symbology
│   ├── path.go                 Path, SubPath, MoveTo/LineTo/CurveTo/Close, bounds
│   ├── scene.go                Scene: what both renderers consume (added)
│   ├── raster.go               Scene -> *image.RGBA (PNG / JPEG)
│   ├── svg.go                  Scene -> SVG string
│   └── color.go                hex parsing ("ff0000" -> #ff0000), alpha, named passthrough
│
├── cmd/qrgen/                  CLI
├── docs/images/                README gallery, every image decode-checked in CI
└── testdata/golden/
```

Three files were added to this layout during implementation:

- **`qr/layout.go`** owns the module-to-pixel mapping. It lives in `qr/`, not in
  `render/`, because nothing in `internal/render` may know what a module is. The
  consequence is load-bearing: **shape functions emit paths in absolute pixel
  coordinates.**
- **`internal/render/scene.go`** exists because the logo forces it. A `Path`
  cannot carry a bitmap, yet SVG embeds the logo as `<image>` while the
  rasteriser uses `draw.Draw`. Without a shared `Scene`, each renderer grows its
  own logo path — the duplication §2 forbids.
- **`qr/parse.go`** derives shape names from the registries rather than from a
  second list, so the CLI and any config-driven caller stay correct as shapes
  are added without being edited.

### Why `internal/render/` and not `render/`

Keeping it internal means the API is not yet a public promise. Promoting an
internal package to public later is non-breaking; making a public package
private later is not. It stays internal until at least two symbologies use it
and its shape has settled.

**Critical constraint:** nothing in `internal/render/` may import `qr/` or know
anything QR-specific. It deals in paths, colours, and pixels. If a render
function needs to know about finder patterns, the abstraction is wrong.

### Dependency policy

Keep the runtime dependency surface tiny — consumers should pull in almost nothing.

- **Allowed at runtime:** `golang.org/x/image` (for the `vector` rasteriser and
  `draw.CatmullRom` logo resampling). Standard library for everything else.
- **Test-only:** `github.com/makiuchi-d/gozxing` (QR reader, see §6). Test
  dependencies are pruned from consumer builds under Go module graph pruning,
  so this is acceptable.
- **Adding any other runtime dependency requires an explicit decision recorded
  in `docs/adr/`.** Do not add one silently.

### The encoder question

Generating the module matrix (data encoding, ECC, masking, format/version info)
is roughly 1,200–1,800 lines of well-understood work.

**Phase 1 approach:** define an `Encoder` interface in `qr/encode.go` and back
it with an existing pure-Go encoder so styling work can start immediately.

```go
type Encoder interface {
    // Encode returns a square matrix; true = dark module.
    Encode(content string, ecc ECCLevel) ([][]bool, error)
}
```

Before wiring any third-party encoder, **verify it actually exposes the raw
module grid** (many only expose a finished image). Candidates in order of
preference: `github.com/piglig/go-qr` (Nayuki port), `rsc.io/qr`,
`github.com/boombuler/barcode/qr`. Confirm the API by reading the actual
package source, not from memory.

**Phase 6 approach:** ship a built-in encoder so the only runtime dep is
`x/image`. Nayuki's QR-Code-generator is MIT and is the reference to work from.
Keep the `Encoder` interface either way so this swap is non-breaking.

---

## 3. Public API

Aim for idiomatic Go, not a transliteration of the JS constructor. The JS
library returns a mutable instance; Go should return an immutable value plus
explicit render calls.

```go
package qr // github.com/farizfadian/go-qrcode/qr

type Options struct {
    Content string        // required
    Width   int           // default 380
    Margin  int           // quiet zone in modules, default 4
    ECC     ECCLevel      // default: auto (see §5.1)

    Foreground string     // default "#000000"
    Background string     // default "#ffffff"; "#00000000" = transparent

    Dots    DotOptions
    Corners CornerOptions
    Logo    *LogoOptions  // nil = no logo
}

type DotOptions struct {
    Type  DotType  // default DotSquare
    Color string   // default "" -> inherit Options.Foreground
}

type CornerOptions struct {
    Type   CornerType    // default CornerSquare
    Color  string        // default "" -> inherit Options.Foreground
    Radius CornerRadius  // used when Type includes rounding
}

type CornerRadius struct {
    Inner float64 // default dotSize/4
    Outer float64 // default dotSize/2
}

type LogoOptions struct {
    Image  image.Image // one of Image / Path / Reader must be set
    Path   string
    Reader io.Reader

    Size         float64 // fraction of QR width; 0 = auto
    Radius       float64 // logo image corner radius, default 0
    BorderWidth  float64 // default 10
    BorderRadius float64 // default 8
    BorderColor  string  // default = Background
    BgColor      string  // default "#ffffff"
}

func New(opts Options) (*QR, error)

func (q *QR) PNG(w io.Writer) error
func (q *QR) JPEG(w io.Writer, quality int) error
func (q *QR) SVG(w io.Writer) error
func (q *QR) SVGString() (string, error)
func (q *QR) Image() image.Image
func (q *QR) WritePNGFile(path string) error
```

### API rules

- `New` validates and returns an error. Rendering methods should not fail on
  bad configuration — that is `New`'s job.
- **Never panic in library code.** Return errors. Panics are acceptable only in
  tests and in `cmd/`.
- Zero-value `Options` (other than `Content`) must produce a sane default QR
  code. Every default above is normative.
- Colours accept hex with or without `#`. A bare `ff0000` becomes `#ff0000`.
  Non-hex strings (`rgb(...)`, `red`) pass through untouched to SVG; for the
  raster renderer, parse what you can and return a clear error otherwise.
- Exported identifiers get doc comments, starting with the identifier name.

---

## 4. Feature parity checklist

Port these from the reference library. Tick as implemented.

### Dot types (12)
`square` (default) · `dot` · `dot-small` · `tile` · `rounded` · `diamond` ·
`star` · `fluid` · `fluid-line` · `stripe` · `stripe-row` · `stripe-column`

> **`fluid`, `fluid-line`, `stripe*` are neighbour-aware**, and a neighbour mask
> is **not** enough. Corrected during implementation; the original wording said
> "mask" and it was wrong twice over.
>
> `stripe*` merge runs of modules into one figure and must *write* consumption
> state back so the main loop skips what they swallowed. A read-only snapshot
> cannot express that. The signature is therefore a context:
> `func(c ShapeContext, x, y int) render.Path`.
>
> `ShapeContext` needs **two** accessors, not one. `Dark(x, y)` means "dark and
> still claimable" — it is false for a consumed, finder or logo-covered module,
> and run merging uses it. `Adjacent(x, y)` means "dark, for deciding how edges
> meet" — it ignores consumption, because the main loop works row by row and a
> module's northern and western neighbours are always already drawn by the time
> it is reached. `fluid` asking `Dark` there reports every module as isolated.
>
> Build `stripe*` **first** among the shapes: it stresses run merging hardest.
> But note it does not exercise adjacency at all, so it cannot prove the whole
> design — `fluid` is the second gate.

### Corner types (7)
`square` (default) · `rounded` · `circle` · `rounded-circle` ·
`circle-rounded` · `circle-star` · `circle-diamond`

Corners are the three 7×7 finder patterns. They must be excluded from normal
dot rendering — `matrix.go` classification drives this. Alignment patterns
(the smaller 5×5 blocks) render as **normal dots**, matching the reference.

### Logo
border width · border radius · logo corner radius · background colour ·
auto-sizing · safe-zone clearing of modules underneath · optional vector
version embedded into SVG output

> **Do not add an SVG rasteriser to support vector logos in PNG.** It was
> evaluated: `oksvg` costs four untagged dependencies including
> `golang.org/x/net`, and was measured silently dropping `rx` on `<rect>`,
> turning rounded corners square with no error at all. A logo that comes out
> subtly wrong with nothing failing is worse than one that is not supported.
> `LogoOptions.SVGMarkup` nests the markup into SVG output instead, where the
> displaying engine renders it correctly, and the README points users at their
> design tool for the raster formats.

### Renderers
raster (PNG, JPEG) · SVG string · transparent background (`#00000000`)

### Deliberately dropped
`canvas` / `image` DOM elements, `crossOrigin`, `download` / `downloadName`,
`onError` — all browser-specific. Go's equivalent is writing to an `io.Writer`.

---

## 5. Correctness rules that are easy to get wrong

1. **ECC and logo interact.** A logo occludes data. When `Logo != nil` and the
   user did not set `ECC` explicitly, force `ECCHigh`. If the logo (including
   its border) would occlude more than the correction budget can recover,
   **return an error from `New`** — do not silently emit an unscannable code.
2. **Auto logo size.** Reference behaviour is auto-calculation. Target roughly
   1/5 of QR width, and validate total occluded area against the ECC level
   rather than trusting a fixed ratio.
3. **Never style the quiet zone.** Margin must stay pure background.
4. **Timing patterns and format info are data.** They render as dots. Do not
   special-case them into corner styling.
5. **Low-contrast configurations break scanners.** If foreground and background
   luminance contrast is poor, return an error or a documented warning — do not
   just draw it.
6. **Anti-aliasing at small widths.** Below ~200px, aggressive shapes (`star`,
   `dot-small`) degrade scannability. Document this in the README; consider a
   minimum-width warning.
7. **Inverted polarity does not scan, and contrast will not catch it.** A light
   foreground on a dark background has excellent contrast and is still
   unreadable: the specification assumes dark modules on a light field, so
   reversing the polarity hides the finder patterns from the detector.
   Measured, not assumed. This is a separate check from rule 5 and both are
   still to be implemented.
8. **A styled finder pattern must preserve the 1:1:3:1:1 ratio.** A reader
   locates a symbol by scanning for dark:light:dark:light:dark in that ratio
   through the finder's centre. Any core figure that overfills the ring's
   five-module gap destroys it. The reference library's `circle-diamond` does
   exactly that — it rotates a full three-module square, giving a 4.24-module
   diagonal and a measured ratio of 1:0.38:4.24:0.38:1, which no reader
   recognises. Inscribe such figures in the 3×3 core instead.

---

## 6. Testing — non-negotiable

A styled QR code that looks beautiful and does not scan is a bug, not a style.

**Every rendering test must round-trip through a real decoder.** Use
`github.com/makiuchi-d/gozxing` to decode the generated image and assert the
content matches the input.

**And every such test needs an unstyled control first.** Measured over 344
content/ECC combinations: gozxing fails to decode roughly 2–3% of *valid*
symbols. Three independent encoders (`piglig/go-qr`, `rsc.io/qr`,
`boombuler/barcode`) fail on overlapping inputs, so the fault is the decoder,
not any encoder — do not switch encoders over it. Without a control, the matrix
below goes flaky for reasons unrelated to the renderer, and a flaky safety net
stops being one. Render the same content unstyled at a generous module size,
decode that, and `t.Skip` when it fails.

Required test matrix:

- Every dot type × every corner type — decode round-trip
- Every dot type × logo present — decode round-trip
- Short content (v1) and long content (v25+) — decode round-trip
- Transparent background — flatten onto white, then decode
- SVG output — **SVG and PNG output must decode to the same content and have
  matching module geometry.** Implemented without rasterising the SVG: the test
  parses the emitted `d` attributes back into `render.Path` values and compares
  them with the paths the `Scene` holds. Go has no SVG rasteriser in the
  standard library, and a third-party one's bugs would be indistinguishable from
  ours. The honest limit: this cannot catch a fault that appears only when some
  other engine renders the SVG.
- All four ECC levels
- Golden-image tests in `qr/testdata/golden/` with a `-update` flag to
  regenerate. Note the path: Go resolves `testdata` relative to the package, so
  they live beside `qr/`, not at the repository root. They compare **decoded
  pixels**, not file bytes — a change in Go's PNG encoder would otherwise break
  every one without anything in this library having changed.

  They also compare **with a tolerance**, and that is not laziness. Exact pixel
  equality is not portable: `x/image/vector` ships an amd64 assembly rasteriser
  and a pure-Go fallback, so an anti-aliased curve rounds one level differently
  on arm64. Measured at 8 of 67,600 pixels on the circular finder shapes;
  straight edges are unaffected. A real geometry change moves whole modules —
  thousands of pixels differing by hundreds of levels — so the tolerance sits
  in a wide gap between the two regimes. Both ends of it are pinned by tests.

Also required:

- Table-driven tests throughout
- Fuzz test on `New` for arbitrary content strings
- Benchmarks for the default path and the most expensive shapes. These are not
  decoration: they caught a quadratic `Path.Append` in `New` that cost 1.36
  seconds and 1.2 GB on a version 30 symbol while every one of 293 tests passed.
  Correct and unusably slow is still a bug.
- `go test -race ./...` must pass; a `*QR` is immutable after `New` — keep it
  that way so it is safe for concurrent use

---

## 7. Commands

```bash
go build ./...
go test ./... -race                      # needs a C toolchain; see note below
go test ./... -run TestGolden -update    # regenerate golden images
go vet ./...
gofmt -l .                               # must output nothing
golangci-lint run                        # if installed
go test -bench=. -benchmem ./...
```

CI (`.github/workflows/ci.yml`) runs build, vet, a gofmt check, a `go mod tidy`
check, and tests with the race detector across Go 1.22 (the declared floor),
1.24 and 1.25, on Linux, Windows and macOS. It also asserts the runtime
dependency footprint and decodes every image the README displays.

**The race detector needs cgo**, so it cannot run on a Windows machine without a
C toolchain. CI is the only place it runs. Do not treat a green local
`go test ./...` as evidence of race-freedom.

---

## 8. Roadmap

Work in this order. Do not start a phase before the previous one's tests pass.

- **Phase 1 — skeleton.** `go.mod`, `Encoder` interface + third-party backing,
  `qr/matrix.go` classification, `internal/render/path.go`, raster renderer,
  `square` dots, `square` corners, PNG output. Round-trip decode test green.
- **Phase 2 — SVG renderer.** Same paths, second output. Assert PNG/SVG parity.
- **Phase 3 — shapes.** All 12 dot types, all 7 corner types, independent
  colours, corner radii. Neighbour-aware shapes included.
- **Phase 4 — logo.** Decode, resize, border, radii, background, safe zone,
  ECC validation.
- **Phase 5 — polish.** CLI, examples, README with a rendered gallery,
  benchmarks, fuzzing, godoc pass.
- **Phase 6 — v1.0.0.** Tag and release. Optionally replace the third-party
  encoder with a built-in one (non-breaking behind the `Encoder` interface).

**Actual order taken.** Phases 1–3 ran as written. The CLI, README gallery and
release workflow (Phase 5) were pulled forward ahead of the logo, at Fariz's
request, so the repository had a usable front door earlier. That cost nothing
because `cmd/qrgen` reads its shape list from the registry: every shape added
afterwards appeared in `qrgen -h` with no edit to `cmd/`.

**No release until the library is feature-complete.** Fariz's decision, and it
matches this roadmap: v1.0.0 belongs at Phase 6. Go treats v1.0.0 as an API
stability promise — after it, any breaking change needs a `/v2` module path and
every consumer must edit their imports. `ErrLogoUnsupported` is due to be
removed when the logo lands, so tagging v1.0.0 now would mean either breaking
semver later or carrying a dead error forever.

Only after v1.0.0: consider DataMatrix and Aztec (styling applies), then 1D
symbologies via `boombuler/barcode` (styling does not apply — the value is SVG
output and human-readable text). See §1.

---

## 9. Repository conventions

- **Git account:** personal — `farizfadian`. This is a personal open-source
  project, so Claude may commit and push here.
- **Commits:** conventional commits (`feat:`, `fix:`, `docs:`, `test:`,
  `refactor:`, `chore:`).
- **Co-author trailer on every Claude commit:**
  `Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`
  (use the actual model version, not a bare `Claude`).
- **Branching:** feature branches per phase, PR into `main`.
- **README** must include a rendered gallery of every dot and corner type — it
  is the main selling point over `yeqown/go-qrcode`. The repo name is generic
  and collides with several well-known libraries (`skip2`, `yeqown`, `yougg`),
  so the README's first paragraph must state the differentiator — styled dots,
  styled finder patterns, decorated logo, SVG output — within the first two
  sentences. Do not open with "a QR code library for Go".
- **Attribution:** credit `zxpsuper/qrcode-with-logos` (MIT) in the README and
  LICENSE notices as the design reference.

---

## 10. Working style

- Ask before adding a runtime dependency or changing the public API shape in §3.
- When the reference library's behaviour is ambiguous, read its actual source
  in `src/` rather than guessing from the docs site.
- Prefer a small, correct, well-tested surface over breadth. Twelve dot shapes
  that all decode reliably beat twenty that mostly do.
- Report honestly when something does not work. A phase is not done until its
  round-trip decode tests pass.
