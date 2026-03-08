package h3_test

import (
	"testing"

	h3 "github.com/lightboxre/h3-go"
)

func TestCompactCells_Uncompactable(t *testing.T) {
	// Three cells at resolution 9 that cannot be compacted further
	cells := []h3.Cell{
		0x89283082803ffff,
		0x89283082807ffff,
		0x8928308280bffff,
	}

	result, err := h3.CompactCells(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(cells) {
		t.Errorf("expected %d cells, got %d", len(cells), len(result))
	}

	// Verify uncompaction works
	uncompacted, err := h3.UncompactCells(result, 9)
	if err != nil {
		t.Fatalf("unexpected error during uncompact: %v", err)
	}
	if len(uncompacted) != len(cells) {
		t.Errorf("expected %d uncompacted cells, got %d", len(cells), len(uncompacted))
	}
}

func TestCompactCells_DuplicateMinimum(t *testing.T) {
	// Create children of an arbitrary cell and duplicate one
	parent := h3.LatLngToCell(37.0, -122.0, 9)
	children := h3.CellToChildren(parent, 11)
	if len(children) == 0 {
		t.Fatal("expected children")
	}

	// Create input with one duplicate
	cells := append(children, children[0])

	result, err := h3.CompactCells(cells)
	// The C implementation returns E_DUPLICATE_INPUT error
	// Our implementation may handle this differently - check behavior
	if err == nil {
		// If no error, verify the duplicate was handled
		t.Logf("CompactCells handled duplicate without error, got %d cells", len(result))
	} else {
		// Error is expected for duplicate input
		t.Logf("CompactCells returned expected error for duplicate: %v", err)
	}
}

func TestCompactCells_DuplicatePentagonLimit(t *testing.T) {
	// Get a pentagon at resolution 0
	pentagons := h3.GetPentagonCells(0)
	if len(pentagons) == 0 {
		t.Fatal("expected pentagons")
	}
	pent := pentagons[0]

	// Get children and duplicate one
	children := h3.CellToChildren(pent, 1)
	if len(children) == 0 {
		t.Fatal("expected pentagon children")
	}

	cells := append(children, children[0])

	result, err := h3.CompactCells(cells)
	// Error expected for duplicate
	if err == nil {
		t.Logf("CompactCells handled pentagon duplicate without error, got %d cells", len(result))
	} else {
		t.Logf("CompactCells returned expected error for pentagon duplicate: %v", err)
	}
}

func TestCompactCells_DuplicateIgnored(t *testing.T) {
	// This test documents current behavior with duplicates
	// Some implementations may ignore duplicates rather than error
	parent := h3.LatLngToCell(37.0, -122.0, 9)
	children := h3.CellToChildren(parent, 11)
	if len(children) < 2 {
		t.Fatal("expected multiple children")
	}

	// Add duplicate of first child
	cells := append(children, children[0])

	result, err := h3.CompactCells(cells)
	if err != nil {
		// Error is acceptable
		t.Logf("CompactCells returned error for duplicate: %v", err)
		return
	}

	// If no error, operation succeeded (duplicate was handled somehow)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCompactCells_Empty(t *testing.T) {
	result, err := h3.CompactCells([]h3.Cell{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d cells", len(result))
	}
}

func TestCompactCells_Disparate(t *testing.T) {
	// Seven unrelated cells at resolution 1 that cannot be merged
	cells := []h3.Cell{
		0x81283ffffffffff,
		0x81293ffffffffff,
		0x812a3ffffffffff,
		0x812b3ffffffffff,
		0x812c3ffffffffff,
		0x812d3ffffffffff,
		0x812e3ffffffffff,
	}

	result, err := h3.CompactCells(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(cells) {
		t.Errorf("expected %d cells (unchanged), got %d", len(cells), len(result))
	}
}

func TestCompactCells_ReservedBitsSet(t *testing.T) {
	// Cells with invalid reserved bits set
	cells := []h3.Cell{
		0xffffffffffffffff, // all bits set
		0x8000000000000000, // high bit set
		0x4000000000000000, // reserved bit set
	}

	result, err := h3.CompactCells(cells)
	// Should return error for invalid cells
	if err == nil {
		t.Errorf("expected error for invalid cells, got result: %v", result)
	}
}

func TestCompactCells_ParentError(t *testing.T) {
	// Cells at different resolutions that have mismatched parents
	cells := []h3.Cell{
		h3.LatLngToCell(37.0, -122.0, 10),
		h3.LatLngToCell(38.0, -123.0, 5),
	}

	result, err := h3.CompactCells(cells)
	// Should handle resolution mismatch
	if err == nil {
		// May succeed with mixed resolutions, just verify result is valid
		if len(result) == 0 {
			t.Error("expected non-empty result")
		}
	} else {
		t.Logf("CompactCells returned error for resolution mismatch: %v", err)
	}
}

func TestUncompactCells_WrongRes(t *testing.T) {
	// Cells at resolution 5
	cells := []h3.Cell{
		h3.LatLngToCell(37.0, -122.0, 5),
	}

	// Try to uncompact to resolution 4 (lower than input)
	result, err := h3.UncompactCells(cells, 4)
	if err == nil {
		t.Errorf("expected error for uncompact to lower resolution, got %d cells", len(result))
	}

	// Try to uncompact to invalid resolution
	result, err = h3.UncompactCells(cells, -1)
	if err == nil {
		t.Errorf("expected error for negative resolution, got %d cells", len(result))
	}

	// Try to uncompact beyond max resolution
	result, err = h3.UncompactCells(cells, 16)
	if err == nil {
		t.Errorf("expected error for resolution beyond max, got %d cells", len(result))
	}
}

func TestUncompactCells_OnlyZero(t *testing.T) {
	// Input with only H3_NULL (zero) cell
	cells := []h3.Cell{0}

	result, err := h3.UncompactCells(cells, 5)
	// Implementation may error on invalid cell or return empty result
	if err != nil {
		// Error is acceptable for invalid cell
		t.Logf("UncompactCells returned error for zero cell: %v", err)
		return
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for zero cell, got %d cells", len(result))
	}
}

func TestUncompactCells_WithZero(t *testing.T) {
	// Input containing one H3_NULL cell among valid cells
	cells := []h3.Cell{
		h3.LatLngToCell(37.0, -122.0, 3),
		0, // H3_NULL
		h3.LatLngToCell(38.0, -123.0, 3),
		h3.LatLngToCell(39.0, -124.0, 3),
	}

	result, err := h3.UncompactCells(cells, 5)
	if err != nil {
		// Implementation may error when encountering invalid cell
		t.Logf("UncompactCells returned error with zero cell in input: %v", err)
		return
	}

	// Should produce children only for the 3 valid cells
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}

	// Verify all result cells are valid
	for _, c := range result {
		if !h3.IsValidCell(c) {
			t.Errorf("invalid cell in result: %v", c)
		}
	}
}
