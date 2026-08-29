package render

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrNotSVG reports markup that is not a standalone SVG document.
var ErrNotSVG = errors.New("render: not an SVG document")

// nestSVG places an SVG document inside another one, positioned and sized to
// the given box.
//
// SVG 1.1 allows an <svg> element inside an <svg> element, and gives the nested
// one its own x, y, width, height and coordinate system. That is exactly what is
// wanted here, and it means the nested artwork is rendered by whatever displays
// the page rather than being flattened by us — gradients, curves, strokes and
// text all survive.
//
// The work is therefore only to rewrite the root element's positioning
// attributes. Everything inside it is passed through untouched, which is what
// keeps this honest: nothing is reinterpreted, so nothing can be silently lost.
func nestSVG(markup string, x, y, w, h float64) (string, error) {
	open, rest, err := splitRootTag(markup)
	if err != nil {
		return "", err
	}

	attrs, selfClosing, err := parseRootAttrs(open)
	if err != nil {
		return "", err
	}

	// Our placement wins; anything else the author set is preserved.
	set := map[string]string{
		"x":      num(x),
		"y":      num(y),
		"width":  num(w),
		"height": num(h),
	}
	// preserveAspectRatio keeps a non-square logo from being stretched into the
	// box, matching how the rasteriser letterboxes it.
	if _, ok := attrs["preserveAspectRatio"]; !ok {
		set["preserveAspectRatio"] = "xMidYMid meet"
	}

	var b strings.Builder
	b.WriteString("<svg")
	for _, k := range attrOrder(attrs, set) {
		v, ok := set[k]
		if !ok {
			v = attrs[k]
		}
		fmt.Fprintf(&b, ` %s="%s"`, k, escapeAttr(v))
	}
	if selfClosing {
		b.WriteString("/>")
		return b.String(), nil
	}
	b.WriteString(">")
	b.WriteString(rest)
	return b.String(), nil
}

// splitRootTag returns the opening <svg ...> tag and everything after it,
// skipping any XML declaration, comments or doctype that precede it.
func splitRootTag(markup string) (open, rest string, err error) {
	i := strings.Index(markup, "<svg")
	if i < 0 {
		return "", "", fmt.Errorf("%w: no <svg> element found", ErrNotSVG)
	}
	// Find the end of the opening tag, respecting quoted attribute values.
	inQuote := byte(0)
	for j := i; j < len(markup); j++ {
		c := markup[j]
		switch {
		case inQuote != 0:
			if c == inQuote {
				inQuote = 0
			}
		case c == '"' || c == '\'':
			inQuote = c
		case c == '>':
			return markup[i : j+1], markup[j+1:], nil
		}
	}
	return "", "", fmt.Errorf("%w: the <svg> tag is never closed", ErrNotSVG)
}

// parseRootAttrs reads the attributes of an opening tag. It leans on
// encoding/xml so quoting and entities are handled by the standard library
// rather than by hand.
func parseRootAttrs(open string) (map[string]string, bool, error) {
	selfClosing := strings.HasSuffix(strings.TrimSpace(open), "/>")
	probe := open
	if selfClosing {
		probe = strings.TrimSuffix(strings.TrimSpace(open), "/>") + "></svg>"
	}

	dec := xml.NewDecoder(strings.NewReader(probe))
	tok, err := dec.Token()
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrNotSVG, err)
	}
	start, ok := tok.(xml.StartElement)
	if !ok || start.Name.Local != "svg" {
		return nil, false, fmt.Errorf("%w: root element is not <svg>", ErrNotSVG)
	}

	attrs := make(map[string]string, len(start.Attr))
	for _, a := range start.Attr {
		name := a.Name.Local
		// encoding/xml reports xmlns as a Space; rebuild the source spelling.
		if a.Name.Space == "xmlns" {
			name = "xmlns:" + a.Name.Local
		} else if a.Name.Local == "xmlns" {
			name = "xmlns"
		} else if a.Name.Space != "" {
			name = a.Name.Space + ":" + a.Name.Local
		}
		attrs[name] = a.Value
	}
	return attrs, selfClosing, nil
}

// attrOrder returns the attribute names to emit: the author's, then any of ours
// they did not already have, so output is deterministic.
func attrOrder(attrs, set map[string]string) []string {
	seen := make(map[string]bool, len(attrs)+len(set))
	var out []string
	for _, k := range []string{"xmlns", "xmlns:xlink", "viewBox", "x", "y", "width", "height", "preserveAspectRatio"} {
		if _, inA := attrs[k]; inA {
			out, seen[k] = append(out, k), true
		} else if _, inS := set[k]; inS {
			out, seen[k] = append(out, k), true
		}
	}
	// Anything else the author set, in sorted order for determinism.
	var extra []string
	for k := range attrs {
		if !seen[k] {
			extra = append(extra, k)
		}
	}
	sortStrings(extra)
	return append(out, extra...)
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func escapeAttr(v string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(v)); err != nil {
		return v
	}
	return b.String()
}

// ValidateSVG reports whether markup is a usable standalone SVG document. It is
// called when options are validated, so a malformed logo fails at New rather
// than at render time.
func ValidateSVG(markup string) error {
	if strings.TrimSpace(markup) == "" {
		return fmt.Errorf("%w: empty", ErrNotSVG)
	}
	if _, err := nestSVG(markup, 0, 0, 1, 1); err != nil {
		return err
	}
	// The whole document must parse, not just its root tag: a truncated file
	// would otherwise corrupt the SVG it is embedded in.
	dec := xml.NewDecoder(strings.NewReader(markup))
	for {
		_, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%w: %v", ErrNotSVG, err)
		}
	}
}
