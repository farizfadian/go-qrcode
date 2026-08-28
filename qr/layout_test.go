package qr

import (
	"errors"
	"testing"
)

func TestNewLayoutUsesIntegerModulesAndCentres(t *testing.T) {
	// 37 modules plus a 4-module quiet zone on each side is 45 across.
	// 380 / 45 = 8.44, so the module size is 8 and 380 - 360 = 20 spare pixels
	// split evenly become a 10 pixel offset.
	l, err := newLayout(37, 4, 380)
	if err != nil {
		t.Fatalf("newLayout: %v", err)
	}
	if l.ModuleSize != 8 {
		t.Errorf("ModuleSize = %v, want 8", l.ModuleSize)
	}
	if l.OriginX != 10 || l.OriginY != 10 {
		t.Errorf("Origin = (%v,%v), want (10,10)", l.OriginX, l.OriginY)
	}
	if l.Width != 380 {
		t.Errorf("Width = %d, want 380", l.Width)
	}
}

func TestLayoutRectSkipsTheQuietZone(t *testing.T) {
	l, err := newLayout(21, 4, 290) // 29 across, 290/29 = 10 exactly
	if err != nil {
		t.Fatalf("newLayout: %v", err)
	}
	px, py, size := l.Rect(0, 0)
	if size != 10 {
		t.Errorf("size = %v, want 10", size)
	}
	// Module (0,0) sits after four quiet-zone modules.
	if px != 40 || py != 40 {
		t.Errorf("Rect(0,0) = (%v,%v), want (40,40)", px, py)
	}
	px, _, _ = l.Rect(20, 0)
	if px != 240 {
		t.Errorf("Rect(20,0).px = %v, want 240", px)
	}
}

func TestNewLayoutRejectsWidthBelowOnePixelPerModule(t *testing.T) {
	_, err := newLayout(177, 4, 100) // 185 modules across, 100 pixels
	if !errors.Is(err, ErrWidthTooSmall) {
		t.Fatalf("error = %v, want ErrWidthTooSmall", err)
	}
}

func TestNewLayoutRejectsNonPositiveInput(t *testing.T) {
	if _, err := newLayout(0, 4, 380); err == nil {
		t.Error("newLayout accepted zero modules")
	}
	if _, err := newLayout(21, -1, 380); err == nil {
		t.Error("newLayout accepted a negative margin")
	}
}
