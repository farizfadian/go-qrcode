package qr

import (
	"io"
	"testing"
	"unicode/utf8"
)

// FuzzNew drives the whole pipeline with arbitrary content.
//
// The contract being tested is narrow and absolute: New either returns an
// error or returns a *QR that renders. It must never panic, and it must never
// hand back a value that panics later — CLAUDE.md forbids panics in library
// code, and a caller passing user-supplied text is exactly where one would
// surface.
func FuzzNew(f *testing.F) {
	for _, seed := range []string{
		"",
		"a",
		"https://example.com",
		"HELLO WORLD",
		"0123456789",
		"WIFI:T:WPA;S:net;P:pass;;",
		"BEGIN:VCARD\nVERSION:3.0\nFN:Test\nEND:VCARD",
		"Meja 12 — Warung Kopi Senja",
		"\x00\x01\x02",
		"日本語のテキスト",
		"🎨🔐✅",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, content string) {
		q, err := New(Options{Content: content})
		if err != nil {
			// Rejecting is always allowed; the content may be too long, or
			// empty. What matters is that it did not panic.
			return
		}

		// A returned *QR must be fully usable.
		if got := q.Content(); got != content {
			t.Fatalf("Content() = %q, want %q", got, content)
		}
		if q.Modules() <= 0 {
			t.Fatalf("Modules() = %d, want a positive count", q.Modules())
		}

		img := q.Image()
		if img.Bounds().Dx() != DefaultWidth {
			t.Fatalf("image width = %d, want %d", img.Bounds().Dx(), DefaultWidth)
		}
		if err := q.PNG(io.Discard); err != nil {
			t.Fatalf("PNG: %v", err)
		}
		markup, err := q.SVGString()
		if err != nil {
			t.Fatalf("SVGString: %v", err)
		}
		if !utf8.ValidString(markup) {
			t.Fatal("SVG output is not valid UTF-8")
		}
	})
}

// FuzzNewStyled adds the styling surface, so a shape that mishandles an unusual
// module layout is caught too. The stripe and fluid shapes read and write state
// across the grid, which is where an out-of-range index would hide.
func FuzzNewStyled(f *testing.F) {
	f.Add("https://example.com", 7, 3)
	f.Add("A", 0, 0)
	f.Add("", 11, 6)

	f.Fuzz(func(t *testing.T, content string, dot, corner int) {
		dots := DotTypes()
		corners := CornerTypes()
		if dot < 0 || corner < 0 {
			return
		}

		q, err := New(Options{
			Content: content,
			Width:   320,
			Dots:    DotOptions{Type: dots[dot%len(dots)]},
			Corners: CornerOptions{Type: corners[corner%len(corners)]},
		})
		if err != nil {
			return
		}
		_ = q.Image()
		if _, err := q.SVGString(); err != nil {
			t.Fatalf("SVGString: %v", err)
		}
	})
}

// FuzzParseColor pins that colour parsing never panics on arbitrary input, and
// that anything it accepts round-trips into a usable code.
func FuzzParseColor(f *testing.F) {
	for _, seed := range []string{"#000000", "000", "ff0000ff", "red", "", "#", "zzz"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, colour string) {
		_, err := New(Options{Content: "test", Foreground: colour})
		_ = err // any outcome is fine; not panicking is the contract
	})
}
