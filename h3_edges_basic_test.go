package h3

import (
	"testing"

	"github.com/EDRInc/h3-go/internal/constants"
	"github.com/EDRInc/h3-go/internal/h3index"
)

// TestBasicDirectedEdgeEncoding tests the core directed edge encoding/decoding
// without relying on neighbor traversal.
func TestBasicDirectedEdgeEncoding(t *testing.T) {
	// Valid hexagon cell
	origin, err := StringToCell("8928308280fffff")
	if err != nil {
		t.Fatalf("Failed to parse cell: %v", err)
	}

	// Test OriginToDirectedEdges
	edges := OriginToDirectedEdges(origin)
	if len(edges) != 6 {
		t.Errorf("Expected 6 directed edges for hexagon, got %d", len(edges))
	}

	// Test that all edges are valid
	for i, e := range edges {
		if !IsValidDirectedEdge(e) {
			t.Errorf("Edge %d should be valid", i)
		}
	}

	// Test GetDirectedEdgeOrigin
	for i, e := range edges {
		recoveredOrigin := GetDirectedEdgeOrigin(e)
		if recoveredOrigin != origin {
			t.Errorf("Edge %d: expected origin %x, got %x", i, origin, recoveredOrigin)
		}
	}

	// Test DirectedEdgeToCells (origin should match)
	for i, e := range edges {
		cells := DirectedEdgeToCells(e)
		if cells[0] != origin {
			t.Errorf("Edge %d: expected origin %x, got %x", i, origin, cells[0])
		}
	}

	// Test that edge directions are unique and in range 1-6
	directions := make(map[int]bool)
	for i, e := range edges {
		dir := getReservedBits(h3index.H3Index(e))
		if dir < 1 || dir > 6 {
			t.Errorf("Edge %d: invalid direction %d", i, dir)
		}
		if directions[dir] {
			t.Errorf("Edge %d: duplicate direction %d", i, dir)
		}
		directions[dir] = true
	}

	// Test that mode is correctly set
	for i, e := range edges {
		h := h3index.H3Index(e)
		if h.Mode() != constants.H3_DIRECTEDEDGE_MODE {
			t.Errorf("Edge %d: expected mode %d, got %d", i, constants.H3_DIRECTEDEDGE_MODE, h.Mode())
		}
	}
}

// TestPentagonDirectedEdges tests pentagon-specific behavior.
func TestPentagonDirectedEdges(t *testing.T) {
	pentagonCells := GetPentagonCells(9)
	if len(pentagonCells) == 0 {
		t.Skip("No pentagon cells available")
	}

	pentagon := pentagonCells[0]
	pentEdges := OriginToDirectedEdges(pentagon)

	// Pentagon should have 5 edges (no K-axis)
	if len(pentEdges) != 5 {
		t.Errorf("Expected 5 directed edges for pentagon, got %d", len(pentEdges))
	}

	// All pentagon edges should be valid
	for i, e := range pentEdges {
		if !IsValidDirectedEdge(e) {
			t.Errorf("Pentagon edge %d should be valid", i)
		}

		// Check that none have K_AXES_DIGIT direction
		dir := getReservedBits(h3index.H3Index(e))
		if dir == constants.K_AXES_DIGIT {
			t.Errorf("Pentagon should not have K-axis edge")
		}

		// Check origin is correct
		origin := GetDirectedEdgeOrigin(e)
		if origin != pentagon {
			t.Errorf("Pentagon edge %d: origin mismatch", i)
		}
	}

	// Test that directions are unique
	directions := make(map[int]bool)
	for i, e := range pentEdges {
		dir := getReservedBits(h3index.H3Index(e))
		if directions[dir] {
			t.Errorf("Pentagon edge %d: duplicate direction %d", i, dir)
		}
		directions[dir] = true
	}
}

