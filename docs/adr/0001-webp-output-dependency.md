# ADR 0001 — Accept `nativewebp` as a third runtime dependency for WebP output

Date: 2026-08-29
Status: accepted

## Context

Fariz asked for WebP output alongside PNG, JPEG and SVG.

Go has no WebP encoder. Neither the standard library nor `golang.org/x/image`
provides one — `x/image/webp` is **decode-only**, which was verified by reading
the package: it contains `decode.go` and no `Encode` function.

CLAUDE.md §2 caps runtime dependencies at `github.com/piglig/go-qr` and
`golang.org/x/image`, and requires an explicit recorded decision before adding
another. The two-dependency promise is stated in the README as a selling point
and asserted by a CI job, so this is not a quiet change.

## Options considered

Three third-party encoders exist.

| Package | Pure Go? | Verdict |
|---|---|---|
| `github.com/chai2010/webp` | no — cgo, wraps libwebp | rejected |
| `github.com/kolesa-team/go-webp` | no — cgo, wraps libwebp | rejected |
| `github.com/HugoSmits86/nativewebp` | yes | accepted |

**The cgo options are disqualified outright.** The release workflow
cross-compiles `qrgen` for eleven platform and architecture pairs with
`CGO_ENABLED=0`. A cgo dependency would break every one of those builds, not
merely complicate them.

A fourth option was considered and rejected: leaving WebP out and telling users
to encode `q.Image()` themselves. It keeps the dependency count at two, but
pushes the same dependency onto every user who wants WebP while denying them the
`-out qr.webp` path in the CLI.

A build tag was also considered. It does not work: Go modules download a
dependency listed in `go.mod` whether or not a tag selects it, so the module
graph grows regardless. It would only avoid linking, at the cost of two build
configurations and an API that varies by tag.

## Decision

Add `github.com/HugoSmits86/nativewebp` v1.3.0 as a runtime dependency, and
expose `QR.WebP(io.Writer)` plus `-out x.webp` in the CLI.

## Consequences

Measured before deciding, not assumed.

**It works.** A test image encoded to a valid RIFF/WebP file and decoded back
through `x/image/webp` with a lossless, pixel-identical round trip.

**The size saving is real**, at 600 pixels wide:

| dot shape | PNG | WebP | saving |
|---|---|---|---|
| `square` | 3,067 B | 1,570 B | 48% |
| `fluid` | 11,819 B | 6,794 B | 42% |
| `stripe` | 8,904 B | 5,356 B | 39% |
| `star` | 14,138 B | 11,332 B | 19% |

**The Go floor moves from 1.22 to 1.22.2**, which `nativewebp` requires. This is
a patch-level bump, not a minor one: anyone on a current 1.22.x already has it.

**`golang.org/x/image` moves from v0.23.0 to v0.24.0** by minimum version
selection. This was checked against the constraint that pinned it in the first
place: v0.24.0's own `go` directive is still `go 1.18`, so the floor is
unaffected. v0.25.0 would raise it to 1.23 and must still be avoided.

**The dependency count becomes three.** The README and the CI assertion are
updated to say so. The promise being kept is not "exactly two" but "a small,
pure-Go set you can audit", and CI continues to enforce the exact list.

## Revisit if

- The standard library or `x/image` gains a WebP encoder — drop this dependency.
- `nativewebp` adds cgo, or requires an `x/image` version whose `go` directive
  raises this library's floor above what we are willing to promise.
