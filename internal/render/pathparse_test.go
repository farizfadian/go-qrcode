package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

// extractDAttributes pulls every path's d attribute out of SVG markup, in
// document order.
func extractDAttributes(svg string) []string {
	var out []string
	rest := svg
	for {
		i := strings.Index(rest, ` d="`)
		if i < 0 {
			return out
		}
		rest = rest[i+4:]
		j := strings.IndexByte(rest, '"')
		if j < 0 {
			return out
		}
		out = append(out, rest[:j])
		rest = rest[j+1:]
	}
}

// parsePathD turns a d attribute back into a Path. It understands only the
// absolute commands the serialiser emits: M, L, Q, C and Z. Anything else is an
// error, so a serialiser that starts emitting arcs fails loudly rather than
// silently diverging from the rasteriser.
func parsePathD(d string) (Path, error) {
	fields := strings.Fields(strings.NewReplacer(
		"M", " M ", "L", " L ", "Q", " Q ", "C", " C ", "Z", " Z ", ",", " ",
	).Replace(d))

	need := map[string]int{"M": 2, "L": 2, "Q": 4, "C": 6, "Z": 0}
	var b Builder
	for i := 0; i < len(fields); {
		cmd := fields[i]
		i++
		n, ok := need[cmd]
		if !ok {
			return Path{}, fmt.Errorf("d: unsupported command %q", cmd)
		}
		vals := make([]float64, n)
		for k := 0; k < n; k++ {
			if i+k >= len(fields) {
				return Path{}, fmt.Errorf("d: %s ran out of numbers", cmd)
			}
			v, err := strconv.ParseFloat(fields[i+k], 64)
			if err != nil {
				return Path{}, fmt.Errorf("d: %s: %w", cmd, err)
			}
			vals[k] = v
		}
		i += n
		switch cmd {
		case "M":
			b.MoveTo(vals[0], vals[1])
		case "L":
			b.LineTo(vals[0], vals[1])
		case "Q":
			b.QuadTo(vals[0], vals[1], vals[2], vals[3])
		case "C":
			b.CubeTo(vals[0], vals[1], vals[2], vals[3], vals[4], vals[5])
		case "Z":
			b.Close()
		}
	}
	return b.Path(), nil
}

// assertPathsEqual compares two paths to the precision the serialiser emits.
func assertPathsEqual(t *testing.T, got, want Path) {
	t.Helper()
	const eps = 1e-3
	near := func(a, b float64) bool { return math.Abs(a-b) <= eps }
	nearPt := func(a, b Point) bool { return near(a.X, b.X) && near(a.Y, b.Y) }

	if len(got.SubPaths) != len(want.SubPaths) {
		t.Fatalf("subpaths = %d, want %d", len(got.SubPaths), len(want.SubPaths))
	}
	for i := range want.SubPaths {
		g, w := got.SubPaths[i], want.SubPaths[i]
		if !nearPt(g.Start, w.Start) {
			t.Errorf("subpath %d start = %v, want %v", i, g.Start, w.Start)
		}
		if g.Closed != w.Closed {
			t.Errorf("subpath %d closed = %v, want %v", i, g.Closed, w.Closed)
		}
		if len(g.Segs) != len(w.Segs) {
			t.Fatalf("subpath %d segments = %d, want %d", i, len(g.Segs), len(w.Segs))
		}
		for j := range w.Segs {
			gs, ws := g.Segs[j], w.Segs[j]
			if gs.Kind != ws.Kind {
				t.Errorf("subpath %d seg %d kind = %v, want %v", i, j, gs.Kind, ws.Kind)
				continue
			}
			if !nearPt(gs.To, ws.To) {
				t.Errorf("subpath %d seg %d end = %v, want %v", i, j, gs.To, ws.To)
			}
			if ws.Kind != SegLine && !nearPt(gs.C1, ws.C1) {
				t.Errorf("subpath %d seg %d c1 = %v, want %v", i, j, gs.C1, ws.C1)
			}
			if ws.Kind == SegCube && !nearPt(gs.C2, ws.C2) {
				t.Errorf("subpath %d seg %d c2 = %v, want %v", i, j, gs.C2, ws.C2)
			}
		}
	}
}

func TestParsePathDRejectsUnknownCommands(t *testing.T) {
	if _, err := parsePathD("M 0 0 A 1 1 0 0 1 2 2"); err == nil {
		t.Fatal("parsePathD accepted an arc command the serialiser never emits")
	}
}
