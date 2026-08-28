# Kickoff prompt — paste into Claude Code

This file is scaffolding, not part of the library. Add it to `.gitignore` if you
don't want it published.

Open Claude Code in this repo, then paste the block below as your first message.

---

```
Read CLAUDE.md in full before doing anything else. It is the specification for
this project — follow it, and tell me if anything in it is wrong or unworkable
rather than working around it silently.

We are building this from an empty repository. Today's task is Phase 1 only.
Do not start Phase 2. Do not create any barcode symbology package — see
CLAUDE.md §1 on sequencing.

Before writing any code, do this research and report back to me:

1. Fetch and read the source of https://github.com/zxpsuper/qrcode-with-logos
   — specifically `src/` and `types/`. I want to know how it classifies QR
   modules, how it decides which modules belong to the finder patterns, and how
   `fluid` / `fluid-line` / `stripe*` determine a module's shape from its
   neighbours. Summarise the neighbour-mask scheme it uses.

2. Evaluate pure-Go QR encoders that expose the raw module matrix (not just a
   finished image). Check `github.com/piglig/go-qr`, `rsc.io/qr`, and
   `github.com/boombuler/barcode/qr`. Read the actual package source — do not
   answer from memory. For each: does it expose a per-module accessor, what is
   its license, and is it maintained? Recommend one and say why.

3. Confirm `github.com/makiuchi-d/gozxing` can decode a QR from an
   `image.Image` in a test, and show me the exact call signature.

Stop after those three and wait for my go-ahead.

Once I approve, implement Phase 1 using the layout in CLAUDE.md §2:

- go.mod for module github.com/farizfadian/go-qrcode, Go 1.22
- MIT LICENSE with attribution to zxpsuper/qrcode-with-logos as design reference
- qr/encode.go — the Encoder interface plus the encoder you recommended
- qr/matrix.go — classify every module as finder / separator / timing /
  alignment / format / version / data. This classification is load-bearing for
  everything later, so test it thoroughly against known QR versions.
- internal/render/path.go — the shared vector path model. Design the dot-shape
  function signature to take a neighbour mask NOW, even though Phase 1 only
  implements `square`. Retrofitting this later is expensive.
- internal/render/raster.go using golang.org/x/image/vector
- internal/render/color.go — hex parsing per CLAUDE.md §3
- qr/shape_dots.go and qr/shape_corners.go with `square` only
- The public Options / New / PNG surface exactly as specified in CLAUDE.md §3

Definition of done for Phase 1:
- go test ./... -race passes
- gofmt -l . outputs nothing
- go vet ./... clean
- A round-trip test generates a QR from a URL, decodes it with gozxing, and
  asserts the decoded string matches — at all four ECC levels, at both a short
  and a long content string
- README stub explaining what the library is and what gap it fills

Work in a branch `phase-1-skeleton`. Commit in logical chunks with conventional
commit messages and the co-author trailer from CLAUDE.md §9. Do not push until
tests are green — then push and open a PR into main.

Three constraints I care about:
- Nothing in internal/render/ may import qr/ or know anything QR-specific.
- Do not add any runtime dependency beyond golang.org/x/image and the encoder
  we agreed on. Ask me first if you think you need one.
- No panics in library code. Errors only.
```

---

## Follow-up prompts for later phases

**Phase 2 (SVG):**

```
Phase 1 is merged. Implement Phase 2 per CLAUDE.md: internal/render/svg.go,
sharing the exact same Path model as the raster renderer. Add a parity test that
rasterises the SVG output and asserts it decodes to the same content as the PNG
and that the module geometry matches. If you find yourself duplicating any shape
logic between renderers, stop and tell me — that means the Path abstraction is
leaking.
```

**Phase 3 (shapes):**

```
Implement Phase 3: all 12 dot types and all 7 corner types from CLAUDE.md §4,
with independent colours and corner radii. Start with the neighbour-aware ones
(fluid, fluid-line, stripe, stripe-row, stripe-column) since they will stress
the shape signature hardest — if that signature needs to change, I want to know
on day one, not after nine shapes are written.

Every shape needs a round-trip decode test. A shape that renders beautifully and
does not scan does not ship.
```

**Phase 4 (logo):**

```
Implement Phase 4: logo support per CLAUDE.md §4 and the correctness rules in
§5. Pay particular attention to rules 1 and 2 — the ECC budget validation. I
would rather New() return an error than emit an unscannable code.

Test the full grid: each dot type with a logo present, each corner type with a
logo present, logo at several sizes including one deliberately oversized that
must be rejected.
```
