package qr

import "errors"

// ModuleKind labels a module's role in the symbol. Only KindFinder changes how
// the code is drawn today: finder modules are replaced by a styled corner
// figure. The remaining kinds exist so the classification can be tested against
// known symbols and inspected while debugging.
type ModuleKind uint8

// The module roles defined by ISO/IEC 18004. Alignment patterns render as
// ordinary dots and timing and format modules are data, so only KindFinder is
// excluded from dot rendering.
const (
	KindData ModuleKind = iota
	KindFinder
	KindSeparator
	KindTiming
	KindAlignment
	KindFormat
	KindVersion
)

// ErrBadMatrix reports a module grid that is empty or not square.
var ErrBadMatrix = errors.New("qr: module matrix must be square and non-empty")

// Matrix is an immutable QR module grid with every module classified.
type Matrix struct {
	size int
	dark []bool
	kind []ModuleKind
}

// newMatrix classifies mods, which must be a square [y][x] grid.
func newMatrix(mods [][]bool) (*Matrix, error) {
	n := len(mods)
	if n == 0 {
		return nil, ErrBadMatrix
	}
	for _, row := range mods {
		if len(row) != n {
			return nil, ErrBadMatrix
		}
	}
	m := &Matrix{size: n, dark: make([]bool, n*n), kind: make([]ModuleKind, n*n)}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			m.dark[y*n+x] = mods[y][x]
		}
	}
	m.classify()
	return m, nil
}

// Size returns the side length in modules.
func (m *Matrix) Size() int { return m.size }

// Dark reports whether the module at (x, y) is dark. Coordinates outside the
// grid read as light.
func (m *Matrix) Dark(x, y int) bool {
	if !m.inBounds(x, y) {
		return false
	}
	return m.dark[y*m.size+x]
}

// Kind returns the module's role. Coordinates outside the grid report KindData.
func (m *Matrix) Kind(x, y int) ModuleKind {
	if !m.inBounds(x, y) {
		return KindData
	}
	return m.kind[y*m.size+x]
}

// InFinder reports whether (x, y) lies inside one of the three 7x7 finder
// patterns. This is the classification the renderer actually depends on.
func (m *Matrix) InFinder(x, y int) bool { return m.Kind(x, y) == KindFinder }

func (m *Matrix) inBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < m.size && y < m.size
}

func (m *Matrix) set(x, y int, k ModuleKind) {
	if m.inBounds(x, y) {
		m.kind[y*m.size+x] = k
	}
}

func (m *Matrix) classify() {
	n := m.size
	version := (n - 17) / 4

	// Finder patterns and their separators, at three corners.
	for _, c := range [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}} {
		ox, oy := c[0], c[1]
		for dy := -1; dy <= 7; dy++ {
			for dx := -1; dx <= 7; dx++ {
				x, y := ox+dx, oy+dy
				if !m.inBounds(x, y) {
					continue
				}
				if dx >= 0 && dx < 7 && dy >= 0 && dy < 7 {
					m.set(x, y, KindFinder)
				} else {
					m.set(x, y, KindSeparator)
				}
			}
		}
	}

	// Timing patterns run along row 6 and column 6 between the finders.
	for i := 8; i < n-8; i++ {
		m.set(i, 6, KindTiming)
		m.set(6, i, KindTiming)
	}

	// Alignment patterns, 5x5, at every pairing of the version's centres except
	// the three that would collide with a finder pattern.
	centres := alignmentCentres(version)
	for _, cy := range centres {
		for _, cx := range centres {
			if (cx <= 8 && cy <= 8) || (cx <= 8 && cy >= n-9) || (cx >= n-9 && cy <= 8) {
				continue
			}
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					m.set(cx+dx, cy+dy, KindAlignment)
				}
			}
		}
	}

	// Format information sits beside the finder patterns.
	for i := 0; i <= 8; i++ {
		if i != 6 {
			m.set(i, 8, KindFormat)
			m.set(8, i, KindFormat)
		}
	}
	for i := 0; i < 8; i++ {
		m.set(8, n-1-i, KindFormat)
		m.set(n-1-i, 8, KindFormat)
	}

	// Version information, present from version 7 onwards.
	if version >= 7 {
		for i := 0; i < 6; i++ {
			for j := 0; j < 3; j++ {
				m.set(i, n-11+j, KindVersion)
				m.set(n-11+j, i, KindVersion)
			}
		}
	}
}

// alignmentCentres returns the row and column centres of the alignment patterns
// for a version, per ISO/IEC 18004 annex E. Version 1 has none.
//
// The spacing is derived rather than tabulated. Every centre but the first is
// spaced evenly back from 4*version+10, and the step is rounded up to an even
// number. Version 32 is the single version the general rule does not produce.
func alignmentCentres(version int) []int {
	if version < 2 || version > 40 {
		return nil
	}
	n := version/7 + 2
	step := 26 // version 32 is the one case the formula below does not yield
	if version != 32 {
		step = (version*4 + n*2 + 1) / (n*2 - 2) * 2
	}
	out := make([]int, n)
	out[0] = 6
	for i, pos := n-1, version*4+10; i >= 1; i, pos = i-1, pos-step {
		out[i] = pos
	}
	return out
}
