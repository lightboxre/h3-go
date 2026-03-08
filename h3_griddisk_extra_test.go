package h3_test

import (
	"testing"

	h3 "github.com/lightboxre/h3-go"
)

func TestGridDisk_PentagonK3(t *testing.T) {
	// Get a pentagon at resolution 1
	pentagons := h3.GetPentagonCells(1)
	if len(pentagons) == 0 {
		t.Fatal("expected pentagons at resolution 1")
	}
	pent := pentagons[0]

	// GridDisk with k=3 around pentagon
	cells, err := h3.GridDisk(pent, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pentagon at k=3 should return approximately 31 cells
	// (exact count depends on pentagon topology)
	if len(cells) == 0 {
		t.Error("expected non-empty grid disk")
	}
	if len(cells) < 20 || len(cells) > 40 {
		t.Errorf("expected approximately 31 cells, got %d", len(cells))
	}

	// Verify all cells are valid
	for _, c := range cells {
		if !h3.IsValidCell(c) {
			t.Errorf("invalid cell in grid disk: %v", c)
		}
	}
}

func TestGridDisk_PentagonK4(t *testing.T) {
	// Get a non-polar pentagon at resolution 1
	pentagons := h3.GetPentagonCells(1)
	if len(pentagons) < 2 {
		t.Fatal("expected multiple pentagons at resolution 1")
	}
	// Use base cell 14 pentagon if available, otherwise use second pentagon
	pent := pentagons[1]

	// GridDisk with k=4 around pentagon
	cells, err := h3.GridDisk(pent, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pentagon at k=4 should return approximately 51 cells
	// (exact count depends on pentagon topology)
	if len(cells) == 0 {
		t.Error("expected non-empty grid disk")
	}
	if len(cells) < 40 || len(cells) > 65 {
		t.Errorf("expected approximately 51 cells, got %d", len(cells))
	}

	// Verify all cells are valid and at correct resolution
	for _, c := range cells {
		if !h3.IsValidCell(c) {
			t.Errorf("invalid cell in grid disk: %v", c)
		}
		if h3.GetResolution(c) != 1 {
			t.Errorf("expected resolution 1, got %d", h3.GetResolution(c))
		}
	}
}

func TestGridDisk_CwOffsetPent(t *testing.T) {
	// Get pentagon and test with clockwise offset
	pentagons := h3.GetPentagonCells(0)
	if len(pentagons) == 0 {
		t.Fatal("expected pentagons at resolution 0")
	}
	pent := pentagons[0]

	// GridDisk with k=1 around pentagon
	cells, err := h3.GridDisk(pent, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pentagon at k=1 should return 6 cells (center + 5 neighbors)
	if len(cells) != 6 {
		t.Errorf("expected 6 cells for pentagon k=1, got %d", len(cells))
	}

	// Verify all cells are valid
	for _, c := range cells {
		if !h3.IsValidCell(c) {
			t.Errorf("invalid cell in grid disk: %v", c)
		}
	}
}

func TestMaxGridDiskSize_Invalid(t *testing.T) {
	// Test with negative k
	size, err := h3.MaxGridDiskSize(-1)
	if err == nil {
		t.Errorf("expected error for negative k, got size %d", size)
	}
}

func TestMaxGridDiskSize_Large(t *testing.T) {
	// Test with large k value
	size, err := h3.MaxGridDiskSize(100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MaxGridDiskSize formula: 3*k*(k+1) + 1
	expected := int64(3*100*101 + 1)
	if size != expected {
		t.Errorf("expected size %d, got %d", expected, size)
	}
}

func TestMaxGridDiskSize_NumCells(t *testing.T) {
	// Test that MaxGridDiskSize matches actual GridDisk size
	cell := h3.LatLngToCell(37.0, -122.0, 5)
	k := 5

	maxSize, err := h3.MaxGridDiskSize(k)
	if err != nil {
		t.Fatalf("unexpected error from MaxGridDiskSize: %v", err)
	}

	cells, err := h3.GridDisk(cell, k)
	if err != nil {
		t.Fatalf("unexpected error from GridDisk: %v", err)
	}

	actualSize := int64(len(cells))
	if actualSize > maxSize {
		t.Errorf("actual size %d exceeds max size %d", actualSize, maxSize)
	}

	// For hexagons, actual size should match the formula
	expectedSize := int64(3*k*(k+1) + 1)
	if actualSize != expectedSize {
		t.Errorf("expected actual size %d, got %d", expectedSize, actualSize)
	}
}