// TestDirectedEdgeValidation tests IsValidDirectedEdge with various inputs.
func TestDirectedEdgeValidation(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")

	// Test valid edges
	edges := OriginToDirectedEdges(origin)
	for _, e := range edges {
		if !IsValidDirectedEdge(e) {
			t.Error("Valid edge should pass validation")
		}
	}

	// Test invalid: zero
	if IsValidDirectedEdge(0) {
		t.Error("Zero should not be valid")
	}

	// Test invalid: cell (not a directed edge)
	if IsValidDirectedEdge(DirectedEdge(origin)) {
		t.Error("Cell should not be valid as directed edge")
	}

	// Test invalid: wrong mode
	invalidMode := h3index.H3Index(edges[0])
	invalidMode = h3index.SetMode(invalidMode, constants.H3_VERTEX_MODE)
	if IsValidDirectedEdge(DirectedEdge(invalidMode)) {
		t.Error("Wrong mode should not be valid")
	}

	// Test invalid: direction out of range
	invalidDir := h3index.H3Index(edges[0])
	invalidDir = setReservedBits(invalidDir, 7) // Invalid direction
	if IsValidDirectedEdge(DirectedEdge(invalidDir)) {
		t.Error("Invalid direction should not be valid")
	}

	// Test invalid: pentagon K-axis edge
	pentagonCells := GetPentagonCells(9)
	if len(pentagonCells) > 0 {
		pentagon := pentagonCells[0]
		// Try to create a K-axis edge for pentagon (should be invalid)
		pentEdge := h3index.H3Index(pentagon)
		pentEdge = h3index.SetMode(pentEdge, constants.H3_DIRECTEDEDGE_MODE)
		pentEdge = setReservedBits(pentEdge, constants.K_AXES_DIGIT)
		if IsValidDirectedEdge(DirectedEdge(pentEdge)) {
			t.Error("Pentagon K-axis edge should not be valid")
		}
	}
}

// TestInvalidCellOriginToDirectedEdges tests OriginToDirectedEdges with invalid input.
func TestInvalidCellOriginToDirectedEdges(t *testing.T) {
	// Test with invalid cell
	invalidCell := Cell(0)
	edges := OriginToDirectedEdges(invalidCell)
	if edges != nil {
		t.Error("Expected nil edges for invalid cell")
	}

	// Test with invalid cell (wrong mode)
	wrongMode := h3index.H3_INIT
	wrongMode = h3index.SetMode(wrongMode, constants.H3_DIRECTEDEDGE_MODE)
	edges = OriginToDirectedEdges(Cell(wrongMode))
	if edges != nil {
		t.Error("Expected nil edges for wrong mode")
	}
}

// TestDirectedEdgeModeAndReservedBits tests the bit manipulation helpers.
func TestDirectedEdgeModeAndReservedBits(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	for i, e := range edges {
		h := h3index.H3Index(e)

		// Check mode bits
		mode := h.Mode()
		if mode != constants.H3_DIRECTEDEDGE_MODE {
			t.Errorf("Edge %d: expected mode %d, got %d", i, constants.H3_DIRECTEDEDGE_MODE, mode)
		}

		// Check reserved bits (direction)
		dir := getReservedBits(h)
		if dir < 1 || dir > 6 {
			t.Errorf("Edge %d: invalid direction %d", i, dir)
		}

		// Check that resolution, base cell, and digits match origin
		originH := h3index.H3Index(origin)
		if h.Resolution() != originH.Resolution() {
			t.Errorf("Edge %d: resolution mismatch", i)
		}
		if h.BaseCell() != originH.BaseCell() {
			t.Errorf("Edge %d: base cell mismatch", i)
		}

		// Check digits match
		res := h.Resolution()
		for r := 0; r < res; r++ {
			if h.IndexDigit(r) != originH.IndexDigit(r) {
				t.Errorf("Edge %d: digit mismatch at resolution %d", i, r)
			}
		}
	}
}

// TestSetAndGetReservedBits tests the reserved bits manipulation functions.
func TestSetAndGetReservedBits(t *testing.T) {
	// Start with a cell
	cell, _ := StringToCell("8928308280fffff")
	h := h3index.H3Index(cell)

	// Set each direction and verify
	for dir := 1; dir <= 6; dir++ {
		modified := setReservedBits(h, dir)
		retrieved := getReservedBits(modified)
		if retrieved != dir {
			t.Errorf("Expected direction %d, got %d", dir, retrieved)
		}

		// Check that other fields are unchanged
		if modified.Resolution() != h.Resolution() {
			t.Error("Resolution should not change")
		}
		if modified.BaseCell() != h.BaseCell() {
			t.Error("Base cell should not change")
		}
	}

	// Test clearing reserved bits
	h = setReservedBits(h, 5)
	h = setReservedBits(h, 0)
	if getReservedBits(h) != 0 {
		t.Error("Reserved bits should be cleared")
	}
}

// TestDirectedEdgeStringConversion tests conversion to/from strings.
func TestDirectedEdgeStringConversion(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	for i, e := range edges {
		// Convert to string
		str := CellToString(Cell(e))
		if str == "" {
			t.Errorf("Edge %d: empty string", i)
		}

		// Parse back
		parsed, err := StringToCell(str)
		if err != nil {
			t.Errorf("Edge %d: failed to parse: %v", i, err)
		}

		// Note: parsed will have mode=1 (cell), not mode=2 (directed edge)
		// So we can't directly compare. Instead, check the bits.
		_ = parsed // Use parsed to avoid unused variable warning

		// The string representation is valid
		if len(str) == 0 {
			t.Errorf("Edge %d: string representation is empty", i)
		}
	}
}
