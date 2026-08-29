package render

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
)

const sampleSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
	`<rect x="10" y="10" width="80" height="80" rx="12" fill="#0f766e"/>` +
	`</svg>`

func TestNestSVGPositionsTheDocumentWithoutTouchingItsContents(t *testing.T) {
	got, err := nestSVG(sampleSVG, 10, 20, 30, 40)
	if err != nil {
		t.Fatalf("nestSVG: %v", err)
	}

	for _, want := range []string{
		`x="10"`, `y="20"`, `width="30"`, `height="40"`,
		`viewBox="0 0 100 100"`,         // the author's own attributes survive
		`preserveAspectRatio="xMidYMid`, // added so a logo is not stretched
		`rx="12"`,                       // the payload is passed through verbatim
		`fill="#0f766e"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot: %s", want, got)
		}
	}
}

// Nothing inside the document may be reinterpreted. Anything this code
// understood, it could also silently lose — so it understands only the root
// tag's placement.
func TestNestSVGPassesTheBodyThroughByteForByte(t *testing.T) {
	body := `<defs><linearGradient id="g"><stop offset="0%" stop-color="#fff"/></linearGradient></defs>` +
		`<path d="M0 0 L10 10" stroke="url(#g)" stroke-width="2"/>` +
		`<text x="5" y="5">hi &amp; bye</text>`
	src := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">` + body + `</svg>`

	got, err := nestSVG(src, 0, 0, 100, 100)
	if err != nil {
		t.Fatalf("nestSVG: %v", err)
	}
	if !strings.Contains(got, body) {
		t.Errorf("the body was altered\nwant to contain: %s\ngot: %s", body, got)
	}
}

func TestNestSVGHandlesADeclarationAndSelfClosingRoot(t *testing.T) {
	withDecl := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + sampleSVG
	if _, err := nestSVG(withDecl, 0, 0, 10, 10); err != nil {
		t.Errorf("an XML declaration was rejected: %v", err)
	}

	empty := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"/>`
	got, err := nestSVG(empty, 0, 0, 10, 10)
	if err != nil {
		t.Fatalf("a self-closing root was rejected: %v", err)
	}
	if !strings.HasSuffix(got, "/>") {
		t.Errorf("a self-closing root should stay self-closing: %s", got)
	}
}

func TestNestSVGRejectsWhatIsNotSVG(t *testing.T) {
	for _, tc := range []struct{ name, markup string }{
		{"empty", ""},
		{"plain text", "hello"},
		{"html", "<html><body/></html>"},
		{"unterminated tag", `<svg xmlns="x" viewBox="0 0 1 1"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateSVG(tc.markup); !errors.Is(err, ErrNotSVG) {
				t.Errorf("error = %v, want ErrNotSVG", err)
			}
		})
	}
}

// A truncated document would corrupt the SVG it is embedded in, so the whole
// thing is parsed, not just the root tag.
func TestValidateSVGRejectsATruncatedDocument(t *testing.T) {
	truncated := `<svg xmlns="http://www.w3.org/2000/svg"><g><rect/>`
	if err := ValidateSVG(truncated); !errors.Is(err, ErrNotSVG) {
		t.Errorf("error = %v, want ErrNotSVG for unclosed elements", err)
	}
	if err := ValidateSVG(sampleSVG); err != nil {
		t.Errorf("a valid document was rejected: %v", err)
	}
}

// The embedded result must leave the surrounding document parseable. This is
// the property that matters most: a broken nest breaks the whole QR code.
func TestSceneWithNestedSVGProducesWellFormedXML(t *testing.T) {
	sc := Scene{
		Width: 100, Height: 100,
		Items: []Item{ImageItem{
			SVGMarkup: sampleSVG,
			X:         20, Y: 20, W: 60, H: 60,
		}},
	}
	out, err := SVG(sc)
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	assertWellFormedXML(t, out)
	if !strings.Contains(out, `fill="#0f766e"`) {
		t.Errorf("the vector logo was not embedded:\n%s", out)
	}
	if strings.Contains(out, "data:image/png") {
		t.Error("a bitmap was embedded even though vector markup was available")
	}
}

func TestNestedSVGWithAClipStaysWellFormed(t *testing.T) {
	clip := Circle(50, 50, 20)
	sc := Scene{
		Width: 100, Height: 100,
		Items: []Item{ImageItem{
			SVGMarkup: sampleSVG,
			X:         30, Y: 30, W: 40, H: 40,
			Clip: &clip,
		}},
	}
	out, err := SVG(sc)
	if err != nil {
		t.Fatalf("SVG: %v", err)
	}
	assertWellFormedXML(t, out)
	if !strings.Contains(out, "<clipPath ") || !strings.Contains(out, `clip-path="url(#`) {
		t.Errorf("the clip was dropped:\n%s", out)
	}
}

func assertWellFormedXML(t *testing.T, s string) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(s))
	for {
		_, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			t.Fatalf("output is not well-formed XML: %v\n%s", err, s)
		}
	}
}
