package h3_test

import (
	"testing"

	h3 "github.com/lightboxre/h3-go"
)

func TestCellToChildren_SameRes(t *testing.T) {
	// Get a cell at resolution 5
	cell := h3.LatLngToCell(37.0, -122.0, 5)

	// Request children at same resolution
	children := h3.CellToChildren(cell, 5)

	// Should return the cell itself
	if len(children) != 1 {
		t.Errorf("expected 1 child (self) for same resolution, got %d", len(children))
	}

	if len(children) > 0 && children[0] != cell {
		t.Errorf("expected child to be original cell %v, got %v", cell, children[0])
	}
}

func TestCellToChildren_ChildResTooFine(t *testing.T) {
	// Get a cell at resolution 10
	cell := h3.LatLngToCell(37.0, -122.0, 10)

	// Request children beyond maximum resolution (15)
	children := h3.CellToChildren(cell, 16)

	// Should return nil or empty for invalid resolution
	if children != nil && len(children) > 0 {
		t.Errorf("expected nil or empty for resolution beyond max, got %d children", len(children))
	}
}
